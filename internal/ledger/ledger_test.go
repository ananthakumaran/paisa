package ledger

import (
	"testing"

	"github.com/ananthakumaran/paisa/internal/model/price"
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

	parsedPrices, _ = parseHLedgerPrices("P 2023-05-01 USD 0.9 EUR\n", "INR")
	assert.Len(t, parsedPrices, 0)
	parsedPrices, _ = parseHLedgerPrices("P 2023-05-01 USD $0.9\n", "INR")
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
}
