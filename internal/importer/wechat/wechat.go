// Package wechat parses 微信支付月账单 (WeChat Pay monthly statement) CSV
// exports into the framework's [importer.ParsedTxn] shape.
//
// The export format is a UTF-8 (BOM-prefixed) CSV with a long preamble of
// metadata (account nickname, date range, totals) followed by a separator
// line "----------------------微信支付账单明细列表--------------------" and
// a column header row. Real exports include 11 fields per data row:
//
//	交易时间,交易类型,交易对方,商品,收/支,金额(元),支付方式,当前状态,交易单号,商户单号,备注
//
// Detection is primarily content-based: the very first line of every export
// is the literal "微信支付账单明细". A filename heuristic is provided as a
// fallback for the common case where a user renamed the file. Both branches
// have explicit tests so neither silently regresses.
//
// The importer registers itself via init() so any binary that links the
// package picks it up automatically. The HTTP layer in
// internal/server/import.go discovers it through the shared registry — no
// new wiring is required at the route layer.
package wechat

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/shopspring/decimal"
)

// code/name are the stable identifiers visible to the API. Code MUST NOT
// change once shipped; the frontend pins to it. Name is the human label —
// translation is the UI's job.
const (
	code = "wechat"
	name = "微信支付月账单"
)

// magicHeader is the literal first line of every official WeChat Pay export.
// Matching against this is the most reliable detection signal we have; it has
// remained stable across years of export-format minor revisions.
const magicHeader = "微信支付账单明细"

// listSeparator marks the boundary between the preamble (account info,
// totals) and the actual transaction table. Real exports use 22 leading
// dashes; we match a looser prefix to survive future tweaks.
const listSeparator = "微信支付账单明细列表"

// utf8BOM is the byte-order-mark WeChat prepends to every export.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// expectedFieldCount is what every well-formed transaction row contains.
// The preamble lines have far fewer commas (most are 0), so any row with
// this many fields is a reasonable transaction-row candidate.
const expectedFieldCount = 11

// WeChat is the importer implementation. It is stateless — the same instance
// is shared across concurrent HTTP requests by the registry, which is fine
// because Parse only reads its argument.
type WeChat struct{}

// Code implements [importer.Importer]. See package doc.
func (WeChat) Code() string { return code }

// Name implements [importer.Importer]. See package doc.
func (WeChat) Name() string { return name }

// Detect implements [importer.Importer]. Order of checks:
//  1. Filename hint — case-insensitive substring match on "wechat" or "微信".
//     Cheap and survives content we cannot read (binary, mis-encoded).
//  2. Content header — strip a UTF-8 BOM if present, look at the first line
//     and accept a prefix match of [magicHeader]. We use a prefix (rather
//     than equality) because WeChat sometimes pads the line with trailing
//     commas to satisfy spreadsheet apps.
//
// False positives here are tolerable (the picker UI lets the user override);
// false negatives hide the importer entirely, so we err generous.
func (WeChat) Detect(filename string, content []byte) bool {
	lower := strings.ToLower(filename)
	if strings.Contains(lower, "wechat") || strings.Contains(filename, "微信") {
		return true
	}

	body := bytes.TrimPrefix(content, utf8BOM)
	firstLine := body
	if idx := bytes.IndexByte(body, '\n'); idx >= 0 {
		firstLine = body[:idx]
	}
	return strings.HasPrefix(strings.TrimSpace(string(firstLine)), magicHeader)
}

// Parse implements [importer.Importer]. The algorithm:
//  1. Strip BOM.
//  2. Locate the separator line (listSeparator) — everything above it is
//     preamble noise.
//  3. The next non-empty line is the column header. We tolerate (but do
//     not require) the exact column names — what we rely on is the
//     positional layout the format has used since at least 2020.
//  4. Iterate remaining rows. encoding/csv handles quoted fields and
//     embedded commas correctly.
//  5. For each row: skip failures ("当前状态" contains 失败), skip neutral
//     transfers (收/支 == "/" or ""), strip the "¥" prefix from the amount,
//     and apply sign per the [importer.ParsedTxn] convention (positive =
//     leaves source).
//  6. Suggest a counterpart account from a small set of merchant /
//     transaction-type heuristics — same vocabulary as the alipay importer
//     so users see consistent suggestions across data sources.
func (WeChat) Parse(content []byte) ([]importer.ParsedTxn, error) {
	body := bytes.TrimPrefix(content, utf8BOM)
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, errors.New("wechat: empty file")
	}

	// Find the separator row. Everything from the line AFTER it onwards is
	// CSV data (header + rows). We deliberately use a byte-level scan rather
	// than parsing the whole preamble with encoding/csv: some preamble lines
	// have unbalanced quotes or stray commas inside the metadata, which the
	// strict CSV reader would reject.
	sepIdx := bytes.Index(body, []byte(listSeparator))
	if sepIdx < 0 {
		return nil, errors.New("wechat: separator '微信支付账单明细列表' not found — is this really a WeChat Pay export?")
	}
	// Advance to the end of the separator line.
	rest := body[sepIdx:]
	if nl := bytes.IndexByte(rest, '\n'); nl >= 0 {
		rest = rest[nl+1:]
	} else {
		rest = nil
	}
	if len(bytes.TrimSpace(rest)) == 0 {
		return nil, errors.New("wechat: no transaction rows after separator")
	}

	r := csv.NewReader(bytes.NewReader(rest))
	r.FieldsPerRecord = -1 // tolerate ragged rows; we validate per-row below

	// First non-empty record is the column header. We just discard it —
	// the positional layout is what matters.
	for {
		header, err := r.Read()
		if err == io.EOF {
			return nil, errors.New("wechat: missing column header row")
		}
		if err != nil {
			return nil, fmt.Errorf("wechat: malformed header: %w", err)
		}
		if !isBlankRow(header) {
			break
		}
	}

	var out []importer.ParsedTxn
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("wechat: malformed row: %w", err)
		}
		if isBlankRow(row) {
			continue
		}
		if len(row) < expectedFieldCount {
			// Trailing summary lines ("以上是您的微信支付账单明细", etc.) tend to
			// have a handful of fields. Skip them silently rather than
			// erroring out — a stricter parse would just make the format
			// brittle when WeChat tweaks its footer.
			continue
		}

		txn, ok, perr := parseRow(row)
		if perr != nil {
			return nil, perr
		}
		if !ok {
			continue
		}
		out = append(out, txn)
	}

	return out, nil
}

// parseRow turns one CSV record into a ParsedTxn. The bool return is the
// "keep this row?" signal: false means the row was a deliberate skip
// (failed payment, neutral transfer, etc.) — NOT an error. Returning an
// error is reserved for genuinely malformed data (bad date, unparseable
// amount) which the user needs to know about.
func parseRow(row []string) (importer.ParsedTxn, bool, error) {
	// Positional fields per the documented layout:
	//   0: 交易时间       1: 交易类型     2: 交易对方   3: 商品
	//   4: 收/支         5: 金额(元)     6: 支付方式   7: 当前状态
	//   8: 交易单号       9: 商户单号    10: 备注
	get := func(i int) string {
		if i < len(row) {
			return strings.TrimSpace(row[i])
		}
		return ""
	}
	tradeTime := get(0)
	tradeType := get(1)
	payee := get(2)
	goods := get(3)
	flow := get(4)
	amountStr := get(5)
	status := get(7)

	// Drop failed payments outright — they never settled and including them
	// in the journal would inflate expenses.
	if strings.Contains(status, "失败") {
		return importer.ParsedTxn{}, false, nil
	}

	// Skip neutral / internal-transfer rows. The user can record those by
	// hand or via a future "transfers" importer once we ship one.
	if flow == "" || flow == "/" {
		return importer.ParsedTxn{}, false, nil
	}

	date, err := time.Parse("2006-01-02 15:04:05", tradeTime)
	if err != nil {
		return importer.ParsedTxn{}, false, fmt.Errorf("wechat: invalid date %q: %w", tradeTime, err)
	}

	// Strip leading "¥" — present in modern exports but absent from older
	// or hand-edited ones. The TrimSpace already ran inside `get`.
	amountClean := strings.TrimPrefix(amountStr, "¥")
	amountClean = strings.TrimSpace(amountClean)
	amount, err := decimal.NewFromString(amountClean)
	if err != nil {
		return importer.ParsedTxn{}, false, fmt.Errorf("wechat: invalid amount %q: %w", amountStr, err)
	}

	// Sign convention from importer.ParsedTxn:
	//   positive = money LEAVING source account (支出)
	//   negative = money ENTERING source account (收入)
	switch flow {
	case "支出":
		// already positive
	case "收入":
		amount = amount.Neg()
	default:
		// Anything else we did not explicitly handle (defensive): treat
		// as neutral and skip rather than guess wrong.
		return importer.ParsedTxn{}, false, nil
	}

	txn := importer.ParsedTxn{
		Date:             date,
		Payee:            payee,
		Note:             goods,
		Amount:           amount,
		Currency:         "CNY",
		SuggestedAccount: suggestAccount(payee, goods, tradeType, flow),
		RawText:          strings.Join(row, ","),
	}
	// "/" appears as a placeholder when a field is unused. Don't leak it.
	if txn.Payee == "/" {
		txn.Payee = ""
	}
	if txn.Note == "/" {
		txn.Note = ""
	}
	return txn, true, nil
}

// suggestAccount returns a best-effort counterpart account based on cheap
// substring heuristics. The same vocabulary will be used by the M3-B alipay
// importer once it lands, so users see consistent suggestions whichever
// statement they import. The TF-IDF predictor in M3-F may override these
// at commit time — Detect's job here is just to give a sensible default.
func suggestAccount(payee, goods, tradeType, flow string) string {
	// Red packet rules first — these are unambiguous and should not be
	// shadowed by the merchant table even if a future "星巴克红包" promo
	// shows up.
	if strings.Contains(goods, "微信红包") || strings.Contains(tradeType, "红包") {
		if flow == "支出" {
			return "Expenses:Social:RedPacket"
		}
		if flow == "收入" {
			return "Income:RedPacket"
		}
	}

	// Pure transfers ("转账"): we can't know which account on the other
	// end the user actually wants — leave them with a generic Checking
	// hint so the UI nudges them to pick the right counterpart.
	if tradeType == "转账" {
		return "Assets:Checking"
	}

	// Merchant heuristics. Substring match on payee is enough; the data
	// is dense and the vocabulary is shared with the alipay importer.
	type rule struct {
		needles []string
		account string
	}
	merchantRules := []rule{
		{[]string{"星巴克", "瑞幸"}, "Expenses:Dining"},
		{[]string{"滴滴", "高德"}, "Expenses:Transport:Taxi"},
		{[]string{"京东", "拼多多", "淘宝"}, "Expenses:Shopping"},
	}
	for _, rule := range merchantRules {
		for _, n := range rule.needles {
			if strings.Contains(payee, n) {
				return rule.account
			}
		}
	}

	// Defaults — direction tells us which side of the ledger to fall back
	// to. The user will refine the bucket in the preview UI.
	if flow == "收入" {
		return "Income:Unknown"
	}
	return "Expenses:Unknown"
}

// isBlankRow returns true when every field of the record is empty (after
// trimming). Blank rows appear between the preamble and the data, and we
// don't want them to disturb header / data alignment.
func isBlankRow(row []string) bool {
	for _, f := range row {
		if strings.TrimSpace(f) != "" {
			return false
		}
	}
	return true
}

// init registers the importer with the shared registry. The framework's
// contract is that subpackages self-register at import time; importing this
// package (directly or via a blank import elsewhere) wires it up.
func init() {
	importer.Register(WeChat{})
}
