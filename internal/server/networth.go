package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/fx"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Networth is the per-day net-worth snapshot returned to the UI.
//
// GainAmount is the total gain (= MarketGainAmount + FxGainAmount). The
// market/fx split is exposed for multi-currency portfolios where holdings in
// foreign currencies (e.g. USD or HKD positions valued in a CNY base) need
// the FX move attributed separately from the underlying market move. Older
// clients can keep reading GainAmount; new clients can show the split.
type Networth struct {
	Date                time.Time       `json:"date"`
	InvestmentAmount    decimal.Decimal `json:"investmentAmount"`
	WithdrawalAmount    decimal.Decimal `json:"withdrawalAmount"`
	GainAmount          decimal.Decimal `json:"gainAmount"`
	MarketGainAmount    decimal.Decimal `json:"marketGainAmount"`
	FxGainAmount        decimal.Decimal `json:"fxGainAmount"`
	BalanceAmount       decimal.Decimal `json:"balanceAmount"`
	BalanceUnits        decimal.Decimal `json:"balanceUnits"`
	NetInvestmentAmount decimal.Decimal `json:"netInvestmentAmount"`
}

// computeFxAttribution returns, for the given postings, the current balance
// in the base currency (using "now" FX rates) and the historical cost basis
// in the base currency (using each posting's acquisition-date rate).
//
// The difference is fx_gain — the slice of total gain attributable purely to
// exchange-rate movement between acquisition and today. Postings already in
// the base currency contribute identically to both sums and so wash out.
func computeFxAttribution(store *fx.RateStore, postings []posting.Posting, base string, now time.Time) (decimal.Decimal, decimal.Decimal, error) {
	balanceBase := decimal.Zero
	costBase := decimal.Zero
	for _, p := range postings {
		// Skip ledger-internal capital gains rows; they're accounted for in
		// the withdrawal series, not as currency exposure.
		if p.Account == "Income:CapitalGains" {
			continue
		}
		// Skip currencies we can't price; market_gain falls back to amount
		// in those cases and fx_gain stays neutral.
		if p.Commodity == "" {
			continue
		}
		amount := p.Amount
		if p.Commodity == base {
			balanceBase = balanceBase.Add(amount)
			costBase = costBase.Add(amount)
			continue
		}
		// Re-value at today's rate (balance) and at the posting's date
		// (cost). Failure to find a rate is non-fatal — we fall back to
		// the raw ledger amount so we don't blow up an entire timeline
		// because one rare currency is missing.
		balanceVal, err := store.ConvertToBase(amount, p.Commodity, base, now)
		if err != nil {
			balanceVal = amount
		}
		costVal, err := store.ConvertToBase(amount, p.Commodity, base, p.Date)
		if err != nil {
			costVal = amount
		}
		balanceBase = balanceBase.Add(balanceVal)
		costBase = costBase.Add(costVal)
	}
	return balanceBase, costBase, nil
}

// FxStore exposes the process-wide FX rate store.
func FxStore() *fx.RateStore {
	return fx.Store()
}

// loadFxRatesFromDB seeds the store from the FX prices that were upserted
// into the prices table by the scrapers. Each price row's CommodityID is a
// 6-letter pair like "USDCNY".
func loadFxRatesFromDB(db *gorm.DB, store *fx.RateStore) {
	type fxRow struct {
		Date          time.Time
		CommodityID   string
		Value         decimal.Decimal
		CommodityName string
	}
	var rows []fxRow
	// CommodityType for FX is config.Unknown (set by both scrapers); we
	// further filter by the 6-letter pair shape via LENGTH.
	db.Table("prices").
		Select("date, commodity_id, value, commodity_name").
		Where("commodity_type = ? AND LENGTH(commodity_id) = 6", config.Unknown).
		Scan(&rows)
	for _, r := range rows {
		if len(r.CommodityID) != 6 {
			continue
		}
		from, to := r.CommodityID[:3], r.CommodityID[3:]
		store.Put(from, to, r.Date, r.Value)
	}
}

func GetNetworth(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%", "Income:CapitalGains:%", "Liabilities:%").UntilToday().All()

	postings = service.PopulateMarketPrice(db, postings)
	store := FxStore()
	loadFxRatesFromDB(db, store)
	networthTimeline := computeNetworthTimeline(db, postings, false, store)
	xirr := service.XIRR(db, postings)
	return gin.H{"networthTimeline": networthTimeline, "xirr": xirr}
}

func GetCurrentNetworth(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%", "Income:CapitalGains:%", "Liabilities:%").UntilToday().All()
	postings = service.PopulateMarketPrice(db, postings)
	store := FxStore()
	loadFxRatesFromDB(db, store)
	networth := computeNetworth(db, postings, store)
	xirr := service.XIRR(db, postings)
	return gin.H{"networth": networth, "xirr": xirr}
}

func computeNetworth(db *gorm.DB, postings []posting.Posting, store *fx.RateStore) Networth {
	var networth Networth

	if len(postings) == 0 {
		return networth
	}

	var investment decimal.Decimal = decimal.Zero
	var withdrawal decimal.Decimal = decimal.Zero
	var balance decimal.Decimal = decimal.Zero

	now := utils.EndOfToday()
	base := config.BaseCurrency()
	// fxBalanceBase / fxCostBase track each non-base-currency-denominated
	// asset's current vs acquisition-time value in the base currency. The
	// delta is the fx_gain attribution.
	var fxBalanceBase decimal.Decimal = decimal.Zero
	var fxCostBase decimal.Decimal = decimal.Zero
	for _, p := range postings {
		isInterest := service.IsInterest(db, p)
		isInterestRepayment := service.IsInterestRepayment(db, p)
		isStockSplit := service.IsStockSplit(db, p)
		isCapitalGains := service.IsCapitalGains(p)

		if isInterest || isInterestRepayment {
			balance = balance.Add(p.Amount)
		} else if isCapitalGains {
			withdrawal = withdrawal.Add(p.Amount.Neg())
		} else {
			if p.Amount.GreaterThan(decimal.Zero) && !isStockSplit {
				investment = investment.Add(p.Amount)
			}

			if p.Amount.LessThan(decimal.Zero) && !isStockSplit {
				withdrawal = withdrawal.Add(p.Amount.Neg())
			}

			balance = balance.Add(service.GetMarketPrice(db, p, now))

			// Compute the FX attribution slice — only meaningful when
			// the posting is in a non-base currency. For commodities
			// (stocks/funds), p.Commodity is the ticker, which won't
			// match either the base or any currency, so we skip it;
			// market gain in that case stays fully under market_gain.
			if utils.IsCurrency(p.Commodity) || p.Commodity == base {
				if balVal, costVal, ok := fxRevalue(store, p, base, now); ok {
					fxBalanceBase = fxBalanceBase.Add(balVal)
					fxCostBase = fxCostBase.Add(costVal)
				}
			}
		}
	}

	gain := balance.Add(withdrawal).Sub(investment)
	fxGain := fxBalanceBase.Sub(fxCostBase)
	marketGain := gain.Sub(fxGain)
	netInvestment := investment.Sub(withdrawal)
	networth = Networth{
		Date:                now,
		InvestmentAmount:    investment,
		WithdrawalAmount:    withdrawal,
		GainAmount:          gain,
		MarketGainAmount:    marketGain,
		FxGainAmount:        fxGain,
		BalanceAmount:       balance,
		NetInvestmentAmount: netInvestment,
	}

	return networth
}

// fxRevalue returns (balance_in_base, cost_in_base, ok) for a single posting.
// ok is false if the posting is in the base currency (no FX attribution) or
// if the store has no rate.
func fxRevalue(store *fx.RateStore, p posting.Posting, base string, now time.Time) (decimal.Decimal, decimal.Decimal, bool) {
	if store == nil || p.Commodity == base {
		return decimal.Zero, decimal.Zero, false
	}
	balVal, errBal := store.ConvertToBase(p.Amount, p.Commodity, base, now)
	if errBal != nil {
		return decimal.Zero, decimal.Zero, false
	}
	costVal, errCost := store.ConvertToBase(p.Amount, p.Commodity, base, p.Date)
	if errCost != nil {
		return decimal.Zero, decimal.Zero, false
	}
	return balVal, costVal, true
}

func computeNetworthTimeline(db *gorm.DB, postings []posting.Posting, computeBalanceUnits bool, store *fx.RateStore) []Networth {
	var networths []Networth

	var p posting.Posting

	if len(postings) == 0 {
		return []Networth{}
	}

	// fxLot keeps the cost-basis in base currency (computed at acquisition
	// date) so we can compare it against the current re-valuation on each
	// timeline day.
	type fxLot struct {
		commodity string
		date      time.Time
		amount    decimal.Decimal
		// costInBase is the lot's value in base currency at acquisition
		// time. nil-ish (zero amount with zero cost) is treated as "no
		// FX attribution available", so the lot contributes nothing.
		costInBase decimal.Decimal
		hasCost    bool
	}
	type RunningSum struct {
		investment   decimal.Decimal
		withdrawal   decimal.Decimal
		balance      decimal.Decimal
		balanceUnits decimal.Decimal
	}

	accumulator := make(map[string]RunningSum)
	var fxLots []fxLot
	base := config.BaseCurrency()

	end := utils.EndOfToday()
	for start := postings[0].Date; start.Before(end); start = start.AddDate(0, 0, 1) {
		for len(postings) > 0 && (postings[0].Date.Before(start) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			rs := accumulator[p.Commodity]

			isInterest := service.IsInterest(db, p)
			isCapitalGains := service.IsCapitalGains(p)

			if p.Amount.GreaterThan(decimal.Zero) && !isInterest {
				rs.investment = rs.investment.Add(p.Amount)
			}

			if p.Amount.LessThan(decimal.Zero) && !isInterest {
				rs.withdrawal = rs.withdrawal.Add(p.Amount.Neg())
			}

			if !isCapitalGains {
				rs.balance = rs.balance.Add(service.GetMarketPrice(db, p, start))
				rs.balanceUnits = rs.balanceUnits.Add(p.Quantity)

				// Snapshot the cost-in-base for foreign-currency cash
				// lots. Commodity-side fx is rolled into the underlying
				// stock's market price (Yahoo's stock scraper already
				// converts to default currency), so we only track
				// currency-denominated lots here.
				if utils.IsCurrency(p.Commodity) && p.Commodity != base {
					cost, err := store.ConvertToBase(p.Amount, p.Commodity, base, p.Date)
					if err == nil {
						fxLots = append(fxLots, fxLot{
							commodity:  p.Commodity,
							date:       p.Date,
							amount:     p.Amount,
							costInBase: cost,
							hasCost:    true,
						})
					}
				}
			}

			accumulator[p.Commodity] = rs

		}

		var investment decimal.Decimal = decimal.Zero
		var withdrawal decimal.Decimal = decimal.Zero
		var balance decimal.Decimal = decimal.Zero
		var balanceUnits decimal.Decimal = decimal.Zero

		for commodity, rs := range accumulator {
			investment = investment.Add(rs.investment)
			withdrawal = withdrawal.Add(rs.withdrawal)

			if utils.IsCurrency(commodity) {
				balance = balance.Add(rs.balance)
			} else {
				if computeBalanceUnits {
					balanceUnits = balanceUnits.Add(rs.balanceUnits)
				}
				price := service.GetUnitPrice(db, commodity, start)
				if !price.Value.Equal(decimal.Zero) {
					balance = balance.Add(rs.balanceUnits.Mul(price.Value))
				} else {
					balance = balance.Add(rs.balance)
				}
			}

		}

		// Compute FX attribution for the day's timeline: revalue every
		// foreign-currency lot at `start`'s rate vs its acquisition rate.
		fxBalance := decimal.Zero
		fxCost := decimal.Zero
		for _, lot := range fxLots {
			if !lot.hasCost {
				continue
			}
			balVal, err := store.ConvertToBase(lot.amount, lot.commodity, base, start)
			if err != nil {
				continue
			}
			fxBalance = fxBalance.Add(balVal)
			fxCost = fxCost.Add(lot.costInBase)
		}
		gain := balance.Add(withdrawal).Sub(investment)
		fxGain := fxBalance.Sub(fxCost)
		marketGain := gain.Sub(fxGain)
		netInvestment := investment.Sub(withdrawal)
		networths = append(networths, Networth{
			Date:                start,
			InvestmentAmount:    investment,
			WithdrawalAmount:    withdrawal,
			GainAmount:          gain,
			MarketGainAmount:    marketGain,
			FxGainAmount:        fxGain,
			BalanceAmount:       balance,
			BalanceUnits:        balanceUnits,
			NetInvestmentAmount: netInvestment,
		})

		if len(postings) == 0 && balance.Abs().LessThan(decimal.NewFromFloat(0.01)) {
			break
		}
	}
	return networths
}
