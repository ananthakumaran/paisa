// Package ibkr parses Interactive Brokers Flex Query / Activity Statement
// CSV files. IBKR statements are unusual: they pack many independent
// "sections" (Trades, Dividends, Fees, …) into a single CSV by prefixing
// every row with a section name and a "Header"/"Data" discriminator. This
// importer dispatches by section so each transaction category lives in its
// own file (trades.go, dividends.go, fees.go).
//
// Sections that are pure metadata — Account Information, Statement period,
// Cash Report, Open Positions — are intentionally ignored: they describe
// state, not movements, and importing them as ledger entries would
// double-count balances. The HTTP layer's preview UI lets users skip /
// edit anything else.
//
// IBKR statements are emitted with DOS line endings (\r\n) and the parser
// must handle them transparently — encoding/csv already strips the trailing
// \r from each field, but bytes.Contains-style detection has to allow for it.
package ibkr

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"

	"github.com/ananthakumaran/paisa/internal/importer"
)

// Code is exported so other packages (server registration, tests) reference
// a single constant rather than the literal string.
const Code = "ibkr"

// IBKR implements importer.Importer.
type IBKR struct{}

func (IBKR) Code() string { return Code }
func (IBKR) Name() string { return "IBKR Activity Statement" }

// Detect is content-based only. We deliberately do NOT look at the filename:
// IBKR users routinely rename their exports, and the file body contains
// strong enough signals (the ClientAccountID preamble or the Account
// Information section header) that filename heuristics would only add false
// positives.
func (IBKR) Detect(filename string, content []byte) bool {
	// Two signals — either is sufficient. The preamble lives at byte 0 of
	// every IBKR Flex Query CSV; the Account Information section header
	// appears in the first ~10 lines and survives any custom Flex Query
	// configuration that disables the preamble.
	return bytes.Contains(content, []byte(`ClientAccountID`)) ||
		bytes.Contains(content, []byte(`"Account Information","Header"`))
}

// Parse splits the file into sections, then delegates each known section to
// its dedicated parser. Sections we don't understand are silently skipped —
// IBKR adds new ones over time (e.g. corporate actions, transfers) and we
// would rather ignore the unknown than crash the whole import.
func (IBKR) Parse(content []byte) ([]importer.ParsedTxn, error) {
	sections := splitBySection(content)
	var txns []importer.ParsedTxn

	if rows := sections["Trades"]; len(rows) > 0 {
		txns = append(txns, parseTrades(rows)...)
	}
	if rows := sections["Dividends"]; len(rows) > 0 {
		txns = append(txns, parseDividends(rows)...)
	}
	if rows := sections["Withholding Tax"]; len(rows) > 0 {
		txns = append(txns, parseWithholdingTax(rows)...)
	}
	if rows := sections["Fees"]; len(rows) > 0 {
		txns = append(txns, parseFees(rows)...)
	}
	if rows := sections["Interest"]; len(rows) > 0 {
		txns = append(txns, parseInterest(rows)...)
	}

	return txns, nil
}

// splitBySection reads the CSV row-by-row and groups rows by their first
// column (the section name). Rows whose second column isn't "Header" or
// "Data" are skipped — IBKR uses other discriminators like "SubTotal" and
// "Total" for summary rows that must NOT become transactions.
//
// The returned map is keyed by section name; each value is the slice of
// rows (header first, then data rows) for that section. Returning rows
// rather than parsed structs keeps section parsers free to interpret their
// own columns (Trades has 16 fields, Dividends has 5).
func splitBySection(content []byte) map[string][][]string {
	if len(content) == 0 {
		return nil
	}
	r := csv.NewReader(bytes.NewReader(content))
	// IBKR rows have wildly variable field counts even within the file
	// (the preamble line has 4 fields; Trades rows have 16). Disable the
	// per-row field check.
	r.FieldsPerRecord = -1
	// IBKR very rarely emits bare CR inside quoted fields, but
	// LazyQuotes guards us against any stray quote-escaping oddities.
	r.LazyQuotes = true

	out := make(map[string][][]string)
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Skip individual malformed rows rather than abort the
			// whole import — a single mangled row should not block
			// all the other (correct) sections.
			continue
		}
		if len(row) < 2 {
			continue
		}
		discriminator := row[1]
		if discriminator != "Header" && discriminator != "Data" {
			continue
		}
		section := row[0]
		out[section] = append(out[section], row)
	}
	return out
}

// columnIndex returns a map from column name (from the Header row) to its
// position. Used by every section parser so they can address columns by
// name rather than by hard-coded index — IBKR has been known to reorder
// columns between Flex Query versions.
func columnIndex(headerRow []string) map[string]int {
	idx := make(map[string]int, len(headerRow))
	for i, name := range headerRow {
		idx[strings.TrimSpace(name)] = i
	}
	return idx
}

// at returns the trimmed value of row[idx[col]], or "" if the column is
// missing. Centralising the lookup keeps each section parser readable and
// makes the "missing column = empty string" policy explicit.
func at(row []string, idx map[string]int, col string) string {
	i, ok := idx[col]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func init() {
	importer.Register(IBKR{})
}
