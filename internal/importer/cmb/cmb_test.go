package cmb_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/ananthakumaran/paisa/internal/importer/cmb"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// loadFixture reads a file from this package's testdata directory.
func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	bs, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return bs
}

// =============================================================================
// Codes / registration
// =============================================================================

func TestCodesAndNames(t *testing.T) {
	d := cmb.CMBDebit{}
	assert.Equal(t, "cmb-debit", d.Code())
	assert.NotEmpty(t, d.Name())

	c := cmb.CMBCredit{}
	assert.Equal(t, "cmb-credit", c.Code())
	assert.NotEmpty(t, c.Name())
}

// TestRegisteredOnInit: both importers must self-register via init() so that a
// blank import from internal/server wires them into the HTTP routes.
func TestRegisteredOnInit(t *testing.T) {
	if assert.NotNil(t, importer.ByCode("cmb-debit"), "cmb-debit must register via init()") {
		assert.Equal(t, "cmb-debit", importer.ByCode("cmb-debit").Code())
	}
	if assert.NotNil(t, importer.ByCode("cmb-credit"), "cmb-credit must register via init()") {
		assert.Equal(t, "cmb-credit", importer.ByCode("cmb-credit").Code())
	}
}

// =============================================================================
// Debit detection
// =============================================================================

// TestDebitDetectByHeader: real exports contain both the bank name and the
// signed-amount column header. Either combo is enough — the importer is
// generous on detection so renamed files still match.
func TestDebitDetectByHeader(t *testing.T) {
	d := cmb.CMBDebit{}
	content := loadFixture(t, "debit-sample.csv")
	assert.True(t, d.Detect("random.csv", content), "expected detect by header content")
}

// TestDebitDetectByFilename: filename hint catches files whose body we can't
// inspect (binary, mis-encoded). Both English "cmb_debit" and Chinese "招行借记"
// must match.
func TestDebitDetectByFilename(t *testing.T) {
	d := cmb.CMBDebit{}
	assert.True(t, d.Detect("cmb_debit_2024_01.csv", []byte("garbage")))
	assert.True(t, d.Detect("招行借记卡202401.csv", []byte("garbage")))
}

// TestDebitDetectNoMatch: unrelated CSV must not match — false positives
// crowd the picker.
func TestDebitDetectNoMatch(t *testing.T) {
	d := cmb.CMBDebit{}
	assert.False(t, d.Detect("bank.csv", []byte("date,amount\n2024-01-01,10\n")))
}

// TestDebitDoesNotMatchCreditFilename: the credit-card branch's hints must
// not light up the debit importer.
func TestDebitDoesNotMatchCreditFilename(t *testing.T) {
	d := cmb.CMBDebit{}
	assert.False(t, d.Detect("cmb_credit_2024_01.xlsx", []byte("garbage")))
}

// =============================================================================
// Credit detection
// =============================================================================

func TestCreditDetectByFilename(t *testing.T) {
	c := cmb.CMBCredit{}
	assert.True(t, c.Detect("cmb_credit_2024_01.xlsx", []byte("anything")))
	assert.True(t, c.Detect("招商银行信用卡账单.xlsx", []byte("anything")))
	assert.True(t, c.Detect("信用卡202401.csv", []byte("anything")))
}

// TestCreditDetectByXLSXMagic: when filename hint is ambiguous, the importer
// can still match on the XLSX (zip) magic-byte prefix combined with a CMB
// signature. We intentionally require BOTH — bare XLSX from any other bank
// must NOT match.
func TestCreditDetectByXLSXMagic(t *testing.T) {
	c := cmb.CMBCredit{}
	xlsxMagic := []byte{0x50, 0x4B, 0x03, 0x04}
	bareXLSX := append([]byte{}, xlsxMagic...)
	bareXLSX = append(bareXLSX, []byte("\x00\x00random.xlsx data\n")...)
	assert.False(t, c.Detect("random.xlsx", bareXLSX), "bare xlsx must not match — filename hint required")
}

// TestCreditDetectByCSVHeader: many users export the credit-card statement as
// CSV (the easier path until we add full XLSX support). The header row is a
// reliable signature.
func TestCreditDetectByCSVHeader(t *testing.T) {
	c := cmb.CMBCredit{}
	content := loadFixture(t, "credit-sample.csv")
	assert.True(t, c.Detect("random.csv", content))
}

// TestCreditDetectNoMatch: random CSV without the CMB credit signature must
// not match.
func TestCreditDetectNoMatch(t *testing.T) {
	c := cmb.CMBCredit{}
	assert.False(t, c.Detect("bank.csv", []byte("date,amount\n2024-01-01,10\n")))
}

// =============================================================================
// Debit parse
// =============================================================================

// TestDebitParse: the fixture has 6 rows including an outgoing payment to a
// merchant, a salary income, a credit-card repayment (which triggers the
// counterpart-account hint), and a refund. Verify all of them.
func TestDebitParse(t *testing.T) {
	d := cmb.CMBDebit{}
	content := loadFixture(t, "debit-sample.csv")
	txns, err := d.Parse(content)
	assert.NoError(t, err)
	if !assert.Len(t, txns, 6) {
		for i, tx := range txns {
			t.Logf("txn[%d] payee=%s amount=%s suggested=%s", i, tx.Payee, tx.Amount, tx.SuggestedAccount)
		}
		return
	}

	// Row 1: 星巴克 -38.00 → expense, sign convention: positive
	assert.Equal(t, "星巴克", txns[0].Payee)
	assert.True(t, txns[0].Amount.Equal(decimal.NewFromFloat(38.00)), "want 38.00 got %s", txns[0].Amount)
	assert.Equal(t, "CNY", txns[0].Currency)
	assert.Equal(t, 2024, txns[0].Date.Year())
	assert.Equal(t, 1, int(txns[0].Date.Month()))
	assert.Equal(t, 15, txns[0].Date.Day())
	assert.Equal(t, "Expenses:Dining", txns[0].SuggestedAccount)
	assert.NotEmpty(t, txns[0].RawText)

	// Row 2: XXX有限公司 8000.00 (income) → sign NEGATIVE
	assert.Equal(t, "XXX有限公司", txns[1].Payee)
	assert.True(t, txns[1].Amount.Equal(decimal.NewFromFloat(-8000.00)), "want -8000.00 got %s", txns[1].Amount)
	assert.Equal(t, "Income:Salary", txns[1].SuggestedAccount)

	// Row 3: 滴滴出行 -128.50 → Expenses:Transport:Taxi
	assert.Equal(t, "滴滴出行", txns[2].Payee)
	assert.True(t, txns[2].Amount.Equal(decimal.NewFromFloat(128.50)))
	assert.Equal(t, "Expenses:Transport:Taxi", txns[2].SuggestedAccount)

	// Row 4: 信用卡还款 -2000 → counterpart Liabilities:Credit:CMB (internal transfer)
	assert.Equal(t, "招行信用卡", txns[3].Payee)
	assert.True(t, txns[3].Amount.Equal(decimal.NewFromFloat(2000.00)))
	assert.Equal(t, "Liabilities:Credit:CMB", txns[3].SuggestedAccount)

	// Row 5: 拼多多 -99.99 → Expenses:Shopping
	assert.Equal(t, "拼多多", txns[4].Payee)
	assert.True(t, txns[4].Amount.Equal(decimal.NewFromFloat(99.99)))
	assert.Equal(t, "Expenses:Shopping", txns[4].SuggestedAccount)

	// Row 6: 匿名好友A 500.00 (income, 还款) → still negative (money in)
	assert.Equal(t, "匿名好友A", txns[5].Payee)
	assert.True(t, txns[5].Amount.Equal(decimal.NewFromFloat(-500.00)))
}

// TestDebitParseEmpty: an empty body must return an error, not panic.
func TestDebitParseEmpty(t *testing.T) {
	d := cmb.CMBDebit{}
	_, err := d.Parse([]byte(""))
	assert.Error(t, err)
}

// TestDebitParseRejectsCreditFormat: feeding a credit-card file to the debit
// parser must error rather than silently produce garbage rows.
func TestDebitParseRejectsCreditFormat(t *testing.T) {
	d := cmb.CMBDebit{}
	content := loadFixture(t, "credit-sample.csv")
	_, err := d.Parse(content)
	assert.Error(t, err)
}

// =============================================================================
// Credit parse
// =============================================================================

// TestCreditParse: the fixture has 7 rows — domestic spending, a USD foreign
// transaction (keep original currency), a refund (sign flips), and a
// repayment row that we drop.
func TestCreditParse(t *testing.T) {
	c := cmb.CMBCredit{}
	content := loadFixture(t, "credit-sample.csv")
	txns, err := c.Parse(content)
	assert.NoError(t, err)

	// We KEEP all 7 rows. The "上期还款" row is preserved but with a
	// counterpart hint of Assets:Saving:CMB so it nets out against the
	// debit-side "信用卡还款" entry once both files are imported.
	if !assert.Len(t, txns, 7) {
		for i, tx := range txns {
			t.Logf("txn[%d] payee=%s amount=%s currency=%s suggested=%s", i, tx.Payee, tx.Amount, tx.Currency, tx.SuggestedAccount)
		}
		return
	}

	// Row 1: 星巴克咖啡 -38.00 RMB
	assert.Equal(t, "星巴克咖啡", txns[0].Payee)
	assert.True(t, txns[0].Amount.Equal(decimal.NewFromFloat(38.00)), "want 38.00 got %s", txns[0].Amount)
	assert.Equal(t, "CNY", txns[0].Currency)
	assert.Equal(t, "Expenses:Dining", txns[0].SuggestedAccount)
	assert.Equal(t, 2024, txns[0].Date.Year())
	assert.Equal(t, 1, int(txns[0].Date.Month()))
	assert.Equal(t, 5, txns[0].Date.Day())

	// Row 2: 美团外卖 -58.50 RMB → Expenses:Dining
	assert.Equal(t, "美团外卖", txns[1].Payee)
	assert.True(t, txns[1].Amount.Equal(decimal.NewFromFloat(58.50)))
	assert.Equal(t, "Expenses:Dining", txns[1].SuggestedAccount)

	// Row 3: AMAZON foreign currency — keep USD, value 50.00 USD positive
	// (positive = money leaving source).
	assert.Equal(t, "AMAZON", txns[2].Payee)
	assert.True(t, txns[2].Amount.Equal(decimal.NewFromFloat(50.00)), "want 50.00 got %s", txns[2].Amount)
	assert.Equal(t, "USD", txns[2].Currency)

	// Row 5 (index 4): 退款-星巴克 +20.00 → refund, sign NEGATIVE
	assert.Equal(t, "退款-星巴克", txns[4].Payee)
	assert.True(t, txns[4].Amount.Equal(decimal.NewFromFloat(-20.00)), "want -20.00 got %s", txns[4].Amount)

	// Row 7 (index 6): 上期还款 +2000 → counterpart should hint at Assets:Saving:CMB
	assert.Equal(t, "上期还款", txns[6].Payee)
	assert.True(t, txns[6].Amount.Equal(decimal.NewFromFloat(-2000.00)))
	assert.Equal(t, "Assets:Saving:CMB", txns[6].SuggestedAccount)
}

// TestCreditParseEmpty: an empty body must return an error, not panic.
func TestCreditParseEmpty(t *testing.T) {
	c := cmb.CMBCredit{}
	_, err := c.Parse([]byte(""))
	assert.Error(t, err)
}
