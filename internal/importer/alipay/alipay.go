// Package alipay parses Alipay (支付宝) monthly transaction CSV exports into
// importer.ParsedTxn values for the M3 import pipeline. The export format is
// GBK-encoded by default, with a multi-line header block and a single CSV
// body table; some recent exports are already UTF-8 (with or without BOM),
// so this package autodetects the encoding before parsing.
//
// The package self-registers with the importer registry at init() so that
// the HTTP layer in internal/server/import.go discovers it automatically.
// See internal/importer/importer.go for the contract.
package alipay

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/ananthakumaran/paisa/internal/prediction"
	"github.com/shopspring/decimal"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Alipay is the stateless importer for 支付宝 monthly statement CSVs.
type Alipay struct{}

// Code is the stable machine identifier persisted in user preferences. Must
// not change once shipped.
func (Alipay) Code() string { return "alipay" }

// Name is the human-readable label surfaced in the import UI.
func (Alipay) Name() string { return "支付宝月账单" }

// headerMarker is the magic substring present in every alipay export's
// preamble (it appears twice, framed in dashes). We use it as the primary
// content-based detection signal.
const headerMarker = "交易记录明细列表"

// filenameHint is the leading slug typical alipay exports use:
// `alipay_record_YYYYMM.csv`. It survives renames in many cases.
const filenameHint = "alipay_record"

// Detect returns true if the file is an alipay statement. We check the
// filename hint first (cheap, no decoding needed) and then fall back to a
// content-based check that handles either encoding.
func (Alipay) Detect(filename string, content []byte) bool {
	if strings.Contains(strings.ToLower(filename), filenameHint) {
		return true
	}
	text, err := decodeToUTF8(content)
	if err != nil {
		return false
	}
	return strings.Contains(text, headerMarker)
}

// Parse converts the file content into ParsedTxn values. It:
//  1. Decodes GBK → UTF-8 (or strips BOM if already UTF-8).
//  2. Locates the column-header row by looking for "交易号" + "交易创建时间".
//  3. Parses the remaining body with encoding/csv.
//  4. Filters out rows whose 交易状态 is not 交易成功 and rows with no
//     direction marker (内部 balance moves).
//  5. Maps each row to a ParsedTxn, applying the sign convention (positive
//     amount = money leaving the source account).
func (Alipay) Parse(content []byte) ([]importer.ParsedTxn, error) {
	if len(bytes.TrimSpace(content)) == 0 {
		return nil, errors.New("alipay: empty file")
	}
	text, err := decodeToUTF8(content)
	if err != nil {
		return nil, fmt.Errorf("alipay: decode: %w", err)
	}

	headerIdx, columns, err := findHeaderRow(text)
	if err != nil {
		return nil, err
	}

	body := text[headerIdx:]
	// Each alipay export ends with a trailer block (totals + footer dashes).
	// We trim everything after the first non-data line we encounter inside
	// the reader loop, but we also drop any obvious leading whitespace.
	r := csv.NewReader(strings.NewReader(body))
	r.FieldsPerRecord = -1 // alipay sometimes ships a trailing empty column
	r.LazyQuotes = true
	// Skip the header row itself; we already parsed it.
	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("alipay: read header: %w", err)
	}

	idx := indexColumns(columns)
	if err := idx.validate(); err != nil {
		return nil, err
	}

	var out []importer.ParsedTxn
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Footer lines like "------------交易记录明细列表------------"
			// trip the CSV parser. Treat any parse error past the table as
			// end-of-data rather than a fatal error.
			break
		}
		if !looksLikeDataRow(row, idx) {
			// Footer / blank / summary line — stop processing.
			break
		}
		txn, ok, err := rowToTxn(row, idx)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		out = append(out, txn)
	}
	return out, nil
}

// columnIndex maps logical fields to their CSV column position. Alipay has
// shipped multiple slightly different header orderings over the years; we
// resolve by header name rather than by fixed position to stay robust.
type columnIndex struct {
	date    int
	typ     int
	payee   int
	note    int
	amount  int
	dir     int
	status  int
	comment int
}

func (c columnIndex) validate() error {
	missing := []string{}
	if c.date < 0 {
		missing = append(missing, "交易创建时间")
	}
	if c.payee < 0 {
		missing = append(missing, "交易对方")
	}
	if c.amount < 0 {
		missing = append(missing, "金额")
	}
	if c.dir < 0 {
		missing = append(missing, "收/支")
	}
	if c.status < 0 {
		missing = append(missing, "交易状态")
	}
	if len(missing) > 0 {
		return fmt.Errorf("alipay: missing required columns: %s", strings.Join(missing, ", "))
	}
	return nil
}

// indexColumns locates each logical field by matching its header label.
// Returns -1 for any field we can't find; the caller (validate) decides
// whether that's fatal.
func indexColumns(headers []string) columnIndex {
	idx := columnIndex{date: -1, typ: -1, payee: -1, note: -1, amount: -1, dir: -1, status: -1, comment: -1}
	for i, raw := range headers {
		h := strings.TrimSpace(raw)
		switch {
		case h == "交易创建时间":
			idx.date = i
		case h == "类型":
			idx.typ = i
		case h == "交易对方":
			idx.payee = i
		case h == "商品名称":
			idx.note = i
		case strings.HasPrefix(h, "金额"): // 金额（元） / 金额(元)
			idx.amount = i
		case h == "收/支":
			idx.dir = i
		case h == "交易状态":
			idx.status = i
		case h == "备注":
			idx.comment = i
		}
	}
	return idx
}

// findHeaderRow walks the file line-by-line until it sees a row whose fields
// include the canonical column names. Returns the byte offset of the header
// row in `text` and the parsed header fields.
func findHeaderRow(text string) (int, []string, error) {
	offset := 0
	for {
		nl := strings.IndexByte(text[offset:], '\n')
		var line string
		if nl < 0 {
			line = text[offset:]
		} else {
			line = text[offset : offset+nl]
		}
		trimmed := strings.TrimSpace(line)
		// Header rows in alipay exports always contain both 交易号 and 交易创建时间.
		if strings.Contains(trimmed, "交易号") && strings.Contains(trimmed, "交易创建时间") {
			cols := splitCSVLine(trimmed)
			return offset, cols, nil
		}
		if nl < 0 {
			return 0, nil, errors.New("alipay: column header row not found")
		}
		offset += nl + 1
	}
}

// splitCSVLine parses one CSV line into fields. We use the stdlib CSV
// reader rather than strings.Split so quoted commas inside fields don't
// break us — though alipay headers don't quote, body rows occasionally do.
func splitCSVLine(line string) []string {
	r := csv.NewReader(strings.NewReader(line))
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	row, err := r.Read()
	if err != nil {
		// Fall back to a naive split so we always return something useful.
		return strings.Split(line, ",")
	}
	return row
}

// looksLikeDataRow checks the row has enough columns and looks like a
// transaction (date column parses, status column is non-empty). Footer
// rows ("交易记录明细列表" etc.) fail this check and stop the loop.
func looksLikeDataRow(row []string, idx columnIndex) bool {
	if len(row) <= idx.amount || len(row) <= idx.dir || len(row) <= idx.status {
		return false
	}
	date := strings.TrimSpace(row[idx.date])
	if date == "" {
		return false
	}
	// First 4 chars should be a 4-digit year. Cheaper than time.Parse.
	if len(date) < 10 {
		return false
	}
	for i := 0; i < 4; i++ {
		if date[i] < '0' || date[i] > '9' {
			return false
		}
	}
	return true
}

// rowToTxn maps one CSV record to a ParsedTxn. Returns ok=false when the
// row should be silently dropped (failed status, zero-amount balance move).
func rowToTxn(row []string, idx columnIndex) (importer.ParsedTxn, bool, error) {
	get := func(i int) string {
		if i < 0 || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	status := get(idx.status)
	if status != "交易成功" {
		return importer.ParsedTxn{}, false, nil
	}

	dir := get(idx.dir)
	if dir == "" {
		// 内部 balance shuffle (e.g. 余额宝 auto-invest) — no real economic
		// event from the user's POV; drop it.
		return importer.ParsedTxn{}, false, nil
	}

	amountStr := get(idx.amount)
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return importer.ParsedTxn{}, false, fmt.Errorf("alipay: invalid amount %q: %w", amountStr, err)
	}
	if amount.IsZero() {
		return importer.ParsedTxn{}, false, nil
	}

	dateStr := get(idx.date)
	date, err := parseDate(dateStr)
	if err != nil {
		return importer.ParsedTxn{}, false, fmt.Errorf("alipay: invalid date %q: %w", dateStr, err)
	}

	payee := get(idx.payee)
	note := get(idx.note)
	comment := get(idx.comment)
	typ := get(idx.typ)
	if comment != "" {
		note = strings.TrimSpace(note + " " + comment)
	}

	// Sign convention: positive = money LEAVING source account.
	//   支出 → +amount, 收入 → -amount.
	signed := amount
	switch dir {
	case "支出":
		signed = amount.Abs()
	case "收入":
		signed = amount.Abs().Neg()
	default:
		// Unknown direction marker — be conservative and drop the row so
		// the user is not silently shown a misclassified entry.
		return importer.ParsedTxn{}, false, nil
	}

	suggested := suggestAccount(payee, note, typ, signed)

	return importer.ParsedTxn{
		Date:             date,
		Payee:            payee,
		Note:             note,
		Amount:           signed,
		Currency:         "CNY",
		SuggestedAccount: suggested,
		RawText:          strings.Join(row, ","),
	}, true, nil
}

// parseDate handles the two date layouts alipay actually uses:
// "2006-01-02 15:04:05" (the common one) and "2006-01-02" (rare, in older
// year-end summary rows).
func parseDate(s string) (time.Time, error) {
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02"}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format: %s", s)
}

// suggestAccount picks a counterpart account for the row.
//
// As of M3-F (#24), the merchant-keyword mapping that used to live inline
// here has moved to the shared seed dictionary in
// internal/prediction/seed_zh_cn.go so the wechat importer (and any future
// CNY-denominated importer) gets the same suggestions for free. This
// function now only handles the cases that depend on alipay-specific
// signals the seed dictionary cannot see — namely the income / expense
// direction (encoded in the signed amount) which decides whether the
// fallback bucket is Income:Unknown vs Expenses:Unknown.
//
// The seed dictionary also contains a small number of income-side hints
// (e.g. "工资" → Income:Salary, "结息" → Income:BankInterest); we feed
// payee + note + typ into MatchSeed so those still fire when the row is
// negative. The previous behaviour of "工资" → Income:Salary therefore
// continues to hold through the seed dictionary rather than the inline
// switch.
//
// nil db is intentional: Parse is stateless by contract (see
// importer.go). Layer-1 (learned mappings) is applied by the server's
// /api/import/parse handler — see internal/server/import.go — after this
// function returns.
func suggestAccount(payee, note, typ string, signed decimal.Decimal) string {
	fallback := "Expenses:Unknown"
	if signed.IsNegative() {
		fallback = "Income:Unknown"
	}
	// The seed dictionary's MatchSeed inspects payee then note. The alipay
	// 交易分类 (typ) is also useful context — e.g. "转账" / "退款" rows
	// often carry the discriminator there. We splice typ into the note
	// channel so the matcher gets all three fields without expanding its
	// public signature.
	noteHay := note
	if typ != "" {
		if noteHay != "" {
			noteHay += " "
		}
		noteHay += typ
	}
	return prediction.SuggestForPayee(nil, payee, noteHay, fallback)
}

// decodeToUTF8 returns `content` as a UTF-8 string, autodetecting GBK vs
// UTF-8. We try UTF-8 first (cheap and the future-proof case), then fall
// back to GBK. A leading UTF-8 BOM is stripped either way.
func decodeToUTF8(content []byte) (string, error) {
	content = bytes.TrimPrefix(content, []byte{0xEF, 0xBB, 0xBF})
	if utf8.Valid(content) {
		return string(content), nil
	}
	// GB18030 is a superset of GBK and decodes both, so we use it as the
	// fallback for the broadest compatibility with real exports.
	decoded, _, err := transform.Bytes(simplifiedchinese.GB18030.NewDecoder(), content)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func init() {
	importer.Register(Alipay{})
}
