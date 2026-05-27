package alipay_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/ananthakumaran/paisa/internal/importer/alipay"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// loadFixture reads the GBK-encoded sample CSV that the rest of the package
// tests are built around. Keeping the read in one helper means a missing or
// corrupted fixture surfaces as a single, obvious failure.
func loadFixture(t *testing.T) []byte {
	t.Helper()
	path := filepath.Join("testdata", "sample.csv")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return b
}

// TestCodeAndName pins the public identifiers. The frontend persists the
// importer code in user preferences, so a typo here would silently break
// existing installs after upgrade.
func TestCodeAndName(t *testing.T) {
	a := alipay.Alipay{}
	assert.Equal(t, "alipay", a.Code())
	assert.NotEmpty(t, a.Name(), "Name() must be non-empty for the UI")
}

// TestDetectByHeader exercises the content-based detection branch using the
// GBK-encoded fixture. The header "交易记录明细列表" is the canonical marker
// that survives both GBK and UTF-8 exports.
func TestDetectByHeader(t *testing.T) {
	a := alipay.Alipay{}
	content := loadFixture(t)
	assert.True(t, a.Detect("anything.csv", content),
		"Detect must match alipay header substring in GBK content")
}

// TestDetectByFilename matches the documented filename hint ("alipay_record_*")
// even if the file body is garbage. Mirrors the stub importer's loose-filename
// branch — the user can always override in the UI.
func TestDetectByFilename(t *testing.T) {
	a := alipay.Alipay{}
	assert.True(t, a.Detect("alipay_record_2024.csv", []byte("not a real alipay file")),
		"Detect must match filename containing alipay_record")
}

// TestDetectNoMatch confirms unrelated files produce false.
func TestDetectNoMatch(t *testing.T) {
	a := alipay.Alipay{}
	assert.False(t, a.Detect("bank.csv", []byte("col1,col2\n1,2\n")))
}

// TestParseFixtureCounts asserts the row-count contract: out of 8 rows in
// the fixture, only the ones with a successful status AND a non-empty
// 收/支 column should produce a ParsedTxn. That filters out the
// "交易关闭" failed coffee and the "0金额内部转账" balance line.
func TestParseFixtureCounts(t *testing.T) {
	a := alipay.Alipay{}
	txns, err := a.Parse(loadFixture(t))
	if !assert.NoError(t, err) {
		return
	}
	// Fixture composition: 8 rows total → 6 valid:
	//   - 4 expenses (Starbucks, Didi, JD, transfer)
	//   - 1 refund (income)
	//   - 1 salary (income)
	// dropped: 1 "交易关闭" + 1 "0 amount, no sign" internal transfer
	assert.Len(t, txns, 6)
}

// TestParseFirstRow verifies the field-by-field mapping for a typical
// expense row: GBK-decoded payee, correct sign (negative = money leaving
// source), date parsed at midnight in local time, and currency set to CNY.
func TestParseFirstRow(t *testing.T) {
	a := alipay.Alipay{}
	txns, err := a.Parse(loadFixture(t))
	if !assert.NoError(t, err) || !assert.NotEmpty(t, txns) {
		return
	}
	first := txns[0]
	assert.Equal(t, "星巴克咖啡", first.Payee)
	assert.Equal(t, "星巴克拿铁大杯", first.Note)
	assert.True(t, first.Amount.Equal(decimal.NewFromFloat(38.00)),
		"expense amount must be positive (money leaving source), got %s", first.Amount)
	assert.Equal(t, "CNY", first.Currency)
	assert.Equal(t, 2024, first.Date.Year())
	assert.Equal(t, 1, int(first.Date.Month()))
	assert.Equal(t, 2, first.Date.Day())
	assert.NotEmpty(t, first.RawText, "RawText must preserve the source line for the UI")
}

// TestParseRefundIsIncome ensures a 收入 row (refund) maps to a NEGATIVE
// amount per the ParsedTxn convention (money entering the source account).
func TestParseRefundIsIncome(t *testing.T) {
	a := alipay.Alipay{}
	txns, err := a.Parse(loadFixture(t))
	if !assert.NoError(t, err) {
		return
	}
	var refund *importer.ParsedTxn
	for i := range txns {
		if strings.Contains(txns[i].Note, "退款") {
			refund = &txns[i]
			break
		}
	}
	if !assert.NotNil(t, refund, "expected a refund row to be present") {
		return
	}
	assert.True(t, refund.Amount.IsNegative(),
		"refund amount must be negative (money entering source), got %s", refund.Amount)
	assert.True(t, refund.Amount.Equal(decimal.NewFromFloat(-38.00)))
}

// TestParseSalaryIsIncome ensures the 工资 row produces a negative-amount
// txn AND that the SuggestedAccount falls under Income:Salary.
func TestParseSalaryIsIncome(t *testing.T) {
	a := alipay.Alipay{}
	txns, err := a.Parse(loadFixture(t))
	if !assert.NoError(t, err) {
		return
	}
	var salary *importer.ParsedTxn
	for i := range txns {
		if strings.Contains(txns[i].Note, "工资") {
			salary = &txns[i]
			break
		}
	}
	if !assert.NotNil(t, salary, "expected a salary row to be present") {
		return
	}
	assert.True(t, salary.Amount.IsNegative(),
		"salary income must be negative per ParsedTxn convention, got %s", salary.Amount)
	assert.True(t, salary.Amount.Equal(decimal.NewFromFloat(-8500.00)))
	assert.Equal(t, "Income:Salary", salary.SuggestedAccount)
}

// TestStatusFilterDropsClosed makes sure the "交易关闭" Starbucks row at
// 2024-01-10 is filtered out so users don't see ghost transactions.
func TestStatusFilterDropsClosed(t *testing.T) {
	a := alipay.Alipay{}
	txns, err := a.Parse(loadFixture(t))
	if !assert.NoError(t, err) {
		return
	}
	for _, tx := range txns {
		if tx.Date.Month() == 1 && tx.Date.Day() == 10 {
			t.Fatalf("交易关闭 row was not filtered: %+v", tx)
		}
	}
}

// TestZeroAmountInternalTransferDropped covers the balance-shuffle row with
// blank 收/支 and 金额=0 (e.g. 余额宝 auto-transfer entries). These have no
// real impact on the user's ledger and must be skipped.
func TestZeroAmountInternalTransferDropped(t *testing.T) {
	a := alipay.Alipay{}
	txns, err := a.Parse(loadFixture(t))
	if !assert.NoError(t, err) {
		return
	}
	for _, tx := range txns {
		if tx.Date.Month() == 1 && tx.Date.Day() == 20 {
			t.Fatalf("0-amount internal transfer was not filtered: %+v", tx)
		}
	}
}

// TestSuggestedAccountHeuristics exercises the merchant-keyword mapping.
// We look up by (payeeContains, sign) because the fixture deliberately
// reuses the 星巴克 payee across an expense AND a refund — a flat map
// would shadow one of them.
func TestSuggestedAccountHeuristics(t *testing.T) {
	a := alipay.Alipay{}
	txns, err := a.Parse(loadFixture(t))
	if !assert.NoError(t, err) {
		return
	}
	find := func(payeeContains string, wantOutflow bool) *importer.ParsedTxn {
		for i := range txns {
			if !strings.Contains(txns[i].Payee, payeeContains) {
				continue
			}
			if wantOutflow && txns[i].Amount.IsPositive() {
				return &txns[i]
			}
			if !wantOutflow && txns[i].Amount.IsNegative() {
				return &txns[i]
			}
		}
		return nil
	}
	if tx := find("星巴克", true); assert.NotNil(t, tx, "expected starbucks expense") {
		assert.Equal(t, "Expenses:Dining", tx.SuggestedAccount)
	}
	if tx := find("滴滴", true); assert.NotNil(t, tx, "expected didi expense") {
		assert.Equal(t, "Expenses:Transport:Taxi", tx.SuggestedAccount)
	}
	if tx := find("京东", true); assert.NotNil(t, tx, "expected jd expense") {
		assert.Equal(t, "Expenses:Shopping", tx.SuggestedAccount)
	}
	if tx := find("张三", true); assert.NotNil(t, tx, "expected 张三 transfer") {
		assert.Equal(t, "Expenses:Unknown", tx.SuggestedAccount)
	}
}

// TestParseUTF8Input accepts a UTF-8 (no BOM) version of the same data —
// some recent alipay exports drop GBK. We encode the fixture roundtrip from
// GBK to UTF-8 in-test to keep both branches covered without shipping two
// fixtures.
func TestParseUTF8Input(t *testing.T) {
	a := alipay.Alipay{}
	gbk := loadFixture(t)
	dec := simplifiedchinese.GBK.NewDecoder()
	utf8, err := dec.Bytes(gbk)
	if !assert.NoError(t, err) {
		return
	}
	assert.True(t, a.Detect("anything.csv", utf8), "Detect must also recognize UTF-8 alipay exports")
	txns, err := a.Parse(utf8)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, txns)
	}
}

// TestParseUTF8WithBOM covers a third encoding variant: UTF-8 with a leading
// byte-order-mark, which Excel-on-Windows tends to add when re-saving.
func TestParseUTF8WithBOM(t *testing.T) {
	a := alipay.Alipay{}
	gbk := loadFixture(t)
	dec := simplifiedchinese.GBK.NewDecoder()
	utf8, err := dec.Bytes(gbk)
	if !assert.NoError(t, err) {
		return
	}
	bom := []byte{0xEF, 0xBB, 0xBF}
	withBOM := append(bom, utf8...)
	txns, err := a.Parse(withBOM)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, txns)
		// First payee must NOT contain a stray BOM character or garbled bytes.
		// 0xFEFF is the BOM code point; a leftover indicates the BOM leaked
		// through the encoding step.
		for _, tx := range txns {
			assert.False(t, bytes.ContainsRune([]byte(tx.Payee), '\uFEFF'),
				"BOM must be stripped before parsing, got payee %q", tx.Payee)
		}
	}
}

// TestParseEmptyFileErrors makes sure an empty / completely unrecognized
// payload produces an informative error, not a panic or a nil slice masking
// a problem.
func TestParseEmptyFileErrors(t *testing.T) {
	a := alipay.Alipay{}
	_, err := a.Parse([]byte(""))
	assert.Error(t, err)
}

// TestSelfRegistration is the wiring contract for the HTTP layer: the
// alipay package must register itself in the global importer registry at
// init() so that internal/server/import.go discovers it without any further
// glue. Mirrors the documented pattern at the top of internal/importer.
func TestSelfRegistration(t *testing.T) {
	found := importer.ByCode("alipay")
	assert.NotNil(t, found, "alipay must self-register via init()")
}
