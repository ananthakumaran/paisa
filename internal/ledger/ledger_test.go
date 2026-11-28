package ledger

import (
	"testing"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCliSelectsConfiguredEngine(t *testing.T) {
	for _, tc := range []struct {
		name, config string
		want         Ledger
	}{
		{"default hledger", "", HLedgerCLI{}},
		{"hledger", "ledger_cli: hledger\n", HLedgerCLI{}},
		{"ledger", "ledger_cli: ledger\n", LedgerCLI{}},
		{"beancount", "ledger_cli: beancount\n", Beancount{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content := []byte("journal_path: main.ledger\ndb_path: paisa.db\n" + tc.config)
			require.NoError(t, config.LoadConfig(content, ""))
			require.IsType(t, tc.want, Cli())
		})
	}
	// Leave package-global configuration in its fresh-install state for other tests.
	require.NoError(t, config.LoadConfig([]byte("journal_path: main.ledger\ndb_path: paisa.db\n"), ""))
}

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

func TestParsePricesCommodityNames(t *testing.T) {
	parsers := []struct {
		name   string
		prefix string
		parse  func(string, string) ([]price.Price, error)
	}{
		{"Ledger", "P 2022/02/01 00:00:00 ", parseLedgerPrices},
		{"hledger", "P 2022-02-01 ", parseHLedgerPrices},
	}
	cases := []struct {
		name      string
		fields    string
		currency  string
		commodity string
		value     float64
	}{
		{"Zerodha fund", `"AXIS SMALL CAP FUND - DIRECT PLAN" 12.9 INR`, "INR", "AXIS SMALL CAP FUND - DIRECT PLAN", 12.9},
		{"space without punctuation", `"Small Cap" 12.9 INR`, "INR", "Small Cap", 12.9},
		{"consecutive spaces", `"Small  Cap" 12.9 INR`, "INR", "Small  Cap", 12.9},
		{"Unicode and punctuation", `"Épargne & Revenu" 12.9 EUR`, "EUR", "Épargne & Revenu", 12.9},
		{"starts with digits", `"2025 Growth Fund" 12.9 INR`, "INR", "2025 Growth Fund", 12.9},
		{"punctuation without spaces", `"FUND-2025.A" 12.9 INR`, "INR", "FUND-2025.A", 12.9},
		{"quoted currency after amount", `"Small Cap" 1.25 "Quote Currency 2"`, "Quote Currency 2", "Small Cap", 1.25},
		{"quoted currency before amount", `"Small Cap" "Quote Currency 2" 1.25`, "Quote Currency 2", "Small Cap", 1.25},
		{"bare commodity", `NIFTY 12.9 INR`, "INR", "NIFTY", 12.9},
	}

	for _, parser := range parsers {
		t.Run(parser.name, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					prices, err := parser.parse(parser.prefix+tc.fields+"\n", tc.currency)
					require.NoError(t, err)
					require.Len(t, prices, 1)
					assertPriceEqual(t, prices[0], "2022/02/01", tc.commodity, tc.value)
					assert.Equal(t, tc.commodity, prices[0].CommodityID)
				})
			}

			t.Run("mixed quoted and bare rows", func(t *testing.T) {
				output := parser.prefix + "NIFTY 100 INR\n" +
					parser.prefix + "\"Small Cap\" 12.9 INR\n" +
					parser.prefix + "SENSEX 200 INR\n"
				prices, err := parser.parse(output, "INR")
				require.NoError(t, err)
				require.Len(t, prices, 3)
				assertPriceEqual(t, prices[0], "2022/02/01", "NIFTY", 100)
				assertPriceEqual(t, prices[1], "2022/02/01", "Small Cap", 12.9)
				assertPriceEqual(t, prices[2], "2022/02/01", "SENSEX", 200)
			})
		})
	}
}

func TestParseBeanQueryPrices(t *testing.T) {
	const header = "date,currency,str(number(amount)),currency(amount)\r\n"
	data := header +
		"2022-01-07,NIFTY,100.273,INR\r\n" +
		"2022-01-08,INR,0.0125,USD\r\n" +
		"2022-01-09,ABC1DE,1E-8,INR\r\n" +
		"2022-01-10,INR,0,EUR\r\n" +
		"2022-01-11,USD,0.9,EUR\r\n"
	prices, err := parseBeanQueryPrices([]byte(data), "INR")
	require.NoError(t, err)
	require.Len(t, prices, 3)
	assertPriceEqual(t, prices[0], "2022/01/07", "NIFTY", 100.273)
	assertPriceEqual(t, prices[1], "2022/01/08", "USD", 80)
	assertPriceEqual(t, prices[2], "2022/01/09", "ABC1DE", 1e-8)
	assert.Equal(t, "0.00000001", prices[2].Value.String())

	prices, err = parseBeanQueryPrices([]byte(header), "INR")
	require.NoError(t, err)
	assert.Empty(t, prices)

	_, err = parseBeanQueryPrices(nil, "INR")
	require.ErrorContains(t, err, "no price CSV header")
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
