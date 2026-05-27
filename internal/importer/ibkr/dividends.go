package ibkr

import (
	"regexp"
	"strings"

	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/shopspring/decimal"
)

// dividendSymbolRe extracts the ticker from the Description column. IBKR
// formats dividend descriptions as "TICKER(ISIN) Cash Dividend USD 0.24
// per Share" (or similar). The ticker is always the leading run of
// alphanumeric / dot characters before the opening parenthesis.
var dividendSymbolRe = regexp.MustCompile(`^([A-Z0-9\.]+)`)

// parseDividends handles cash dividends received. Money enters the source
// (cash) account, so per ParsedTxn convention the amount is NEGATIVE. The
// counterpart account is `Income:Dividend:<Symbol>` — the per-symbol
// hierarchy lets users group dividend income by holding in reports.
func parseDividends(rows [][]string) []importer.ParsedTxn {
	if len(rows) < 2 {
		return nil
	}
	idx := columnIndex(rows[0])

	var out []importer.ParsedTxn
	for _, row := range rows[1:] {
		currency := at(row, idx, "Currency")
		// IBKR's Dividends section ends with a "Total" summary row
		// that has an empty Date — skip anything that doesn't look
		// like a real dividend.
		dateStr := at(row, idx, "Date")
		desc := at(row, idx, "Description")
		amountStr := at(row, idx, "Amount")
		if dateStr == "" || desc == "" || amountStr == "" {
			continue
		}

		date, err := parseIBKRDateTime(dateStr)
		if err != nil {
			continue
		}
		amount, err := decimal.NewFromString(stripThousands(amountStr))
		if err != nil {
			continue
		}

		// IBKR's Amount column for dividends is positive (a credit
		// to the user). Our convention wants it negative ("money
		// entering the source").
		amount = amount.Neg()

		symbol := extractDividendSymbol(desc)
		suggested := "Income:Dividend"
		if symbol != "" {
			suggested = "Income:Dividend:" + symbol
		}

		out = append(out, importer.ParsedTxn{
			Date:             date,
			Payee:            "Dividend " + symbol,
			Note:             desc,
			Amount:           amount,
			Currency:         currency,
			SuggestedAccount: suggested,
			RawText:          strings.Join(row, ","),
		})
	}
	return out
}

// parseWithholdingTax handles US (or other foreign) withholding tax on
// dividends. Money LEAVES the source account, so amount is POSITIVE (the
// CSV stores it negative; we flip the sign). The counterpart is a single
// foreign-withholding expense bucket — users can split per-country if they
// care.
func parseWithholdingTax(rows [][]string) []importer.ParsedTxn {
	if len(rows) < 2 {
		return nil
	}
	idx := columnIndex(rows[0])

	var out []importer.ParsedTxn
	for _, row := range rows[1:] {
		currency := at(row, idx, "Currency")
		dateStr := at(row, idx, "Date")
		desc := at(row, idx, "Description")
		amountStr := at(row, idx, "Amount")
		if dateStr == "" || desc == "" || amountStr == "" {
			continue
		}

		date, err := parseIBKRDateTime(dateStr)
		if err != nil {
			continue
		}
		amount, err := decimal.NewFromString(stripThousands(amountStr))
		if err != nil {
			continue
		}

		// IBKR's Amount column for WHT is negative (debit to user).
		// Flip to positive ("money leaving source").
		amount = amount.Neg()

		out = append(out, importer.ParsedTxn{
			Date:             date,
			Payee:            "Withholding Tax",
			Note:             desc,
			Amount:           amount,
			Currency:         currency,
			SuggestedAccount: "Expenses:Tax:Foreign:Withholding",
			RawText:          strings.Join(row, ","),
		})
	}
	return out
}

func extractDividendSymbol(desc string) string {
	m := dividendSymbolRe.FindStringSubmatch(desc)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}
