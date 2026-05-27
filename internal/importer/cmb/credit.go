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

// creditCode / creditName are the stable identifiers visible to the API.
const (
	creditCode = "cmb-credit"
	creditName = "招商银行信用卡账单"
)

// creditHeaderMarker is the unique column-header substring used by CMB
// credit-card exports — present in both the Web-Banking CSV and the
// "信用卡掌上生活" app's XLSX/CSV variants.
const creditHeaderMarker = "卡号末四位"

// CMBCredit implements [importer.Importer] for 招商银行信用卡 statements.
//
// On format:
//
//	The official desktop / mobile export is XLSX (with multiple sheets —
//	summary + transactions). Many users export the transactions sheet as
//	CSV before importing into a third-party tool. We support BOTH paths in
//	[Detect] — the filename hint is the primary signal for binary XLSX
//	(we cannot meaningfully read XLSX without an extra dependency, so the
//	parser handles CSV here; XLSX support can be added later behind the
//	same Detect path by inspecting application/vnd.openxmlformats…).
//
// Parse: only the CSV branch is implemented. If we ever ship binary XLSX
// parsing it would be a transparent extension here; Detect already accepts
// .xlsx filenames, so the user-facing UX is forward-compatible.
type CMBCredit struct{}

// Code implements [importer.Importer]. See package doc.
func (CMBCredit) Code() string { return creditCode }

// Name implements [importer.Importer]. See package doc.
func (CMBCredit) Name() string { return creditName }

// Detect implements [importer.Importer]. We try, in order:
//  1. Filename hint — "cmb_credit", "信用卡", or "招行信用卡" anywhere in the
//     filename (case-insensitive for the English token).
//  2. Content fingerprint — the CSV variant contains "卡号末四位" in its
//     header. This is unique enough to use as a content signature.
//
// We never light up on raw XLSX magic alone; that would catch every Excel
// upload regardless of bank. Filename remains the primary disambiguator for
// binary payloads.
func (CMBCredit) Detect(filename string, content []byte) bool {
	lower := strings.ToLower(filename)
	if strings.Contains(lower, "cmb_credit") ||
		strings.Contains(filename, "信用卡") ||
		strings.Contains(filename, "招行信用卡") {
		return true
	}
	body := trimBOM(content)
	// XLSX magic + "信用卡" filename would have matched above; bare XLSX
	// without the filename hint is intentionally rejected — see test.
	if bytes.HasPrefix(body, xlsxMagic) {
		return false
	}
	if bytes.Contains(body, []byte(creditHeaderMarker)) {
		return true
	}
	return false
}

// Parse implements [importer.Importer]. Reads the CSV variant of the
// credit-card monthly statement and emits one [importer.ParsedTxn] per
// transaction row, including the original currency (RMB rows stay CNY,
// foreign rows keep their original currency code). M1-F's归一 handles the
// FX conversion downstream.
//
// Row sign convention:
//
//	人民币金额 < 0  → spending  → ParsedTxn.Amount POSITIVE
//	人民币金额 > 0  → refund/repayment → ParsedTxn.Amount NEGATIVE
//
// We DO keep repayment rows (匹配 "还款") because the user may want them in
// the ledger as a transfer from the debit account. The suggested counterpart
// points at Assets:Saving:CMB to make that wiring explicit in the preview UI.
func (CMBCredit) Parse(content []byte) ([]importer.ParsedTxn, error) {
	body := trimBOM(content)
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, errors.New("cmb-credit: empty file")
	}
	if bytes.HasPrefix(body, xlsxMagic) {
		// Forward-compat: when we add XLSX support, branch here. For now we
		// surface a clear error so the user knows to export-as-CSV first.
		return nil, errors.New("cmb-credit: binary XLSX is not supported yet — export the transactions sheet as CSV first")
	}

	// Locate the transactions header — the unique 卡号末四位 token only
	// appears in the column header.
	headerIdx := bytes.Index(body, []byte("交易日期"))
	if headerIdx < 0 {
		return nil, errors.New("cmb-credit: transactions header '交易日期' not found — is this a CMB credit-card export?")
	}
	// Back up one byte if the header cell is double-quoted so encoding/csv
	// sees a well-formed quoted field rather than a bare quote mid-row.
	if headerIdx > 0 && body[headerIdx-1] == '"' {
		headerIdx--
	}
	rest := body[headerIdx:]

	r := csv.NewReader(bytes.NewReader(rest))
	r.FieldsPerRecord = -1

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("cmb-credit: malformed header: %w", err)
	}
	if !looksLikeCreditHeader(header) {
		return nil, errors.New("cmb-credit: unexpected header row; expected CMB credit-card column layout")
	}
	colIdx := indexCreditColumns(header)

	var out []importer.ParsedTxn
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("cmb-credit: malformed row: %w", err)
		}
		if isBlankRow(row) {
			continue
		}
		if len(row) < 5 {
			continue
		}
		txn, ok, perr := parseCreditRow(row, colIdx)
		if perr != nil {
			return nil, perr
		}
		if !ok {
			continue
		}
		out = append(out, txn)
	}

	if len(out) == 0 {
		return nil, errors.New("cmb-credit: no transactions found after header")
	}
	return out, nil
}

// creditCols holds the resolved zero-based index of every column we care
// about. -1 means "missing"; the parser falls back to sensible defaults for
// optional fields (e.g. 美元金额 may be absent in pure-RMB exports).
type creditCols struct {
	tradeDate int
	postDate  int
	note      int
	last4     int
	rmb       int
	usd       int
	currency  int
}

// looksLikeCreditHeader verifies the parsed header contains every column we
// rely on for positional parsing.
func looksLikeCreditHeader(header []string) bool {
	required := []string{"交易日期", "交易摘要", "人民币金额"}
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

// indexCreditColumns resolves human-readable column names to their actual
// positions in the header. Tolerates layout drift — if CMB inserts a new
// column we don't crash, we just don't read the unknown column.
func indexCreditColumns(header []string) creditCols {
	c := creditCols{
		tradeDate: -1, postDate: -1, note: -1, last4: -1,
		rmb: -1, usd: -1, currency: -1,
	}
	for i, h := range header {
		switch strings.TrimSpace(h) {
		case "交易日期":
			c.tradeDate = i
		case "入账日期":
			c.postDate = i
		case "交易摘要":
			c.note = i
		case "卡号末四位":
			c.last4 = i
		case "人民币金额":
			c.rmb = i
		case "美元金额":
			c.usd = i
		case "交易币种":
			c.currency = i
		}
	}
	return c
}

// parseCreditRow turns one CSV record into a ParsedTxn. See file doc for
// the sign convention. We keep repayment rows but flag them with a
// counterpart hint pointing at Assets:Saving:CMB.
func parseCreditRow(row []string, c creditCols) (importer.ParsedTxn, bool, error) {
	get := func(i int) string {
		if i >= 0 && i < len(row) {
			return strings.TrimSpace(row[i])
		}
		return ""
	}
	dateStr := get(c.tradeDate)
	note := get(c.note)
	rmbStr := get(c.rmb)
	usdStr := get(c.usd)
	currency := get(c.currency)

	if dateStr == "" {
		return importer.ParsedTxn{}, false, nil
	}

	date, err := parseDebitDate(dateStr) // same set of layouts; reuse parser
	if err != nil {
		return importer.ParsedTxn{}, false, fmt.Errorf("cmb-credit: invalid date %q: %w", dateStr, err)
	}

	// Pick the amount column: USD column wins iff currency is non-CNY AND
	// the USD field is non-empty. Everything else uses the RMB column —
	// CMB always populates 人民币金额 even for foreign-currency rows (as
	// the converted CNY value).
	useUSD := strings.EqualFold(currency, "USD") && usdStr != ""
	src := rmbStr
	if useUSD {
		src = usdStr
	}
	rawAmount, ok, err := trimDecimal(src)
	if err != nil {
		return importer.ParsedTxn{}, false, fmt.Errorf("cmb-credit: invalid amount %q: %w", src, err)
	}
	if !ok || rawAmount.IsZero() {
		return importer.ParsedTxn{}, false, nil
	}

	// Sign convention: source has 人民币金额 negative for spending; flip.
	amount := rawAmount.Neg()

	outCurrency := "CNY"
	if useUSD {
		outCurrency = "USD"
	} else if currency != "" && !strings.EqualFold(currency, "RMB") && !strings.EqualFold(currency, "CNY") {
		// Honour an explicit foreign currency token even if the export only
		// gave us 人民币金额 (some older formats omit the foreign-currency
		// column entirely). We keep the CNY amount but flag the currency
		// for M1-F to convert.
		outCurrency = strings.ToUpper(currency)
	}

	suggested := suggestCreditCounterpart(note, amount.IsPositive())

	txn := importer.ParsedTxn{
		Date:             date,
		Payee:            note, // credit-card statements have no separate payee field
		Note:             note,
		Amount:           amount,
		Currency:         outCurrency,
		SuggestedAccount: suggested,
		RawText:          strings.Join(row, ","),
	}
	return txn, true, nil
}

// suggestCreditCounterpart picks a counterpart hint for a credit-card row.
// Repayments (incoming, contains "还款") get Assets:Saving:CMB so the user
// can wire them as internal transfers in the UI; everything else is a normal
// expense / refund.
func suggestCreditCounterpart(note string, outgoing bool) string {
	if strings.Contains(note, "还款") {
		return "Assets:Saving:CMB"
	}
	if outgoing {
		return suggestExpenseAccount(note, note)
	}
	// Incoming on a credit card is normally a refund.
	if strings.Contains(note, "退款") {
		return "Income:Refund"
	}
	return suggestIncomeAccount(note, note)
}

// We re-export ParseTimeForTest? No — keep the package surface tiny. The
// debit's parseDebitDate is package-internal and reused via direct call.
// Documented here for the next reader:
//
//	Why share parseDebitDate? Both statements use the same date layouts
//	(2006/01/02 from Web Banking, 2006-01-02 from third-party converters)
//	and we want a single source of truth. The function lives in debit.go to
//	keep credit.go focused on credit-specific concerns.
var _ = time.Time{} // keep "time" import alive even if Go ever inlines parseDebitDate
