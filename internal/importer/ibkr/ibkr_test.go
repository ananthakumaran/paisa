package ibkr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// fixtureBytes reads testdata/sample.csv. The file is checked in with CRLF
// line endings on purpose: real IBKR Flex Query exports use DOS line endings,
// and the parser must handle them. Failing to handle CR would smuggle a
// trailing \r into every parsed field and silently break account suggestions.
func fixtureBytes(t *testing.T) []byte {
	t.Helper()
	p := filepath.Join("testdata", "sample.csv")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return b
}

// TestDetect exercises both detection paths: the magic ClientAccountID
// preamble (always at the very top of an IBKR file) and the Account
// Information section header (present even if the preamble is missing).
func TestDetect(t *testing.T) {
	imp := IBKR{}
	content := fixtureBytes(t)
	assert.True(t, imp.Detect("any.csv", content), "fixture must be detected")

	// Just the preamble — no other sections.
	preambleOnly := []byte(`ClientAccountID,U0000000,Description,"Activity Statement"` + "\r\n")
	assert.True(t, imp.Detect("any.csv", preambleOnly))

	// Just the Account Information header — no preamble.
	sectionOnly := []byte(`"Account Information","Header","Field Name","Field Value"` + "\r\n")
	assert.True(t, imp.Detect("any.csv", sectionOnly))

	// Negative case: a different bank's CSV must NOT match.
	other := []byte("date,payee,amount\r\n2024-01-02,Coffee,4.50\r\n")
	assert.False(t, imp.Detect("other.csv", other))
}

func TestImporterIdentity(t *testing.T) {
	imp := IBKR{}
	assert.Equal(t, "ibkr", imp.Code())
	assert.NotEmpty(t, imp.Name())
}

// TestSplitBySection makes sure the section dispatcher groups rows by their
// first column. Account Information / Statement / Cash Report rows must NOT
// be returned as transactions; only the transactional sections matter.
func TestSplitBySection(t *testing.T) {
	sections := splitBySection(fixtureBytes(t))
	assert.NotEmpty(t, sections["Trades"], "Trades section must be present")
	assert.NotEmpty(t, sections["Dividends"], "Dividends section must be present")
	assert.NotEmpty(t, sections["Withholding Tax"], "Withholding Tax section must be present")
	assert.NotEmpty(t, sections["Fees"], "Fees section must be present")
	assert.NotEmpty(t, sections["Interest"], "Interest section must be present")

	// Trades section should contain exactly the three Data rows from the
	// fixture (UBER buy, AAPL sell, 700 HK buy) plus the Header row.
	trades := sections["Trades"]
	assert.Equal(t, 4, len(trades), "Trades: 1 header + 3 data rows")
}

// TestParseTrades_Buy: positive Quantity = acquire. Per ParsedTxn convention,
// Amount is POSITIVE when money leaves the source account (cash → stock).
func TestParseTrades_Buy(t *testing.T) {
	imp := IBKR{}
	txns, err := imp.Parse(fixtureBytes(t))
	assert.NoError(t, err)

	var uber *importer.ParsedTxn
	for i := range txns {
		if txns[i].Payee == "Buy UBER" {
			uber = &txns[i]
			break
		}
	}
	if !assert.NotNil(t, uber, "expected a 'Buy UBER' txn") {
		return
	}
	assert.Equal(t, "USD", uber.Currency)
	// Proceeds = -450 (money out), commission -1 → total cost 451.
	assert.True(t, uber.Amount.Equal(decimal.NewFromFloat(451)),
		"buy amount should equal |proceeds| + |commission|, got %s", uber.Amount)
	assert.Equal(t, "Assets:Brokerage:IBKR:Stock:UBER", uber.SuggestedAccount)
	assert.Equal(t, 2024, uber.Date.Year())
	assert.Equal(t, 1, int(uber.Date.Month()))
	assert.Equal(t, 15, uber.Date.Day())
	assert.NotEmpty(t, uber.RawText)
}

// TestParseTrades_Sell: negative Quantity = dispose. The amount should be
// negative (money entering source account) so the import preview shows the
// proceeds correctly.
func TestParseTrades_Sell(t *testing.T) {
	imp := IBKR{}
	txns, err := imp.Parse(fixtureBytes(t))
	assert.NoError(t, err)

	var aapl *importer.ParsedTxn
	for i := range txns {
		if txns[i].Payee == "Sell AAPL" {
			aapl = &txns[i]
			break
		}
	}
	if !assert.NotNil(t, aapl, "expected a 'Sell AAPL' txn") {
		return
	}
	assert.Equal(t, "USD", aapl.Currency)
	// Proceeds = 950 (money in), commission -1 → net 949 entering source.
	// Per sign convention, money entering source = NEGATIVE amount.
	assert.True(t, aapl.Amount.Equal(decimal.NewFromFloat(-949)),
		"sell amount should be negative net proceeds, got %s", aapl.Amount)
	assert.Equal(t, "Assets:Brokerage:IBKR:Stock:AAPL", aapl.SuggestedAccount)
}

// TestParseTrades_NonUSD verifies currency is preserved (no auto-conversion).
func TestParseTrades_NonUSD(t *testing.T) {
	imp := IBKR{}
	txns, err := imp.Parse(fixtureBytes(t))
	assert.NoError(t, err)

	var hk *importer.ParsedTxn
	for i := range txns {
		if txns[i].Payee == "Buy 700" {
			hk = &txns[i]
			break
		}
	}
	if !assert.NotNil(t, hk, "expected a 'Buy 700' (Tencent) txn") {
		return
	}
	assert.Equal(t, "HKD", hk.Currency, "currency must be preserved, not converted")
	assert.True(t, hk.Amount.Equal(decimal.NewFromFloat(32015)),
		"buy 100 @ 320 HKD + 15 comm = 32015 HKD, got %s", hk.Amount)
}

// TestParseDividends: dividend received = money entering source =
// NEGATIVE amount with SuggestedAccount under Income.
func TestParseDividends(t *testing.T) {
	imp := IBKR{}
	txns, err := imp.Parse(fixtureBytes(t))
	assert.NoError(t, err)

	var div *importer.ParsedTxn
	for i := range txns {
		if txns[i].SuggestedAccount == "Income:Dividend:AAPL" {
			div = &txns[i]
			break
		}
	}
	if !assert.NotNil(t, div, "expected an AAPL dividend txn") {
		return
	}
	assert.Equal(t, "USD", div.Currency)
	assert.True(t, div.Amount.Equal(decimal.NewFromFloat(-2.40)),
		"dividend should be -2.40 (money in), got %s", div.Amount)
	assert.Equal(t, 2024, div.Date.Year())
	assert.Equal(t, 25, div.Date.Day())
}

// TestParseWithholdingTax: withholding tax = money leaving source =
// POSITIVE amount, expense account.
func TestParseWithholdingTax(t *testing.T) {
	imp := IBKR{}
	txns, err := imp.Parse(fixtureBytes(t))
	assert.NoError(t, err)

	var tax *importer.ParsedTxn
	for i := range txns {
		if txns[i].SuggestedAccount == "Expenses:Tax:Foreign:Withholding" {
			tax = &txns[i]
			break
		}
	}
	if !assert.NotNil(t, tax, "expected a withholding tax txn") {
		return
	}
	assert.Equal(t, "USD", tax.Currency)
	assert.True(t, tax.Amount.Equal(decimal.NewFromFloat(0.36)),
		"withholding tax should be +0.36 (money out), got %s", tax.Amount)
}

// TestParseFees: brokerage fee = money leaving = POSITIVE amount.
func TestParseFees(t *testing.T) {
	imp := IBKR{}
	txns, err := imp.Parse(fixtureBytes(t))
	assert.NoError(t, err)

	var fee *importer.ParsedTxn
	for i := range txns {
		if txns[i].SuggestedAccount == "Expenses:Brokerage:IBKR:Fee" {
			fee = &txns[i]
			break
		}
	}
	if !assert.NotNil(t, fee, "expected a fee txn") {
		return
	}
	assert.True(t, fee.Amount.Equal(decimal.NewFromFloat(1.00)),
		"fee should be +1.00 (money out), got %s", fee.Amount)
}

// TestParseInterest: credit interest = money entering = NEGATIVE amount.
func TestParseInterest(t *testing.T) {
	imp := IBKR{}
	txns, err := imp.Parse(fixtureBytes(t))
	assert.NoError(t, err)

	var interest *importer.ParsedTxn
	for i := range txns {
		if txns[i].SuggestedAccount == "Income:Brokerage:IBKR:Interest" {
			interest = &txns[i]
			break
		}
	}
	if !assert.NotNil(t, interest, "expected an interest txn") {
		return
	}
	assert.True(t, interest.Amount.Equal(decimal.NewFromFloat(-0.50)),
		"credit interest should be -0.50 (money in), got %s", interest.Amount)
}

// TestParseCountAllSections sanity-checks the total parsed count so silent
// regressions (e.g. a section gets dropped) surface as a single failure.
// Expected: 3 trades + 1 dividend + 1 wht + 1 fee + 1 interest = 7.
func TestParseCountAllSections(t *testing.T) {
	imp := IBKR{}
	txns, err := imp.Parse(fixtureBytes(t))
	assert.NoError(t, err)
	assert.Equal(t, 7, len(txns),
		"expected 3 trades + 1 div + 1 wht + 1 fee + 1 interest, got %d", len(txns))
}

// TestParseEmpty: empty input shouldn't panic — just return no txns.
func TestParseEmpty(t *testing.T) {
	imp := IBKR{}
	txns, err := imp.Parse([]byte(""))
	assert.NoError(t, err)
	assert.Empty(t, txns)
}

// TestParseHandlesLFOnly: just in case someone normalises the file before
// upload, plain LF (no CR) must also parse identically.
func TestParseHandlesLFOnly(t *testing.T) {
	imp := IBKR{}
	lf := []byte(`"Trades","Header","DataDiscriminator","Asset Category","Currency","Symbol","Date/Time","Quantity","T. Price","C. Price","Proceeds","Comm/Fee","Basis","Realized P/L","MTM P/L","Code"
"Trades","Data","Order","Stocks","USD","UBER","2024-01-15;10:30:00","10","45.00","45.05","-450","-1.00","451","0","0.5","O"
`)
	txns, err := imp.Parse(lf)
	assert.NoError(t, err)
	if assert.Len(t, txns, 1) {
		assert.Equal(t, "Buy UBER", txns[0].Payee)
	}
}
