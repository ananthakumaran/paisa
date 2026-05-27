package cmb

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/importer"
)

// debitCode / debitName are the stable identifiers visible to the API. Code
// MUST NOT change once shipped; the frontend pins to it.
const (
	debitCode = "cmb-debit"
	debitName = "招商银行借记卡账单"
)

// debitHeaderMarker is the unique column-header substring used by CMB debit
// CSV exports. We look for it as the most reliable content signature.
const debitHeaderMarker = "收入(+)/支出(-)"

// debitTransactionHeader is the full ordered column list as exported by Web
// Banking. We compare per-cell (after csv decoding) to handle layout drift
// gracefully — if CMB adds a column we still find the ones we need.
var debitTransactionColumns = []string{
	"记账日期", "记账时间", "收入(+)/支出(-)", "余额",
	"交易摘要", "对手户名", "对手账号", "币种",
}

// CMBDebit implements [importer.Importer] for the 招商银行借记卡 export.
// Stateless — the registry shares a single instance across concurrent
// requests; Parse only reads its argument.
type CMBDebit struct{}

// Code implements [importer.Importer]. See package doc.
func (CMBDebit) Code() string { return debitCode }

// Name implements [importer.Importer]. See package doc.
func (CMBDebit) Name() string { return debitName }

// Detect implements [importer.Importer]. Order of checks:
//  1. Filename hint — case-insensitive "cmb_debit" or literal "招行借记".
//     Cheap and survives binary / mis-encoded content.
//  2. Content fingerprint — body contains BOTH "招商银行" and the unique
//     "收入(+)/支出(-)" column header marker. Requiring both keeps random
//     CSVs from matching while still letting renamed-file exports through.
//
// We deliberately do NOT match the credit-card filename hints here — those
// belong to CMBCredit.
func (CMBDebit) Detect(filename string, content []byte) bool {
	lower := strings.ToLower(filename)
	// Don't claim files clearly intended for the credit-card importer.
	if strings.Contains(lower, "credit") || strings.Contains(filename, "信用卡") {
		return false
	}
	if strings.Contains(lower, "cmb_debit") || strings.Contains(filename, "招行借记") {
		return true
	}
	body := trimBOM(content)
	if hasCMBSignature(body) && bytes.Contains(body, []byte(debitHeaderMarker)) {
		return true
	}
	return false
}

// Parse implements [importer.Importer]. The export layout is two CSV blocks
// stacked on top of each other:
//
//  1. Account summary block — quoted columns including 账户简称, 姓名, 余额.
//  2. A blank line.
//  3. Transactions block — the 记账日期 …→ 交易回单 header followed by data
//     rows.
//
// We don't actually parse the summary block; we just scan past it until we
// find the transaction header row, then hand the rest to encoding/csv with
// FieldsPerRecord=-1 to tolerate ragged trailing summaries.
func (CMBDebit) Parse(content []byte) ([]importer.ParsedTxn, error) {
	body := trimBOM(content)
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, errors.New("cmb-debit: empty file")
	}

	// Locate the transactions header — its first column is the literal
	// 记账日期. Search for the whole "记账日期" string; doing a per-cell scan
	// after csv parsing the preamble is brittle because the preamble itself
	// is also CSV.
	headerIdx := bytes.Index(body, []byte("记账日期"))
	if headerIdx < 0 {
		return nil, errors.New("cmb-debit: transactions header '记账日期' not found — is this a CMB debit export?")
	}
	// Back up one byte if the header cell is double-quoted so encoding/csv
	// sees a well-formed quoted field rather than a bare quote mid-row.
	if headerIdx > 0 && body[headerIdx-1] == '"' {
		headerIdx--
	}
	rest := body[headerIdx:]

	r := csv.NewReader(bytes.NewReader(rest))
	// CMB exports occasionally leave the last column unquoted; tolerate
	// ragged rows and validate per-row instead.
	r.FieldsPerRecord = -1

	// First record must be the column header.
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("cmb-debit: malformed header: %w", err)
	}
	// Validate that the header looks like a CMB debit export — the column
	// positions are what we depend on, so they must match.
	if !looksLikeDebitHeader(header) {
		return nil, errors.New("cmb-debit: unexpected header row; expected CMB debit-card column layout")
	}

	var out []importer.ParsedTxn
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("cmb-debit: malformed row: %w", err)
		}
		if isBlankRow(row) {
			continue
		}
		if len(row) < 6 {
			// CMB sometimes appends a footer summary with very few columns;
			// skip rather than erroring out.
			continue
		}
		txn, ok, perr := parseDebitRow(row)
		if perr != nil {
			return nil, perr
		}
		if !ok {
			continue
		}
		out = append(out, txn)
	}

	if len(out) == 0 {
		return nil, errors.New("cmb-debit: no transactions found after header")
	}
	return out, nil
}

// looksLikeDebitHeader checks that the parsed header contains the expected
// CMB debit columns in some order. We tolerate (but do not require) extra
// columns — what we need is for the index lookup in parseDebitRow to land
// on the right cells.
func looksLikeDebitHeader(header []string) bool {
	required := []string{"记账日期", "收入(+)/支出(-)", "交易摘要"}
	have := make(map[string]bool, len(header))
	for _, h := range header {
		have[strings.TrimSpace(h)] = true
	}
	for _, r := range required {
		if !have[r] {
			return false
		}
	}
	return true
}

// parseDebitRow extracts a ParsedTxn from one transactions-block row. The
// bool return is "keep this row?" — false means a deliberate skip (e.g.
// blank). Returning an error is reserved for malformed data the user needs
// to know about.
//
// Layout (0-indexed):
//
//	0: 记账日期       1: 记账时间   2: 收入(+)/支出(-)   3: 余额
//	4: 交易摘要       5: 对手户名   6: 对手账号           7: 币种
//	8: 交易类型       9: 交易回单
func parseDebitRow(row []string) (importer.ParsedTxn, bool, error) {
	get := func(i int) string {
		if i < len(row) {
			return strings.TrimSpace(row[i])
		}
		return ""
	}
	dateStr := get(0)
	amountStr := get(2)
	note := get(4)
	payee := get(5)

	if dateStr == "" && amountStr == "" {
		return importer.ParsedTxn{}, false, nil
	}

	date, err := parseDebitDate(dateStr)
	if err != nil {
		return importer.ParsedTxn{}, false, fmt.Errorf("cmb-debit: invalid date %q: %w", dateStr, err)
	}

	rawAmount, ok, err := trimDecimal(amountStr)
	if err != nil {
		return importer.ParsedTxn{}, false, fmt.Errorf("cmb-debit: invalid amount %q: %w", amountStr, err)
	}
	if !ok {
		return importer.ParsedTxn{}, false, nil
	}

	// CMB's 收入(+)/支出(-) already carries the sign:
	//   rawAmount < 0 → 支出 (money LEAVING)   → ParsedTxn.Amount positive
	//   rawAmount > 0 → 收入 (money ENTERING)  → ParsedTxn.Amount negative
	// We deliberately preserve a zero amount as a skip; CMB never emits one.
	if rawAmount.IsZero() {
		return importer.ParsedTxn{}, false, nil
	}
	amount := rawAmount.Neg()

	suggested := suggestDebitCounterpart(payee, note, amount.IsPositive())

	txn := importer.ParsedTxn{
		Date:             date,
		Payee:            payee,
		Note:             note,
		Amount:           amount,
		Currency:         "CNY",
		SuggestedAccount: suggested,
		RawText:          strings.Join(row, ","),
	}
	return txn, true, nil
}

// parseDebitDate handles the two date formats we have observed in real
// exports: 2024/01/15 (Web Banking) and 2024-01-15 (the desktop companion
// app's CSV export-as-text mode).
func parseDebitDate(s string) (time.Time, error) {
	for _, layout := range []string{"2006/01/02", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised date format")
}

// suggestDebitCounterpart picks a counterpart account for the OTHER side of
// the transaction (the user's source-side is the debit card itself). For
// expenses (outgoing) we consult the shared merchant heuristics; for income
// we consult the income heuristics; and a credit-card repayment short-circuits
// to Liabilities:Credit:CMB so the user can wire it as an internal transfer
// in the preview UI.
func suggestDebitCounterpart(payee, note string, outgoing bool) string {
	hay := payee + " " + note
	// Credit-card repayments are the only "cross-account hint" we surface
	// from the debit side. They are unambiguous in CMB exports — either
	// 对手户名 contains "信用卡" or 交易摘要 contains "信用卡还款".
	if strings.Contains(hay, "信用卡") || (outgoing && strings.Contains(note, "还款")) {
		return "Liabilities:Credit:CMB"
	}
	if outgoing {
		return suggestExpenseAccount(payee, note)
	}
	return suggestIncomeAccount(payee, note)
}

// isBlankRow returns true if every field of the record is empty after
// trimming. Used to skip the blank line between the summary and transaction
// blocks if it survives encoding/csv's tokenisation.
func isBlankRow(row []string) bool {
	for _, f := range row {
		if strings.TrimSpace(f) != "" {
			return false
		}
	}
	return true
}
