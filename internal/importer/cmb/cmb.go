// Package cmb parses 招商银行 (China Merchants Bank) personal-account
// statements into the framework's [importer.ParsedTxn] shape. Two flavours
// share this package because they share a brand (CMB) and a default account
// vocabulary, but their source formats are different enough that we ship
// them as separate [importer.Importer] implementations:
//
//   - CMBDebit  (code "cmb-debit")  — 借记卡历史交易明细 CSV export
//   - CMBCredit (code "cmb-credit") — 信用卡月账单 Excel/CSV export
//
// Both importers self-register via init() so the HTTP layer in
// internal/server picks them up through a single blank import. The framework
// itself is format-agnostic — see internal/importer/importer.go.
//
// Sign convention (see ParsedTxn.Amount):
//
//	POSITIVE = money LEAVING source account (支出 / 消费)
//	NEGATIVE = money ENTERING source account (收入 / 退款 / 还款收到)
//
// The CMB debit CSV already encodes the sign in 收入(+)/支出(-); the credit
// statement records 人民币金额 as negative for spending and positive for
// refunds — both representations are normalised to the convention above.
//
// Cross-account hints: when the debit-side row looks like a credit-card
// repayment (对手户名 contains "信用卡" OR 交易摘要 contains "还款"), the
// importer suggests Liabilities:Credit:CMB as the counterpart so the two
// importers can be wired together as an internal transfer. The symmetric
// case on the credit side ("上期还款" / "还款") suggests Assets:Saving:CMB.
package cmb

import (
	"bytes"
	"strings"

	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/shopspring/decimal"
)

// init registers both flavours with the shared registry so the HTTP layer in
// internal/server picks them up through a single blank import. The framework
// contract is "subpackages self-register at init() time"; we honour it here.
func init() {
	importer.Register(CMBDebit{})
	importer.Register(CMBCredit{})
}

// utf8BOM is the byte-order-mark that some exports prepend. The CMB debit
// CSV from Web Banking does NOT include one in practice, but the desktop
// "全民生活" companion app sometimes does — strip defensively.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// xlsxMagic is the four-byte prefix shared by every XLSX (zip) file. Used
// only as a NEGATIVE signal in detection: we don't claim XLSX files unless
// the filename also looks like a CMB credit-card export.
var xlsxMagic = []byte{0x50, 0x4B, 0x03, 0x04}

// trimBOM removes a leading UTF-8 BOM if present.
func trimBOM(content []byte) []byte {
	return bytes.TrimPrefix(content, utf8BOM)
}

// hasCMBSignature reports whether the file body contains a string that real
// CMB exports use in their preamble. Cheap substring scan — we deliberately
// look at the WHOLE payload (not just the first line) because CMB puts the
// "招商银行" tag on different rows in different export variants.
func hasCMBSignature(content []byte) bool {
	return bytes.Contains(content, []byte("招商银行"))
}

// trimDecimal turns the assorted-string-amount representations CMB uses into
// a decimal. The export sometimes wraps numbers in double quotes (already
// stripped by encoding/csv), sometimes pads with whitespace, and very rarely
// prefixes "￥". An empty string returns a zero decimal and the OK bool is
// false so callers can distinguish "missing" from "0.00".
func trimDecimal(s string) (decimal.Decimal, bool, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "￥")
	s = strings.TrimPrefix(s, "¥")
	s = strings.TrimSpace(s)
	if s == "" {
		return decimal.Zero, false, nil
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, false, err
	}
	return d, true, nil
}

// suggestExpenseAccount returns a counterpart-account hint for an outgoing
// transaction based on cheap substring matches against payee/note. The
// vocabulary is shared with the alipay / wechat importers (M3-B / M3-C) so
// users see consistent hints regardless of which statement they import. The
// TF-IDF predictor in M3-F may further refine these at commit time.
func suggestExpenseAccount(payee, note string) string {
	hay := payee + " " + note
	type rule struct {
		needles []string
		account string
	}
	rules := []rule{
		{[]string{"星巴克", "瑞幸", "美团", "饿了么"}, "Expenses:Dining"},
		{[]string{"滴滴", "高德", "出租", "打车"}, "Expenses:Transport:Taxi"},
		{[]string{"京东", "淘宝", "拼多多", "天猫", "AMAZON", "亚马逊"}, "Expenses:Shopping"},
		{[]string{"中国移动", "中国联通", "中国电信"}, "Expenses:Utilities:Phone"},
		{[]string{"国家电网", "电费", "水费", "燃气", "煤气"}, "Expenses:Utilities"},
	}
	for _, r := range rules {
		for _, n := range r.needles {
			if strings.Contains(hay, n) {
				return r.account
			}
		}
	}
	return "Expenses:Unknown"
}

// suggestIncomeAccount returns the income-side counterpart hint. The two
// most common cases on a debit card are salary and refunds; everything else
// falls back to a generic Income bucket the user can refine in the preview UI.
func suggestIncomeAccount(payee, note string) string {
	hay := payee + " " + note
	switch {
	case strings.Contains(hay, "工资") || strings.Contains(hay, "薪资"):
		return "Income:Salary"
	case strings.Contains(hay, "利息"):
		return "Income:Interest"
	case strings.Contains(hay, "退款"):
		return "Income:Refund"
	}
	return "Income:Unknown"
}
