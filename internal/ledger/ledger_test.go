package ledger

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func assertPriceEqual(t *testing.T, actual price.Price, date string, commodityName string, value float64) {
	assert.Equal(t, commodityName, actual.CommodityName, "they should be equal")
	assert.Equal(t, date, actual.Date.Format("2006/01/02"), "they should be equal")
	assert.Equal(t, value, actual.Value.InexactFloat64(), "they should be equal")
}

func TestParseLegerPrices(t *testing.T) {
	parsedPrices, _ := parseLedgerPrices("P 2023/05/01 00:00:00 USD 0.9 EUR\n", "EUR")
	assertPriceEqual(t, parsedPrices[0], "2023/05/01", "USD", 0.9)
	parsedPrices, _ = parseLedgerPrices("P 2023/05/01 00:00:00 EUR $1.1\n", "$")
	assertPriceEqual(t, parsedPrices[0], "2023/05/01", "EUR", 1.1)
	parsedPrices, _ = parseLedgerPrices("P 2023/05/01 00:00:00 EUR $-1.1\n", "$")
	assertPriceEqual(t, parsedPrices[0], "2023/05/01", "EUR", -1.1)
	parsedPrices, _ = parseLedgerPrices("P 2023/05/01 00:00:00 EUR ₹70\n", "₹")
	assertPriceEqual(t, parsedPrices[0], "2023/05/01", "EUR", 70)

	parsedPrices, _ = parseLedgerPrices("P 2023/05/01 00:00:00 USD 0.9 EUR\n", "INR")
	assert.Len(t, parsedPrices, 0)
	parsedPrices, _ = parseLedgerPrices("P 2023/05/01 00:00:00 USD $0.9\n", "INR")
	assert.Len(t, parsedPrices, 0)

	parsedPrices, _ = parseLedgerPrices("P 2022/01/29 00:50:00 UAH 0.026 EUR\n", "EUR")
	assertPriceEqual(t, parsedPrices[0], "2022/01/29", "UAH", 0.026)
}

func TestParseHLegerPrices(t *testing.T) {
	parsedPrices, _ := parseHLedgerPrices("P 2023-05-01 USD 0.9 EUR\n", "EUR")
	assertPriceEqual(t, parsedPrices[0], "2023/05/01", "USD", 0.9)
	parsedPrices, _ = parseHLedgerPrices("P 2023-05-01 EUR $1.1\n", "$")
	assertPriceEqual(t, parsedPrices[0], "2023/05/01", "EUR", 1.1)

	parsedPrices, _ = parseHLedgerPrices("P 2023-05-01 EUR USD 1.1\n", "USD")
	assertPriceEqual(t, parsedPrices[0], "2023/05/01", "EUR", 1.1)

	parsedPrices, _ = parseHLedgerPrices("P 2023-05-01 EUR 1.1$\n", "$")
	assertPriceEqual(t, parsedPrices[0], "2023/05/01", "EUR", 1.1)

	parsedPrices, _ = parseHLedgerPrices(utils.Dos2Unix("P 2023-05-01 EUR 1.1$\r\n"), "$")
	assertPriceEqual(t, parsedPrices[0], "2023/05/01", "EUR", 1.1)

	parsedPrices, _ = parseHLedgerPrices("P 2023-05-01 \"AAPL0\" \"USD0\" 45.5\n", "USD0")
	assertPriceEqual(t, parsedPrices[0], "2023/05/01", "AAPL0", 45.5)

	parsedPrices, _ = parseHLedgerPrices("P 2023-05-01 USD 0.9 EUR\n", "INR")
	assert.Len(t, parsedPrices, 0)

	parsedPrices, _ = parseHLedgerPrices("P 2023-05-01 USD $0.9\n", "INR")
	assert.Len(t, parsedPrices, 0)

	parsedPrices, _ = parseHLedgerPrices("P 2023-05-01 USD $0.9\r\n", "INR")
	assert.Len(t, parsedPrices, 0)
}

func TestParseAmount(t *testing.T) {
	commodity, amount, _ := parseAmount("0.9 USD")
	assert.Equal(t, "USD", commodity)
	assert.Equal(t, 0.9, amount.InexactFloat64())

	commodity, amount, _ = parseAmount("$0.9")
	assert.Equal(t, "$", commodity)
	assert.Equal(t, 0.9, amount.InexactFloat64())

	commodity, amount, _ = parseAmount("0.9$")
	assert.Equal(t, "$", commodity)
	assert.Equal(t, 0.9, amount.InexactFloat64())

	commodity, amount, _ = parseAmount("$-0.9")
	assert.Equal(t, "$", commodity)
	assert.Equal(t, -0.9, amount.InexactFloat64())

	commodity, amount, _ = parseAmount("-0.9$")
	assert.Equal(t, "$", commodity)
	assert.Equal(t, -0.9, amount.InexactFloat64())

	commodity, amount, _ = parseAmount("100,000 EUR")
	assert.Equal(t, "EUR", commodity)
	assert.Equal(t, 100000.0, amount.InexactFloat64())

	commodity, amount, _ = parseAmount("100,000.00 \"EUR0-0\"")
	assert.Equal(t, "EUR0-0", commodity)
	assert.Equal(t, 100000.0, amount.InexactFloat64())

	commodity, amount, _ = parseAmount("-100,000.00 \"EUR0-0\"")
	assert.Equal(t, "EUR0-0", commodity)
	assert.Equal(t, -100000.0, amount.InexactFloat64())

	commodity, amount, _ = parseAmount("\"EUR0-0\" -100,000.00")
	assert.Equal(t, "EUR0-0", commodity)
	assert.Equal(t, -100000.0, amount.InexactFloat64())

	commodity, amount, _ = parseAmount("INR 70.0099")
	assert.Equal(t, "INR", commodity)
	assert.Equal(t, 70.0099, amount.InexactFloat64())

	commodity, amount, _ = parseAmount("1E-8 BTC")
	assert.Equal(t, "BTC", commodity)
	assert.Equal(t, 1e-08, amount.InexactFloat64())

	commodity, amount, _ = parseAmount("100E-8    BTC")
	assert.Equal(t, "BTC", commodity)
	assert.Equal(t, 1e-06, amount.InexactFloat64())
}

func TestBuildHLedgerPostingsPrice(t *testing.T) {
	dir := t.TempDir()
	assert.NoError(t, config.LoadConfig([]byte("journal_path: main.ledger\ndb_path: paisa.db"), filepath.Join(dir, "paisa.yaml")))

	for _, version := range []struct{ field, unit, total string }{
		{"aprice", "UnitPrice", "TotalPrice"},
		{"acost", "UnitCost", "TotalCost"},
	} {
		for _, tc := range []struct {
			quantity float64
			price    string
			amount   float64
		}{
			{10, fmt.Sprintf(`{"tag": "%s", "contents": {"acommodity": "INR", "aquantity": {"floatingPoint": 100}}}`, version.unit), 1000},
			{-10, fmt.Sprintf(`{"tag": "%s", "contents": {"acommodity": "INR", "aquantity": {"floatingPoint": 100}}}`, version.unit), -1000},
			{10, fmt.Sprintf(`{"tag": "%s", "contents": {"acommodity": "INR", "aquantity": {"floatingPoint": 900}}}`, version.total), 900},
			{-10, fmt.Sprintf(`{"tag": "%s", "contents": {"acommodity": "INR", "aquantity": {"floatingPoint": -900}}}`, version.total), -900},
			{10, fmt.Sprintf(`{"tag": "%s", "contents": {"acommodity": "INR", "aquantity": {"floatingPoint": 0}}}`, version.unit), 0},
			{10, fmt.Sprintf(`{"tag": "%s", "contents": {"acommodity": "USD", "aquantity": {"floatingPoint": 2}}}`, version.unit), 1600},
			{10, "null", 10},
		} {
			var p HLedgerPosting
			assert.NoError(t, json.Unmarshal([]byte(fmt.Sprintf(`{"paccount": "Assets:Equity:ACME", "pamount": [{"acommodity": "ACME", "aquantity": {"floatingPoint": %f}, "%s": %s}]}`, tc.quantity, version.field, tc.price)), &p))
			var transaction HLedgerTransaction
			assert.NoError(t, json.Unmarshal([]byte(fmt.Sprintf(`{"tsourcepos": [{"sourceName": %q}, {"sourceName": %q}]}`, filepath.Join(dir, "main.ledger"), filepath.Join(dir, "main.ledger"))), &transaction))

			date := time.Date(2020, 1, 1, 0, 0, 0, 0, config.TimeZone())
			usd := price.Price{Date: date, CommodityName: "USD", Value: decimal.NewFromInt(80)}
			postings, err := buildHLedgerPostings(p, transaction, buildPricesTree([]price.Price{usd}), date)
			assert.NoError(t, err)
			assert.Len(t, postings, 1)
			assert.Equal(t, tc.amount, postings[0].Amount.InexactFloat64(), "%s %s", version.field, tc.price)
		}
	}
}
