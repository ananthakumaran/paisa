package ibkr

import (
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/importer"
	"github.com/shopspring/decimal"
)

// parseTrades turns Trades section rows into ParsedTxn. Sign convention
// per importer.ParsedTxn:
//
//   - Buy  (Quantity > 0): money leaves cash → POSITIVE amount = |Proceeds| + |Comm|.
//     SuggestedAccount is the per-symbol stock asset account so the
//     counterpart leg is the new position.
//   - Sell (Quantity < 0): money enters cash → NEGATIVE amount = -(Proceeds - |Comm|).
//     We do NOT model the cost-basis leg here; the user can split the
//     transaction in the preview UI if they want a fully accounted disposal.
//
// IBKR's Date/Time format is "YYYY-MM-DD;HH:MM:SS" (note the SEMICOLON).
// Date-only granularity is sufficient for ledger entries; the time is
// dropped.
func parseTrades(rows [][]string) []importer.ParsedTxn {
	if len(rows) < 2 {
		return nil
	}
	idx := columnIndex(rows[0])

	var out []importer.ParsedTxn
	for _, row := range rows[1:] {
		// Only "Order" rows are real fills. IBKR also emits "ClosedLot"
		// and "Total" sub-rows in the same section that are summaries —
		// importing them would double-count the trade.
		discriminator := at(row, idx, "DataDiscriminator")
		if discriminator != "" && discriminator != "Order" {
			continue
		}

		symbol := at(row, idx, "Symbol")
		if symbol == "" {
			continue
		}
		currency := at(row, idx, "Currency")
		qtyStr := at(row, idx, "Quantity")
		proceedsStr := at(row, idx, "Proceeds")
		commStr := at(row, idx, "Comm/Fee")
		dateStr := at(row, idx, "Date/Time")

		quantity, err := decimal.NewFromString(stripThousands(qtyStr))
		if err != nil {
			continue
		}
		proceeds, err := decimal.NewFromString(stripThousands(proceedsStr))
		if err != nil {
			continue
		}
		// Commission is optional in some Flex Query layouts.
		commission := decimal.Zero
		if commStr != "" {
			if v, err := decimal.NewFromString(stripThousands(commStr)); err == nil {
				commission = v
			}
		}

		date, err := parseIBKRDateTime(dateStr)
		if err != nil {
			continue
		}

		// IBKR sign on commission is always negative when it's a debit.
		// |proceeds| is the cash gross; |commission| is the broker fee.
		// For a buy: cash out = |proceeds| + |commission|.
		// For a sell: cash in = proceeds - |commission| (proceeds already
		// positive on a sell), and per ParsedTxn convention "money in" is
		// expressed as a negative number.
		var amount decimal.Decimal
		var payee string
		if quantity.IsPositive() {
			amount = proceeds.Abs().Add(commission.Abs())
			payee = "Buy " + symbol
		} else {
			net := proceeds.Sub(commission.Abs())
			amount = net.Neg()
			payee = "Sell " + symbol
		}

		out = append(out, importer.ParsedTxn{
			Date:             date,
			Payee:            payee,
			Note:             "IBKR trade " + symbol + " qty=" + qtyStr,
			Amount:           amount,
			Currency:         currency,
			SuggestedAccount: "Assets:Brokerage:IBKR:Stock:" + symbol,
			RawText:          strings.Join(row, ","),
		})
	}
	return out
}

// parseIBKRDateTime handles IBKR's eccentric date formats. The Flex Query
// engine emits "YYYY-MM-DD;HH:MM:SS" for trades (semicolon!), but plain
// "YYYY-MM-DD" for dividends, fees, interest, etc. We accept both so each
// section parser can call this without first reformatting.
func parseIBKRDateTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	// Trade-style: split at the semicolon, keep the date half. We use
	// SplitN rather than Replace so a literal semicolon inside the date
	// can never sneak through.
	if i := strings.Index(s, ";"); i >= 0 {
		s = s[:i]
	}
	// Some exports use a space separator instead of a semicolon — handle
	// that too.
	if i := strings.Index(s, " "); i >= 0 {
		s = s[:i]
	}
	return time.Parse("2006-01-02", s)
}

// stripThousands removes the thousands separators IBKR sometimes inserts
// into large numbers (e.g. `1,250.00`). decimal.NewFromString rejects
// these, so we normalise before parsing.
func stripThousands(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), ",", "")
}
