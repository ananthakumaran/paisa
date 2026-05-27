package ibkr

import (
	"strings"

	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/shopspring/decimal"
)

// parseFees handles the Fees section. IBKR emits one row per fee event
// (statement fee, market-data subscription, exchange fee, …). All are
// debits to the user, so amounts in the CSV are negative; we flip the
// sign per ParsedTxn convention (money LEAVING source = POSITIVE).
//
// The Fees section has an extra "Subtitle" column that groups fees by
// type (e.g. "Other Fees", "Trading Permission"). We surface it in the
// Note so the preview UI can show context, but we don't use it to pick
// the account — `Expenses:Brokerage:IBKR:Fee` is a single bucket users
// can refine in the preview.
func parseFees(rows [][]string) []importer.ParsedTxn {
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
		subtitle := at(row, idx, "Subtitle")
		if dateStr == "" || amountStr == "" {
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
		// CSV amount is negative (debit). Flip to "money leaving"
		// positive convention.
		amount = amount.Neg()

		note := desc
		if subtitle != "" {
			note = subtitle + ": " + desc
		}

		out = append(out, importer.ParsedTxn{
			Date:             date,
			Payee:            "IBKR Fee",
			Note:             note,
			Amount:           amount,
			Currency:         currency,
			SuggestedAccount: "Expenses:Brokerage:IBKR:Fee",
			RawText:          strings.Join(row, ","),
		})
	}
	return out
}

// parseInterest handles the Interest section: credit interest IBKR pays
// on idle cash balances and (rarely) debit interest on borrowed cash. We
// flip the sign so credits show up as NEGATIVE amounts (money entering
// source). Debit interest from IBKR is already negative in the CSV, so
// after the flip it becomes positive — exactly the convention an expense
// would need.
func parseInterest(rows [][]string) []importer.ParsedTxn {
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
		if dateStr == "" || amountStr == "" {
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
		// IBKR's Amount: positive = credit interest (income).
		// Flip so income shows up as negative per our convention.
		amount = amount.Neg()

		// Credit (income) vs debit (expense) determines the
		// counterpart. After the flip:
		//   - negative amount = money in = income
		//   - positive amount = money out = expense (rare; only for
		//     borrowed-cash interest)
		suggested := "Income:Brokerage:IBKR:Interest"
		payee := "IBKR Interest"
		if amount.IsPositive() {
			suggested = "Expenses:Brokerage:IBKR:Interest"
		}

		out = append(out, importer.ParsedTxn{
			Date:             date,
			Payee:            payee,
			Note:             desc,
			Amount:           amount,
			Currency:         currency,
			SuggestedAccount: suggested,
			RawText:          strings.Join(row, ","),
		})
	}
	return out
}
