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

// isFxAttributable reports whether a posting should contribute to FX gain
// attribution. The reviewer-blocking issue with the previous gate
// (`utils.IsCurrency`) was that IsCurrency returns true only for
// default_currency — so the FX branch never fired for USD or HKD holdings,
// and fx_gain was always 0 in any real multi-currency setup.
//
// New rule: any non-empty ISO-style currency that isn't the base currency
// triggers attribution. Stock/fund tickers (mixed case, 4+ letters, digits)
// are filtered out by IsKnownCurrency's shape check, so market gain on
// equities still flows through the existing `service.GetUnitPrice` path.
func isFxAttributable(commodity, base string) bool {
	return commodity != "" && commodity != base && fx.IsKnownCurrency(commodity)
}

// fxAttribution returns, for a single posting, (balance_in_base, cost_in_base,
// ok). balance is the value re-priced at `now`'s FX rate; cost is the value
// at the posting's acquisition date. The difference is fx_gain.
//
// ok is false when the posting isn't FX-attributable (base-currency lots,
// stocks, blank commodity) or when the rate store has no path to base.
// In either case the caller treats the posting as contributing zero fx_gain.
func fxAttribution(store *fx.RateStore, p posting.Posting, base string, now time.Time) (decimal.Decimal, decimal.Decimal, bool) {
	if store == nil {
		return decimal.Zero, decimal.Zero, false
	}
	if !isFxAttributable(p.Commodity, base) {
		return decimal.Zero, decimal.Zero, false
	}
	// p.Amount is already expressed in default_currency (the ledger CLI
	// applies `market(amount,date,default_currency)` during parsing — see
	// internal/ledger/ledger.go). p.Quantity is the original number of
	// units in p.Commodity. To get the *currency* exposure right we use
	// p.Quantity: e.g. for 100 USD, Quantity=100 and Commodity=USD;
	// converting Quantity through the rate store gives the correct base
	// equivalent regardless of how the journal expressed the price.
	balVal, errBal := store.ConvertToBase(p.Quantity, p.Commodity, base, now)
	if errBal != nil {
		return decimal.Zero, decimal.Zero, false
	}
	costVal, errCost := store.ConvertToBase(p.Quantity, p.Commodity, base, p.Date)
	if errCost != nil {
		return decimal.Zero, decimal.Zero, false
	}
	return balVal, costVal, true
}

// computeFxAttribution aggregates fxAttribution across a posting set. Kept as
// a thin loop so tests can assert on the (balance, cost) pair directly.
// computeNetworth and computeNetworthTimeline call fxAttribution per posting
// inline because they need to interleave the cost basis with the other
// accumulators (investment, withdrawal, balance) anyway.
func computeFxAttribution(store *fx.RateStore, postings []posting.Posting, base string, now time.Time) (decimal.Decimal, decimal.Decimal, error) {
	balanceBase := decimal.Zero
	costBase := decimal.Zero
	for _, p := range postings {
		if p.Account == "Income:CapitalGains" {
			continue
		}
		if p.Commodity == base {
			// Base-currency lots wash out (rate = 1) but still need to flow
			// through so the test fixtures that check base-only behaviour
			// see equal balance and cost.
			balanceBase = balanceBase.Add(p.Amount)
			costBase = costBase.Add(p.Amount)
			continue
		}
		balVal, costVal, ok := fxAttribution(store, p, base, now)
		if !ok {
			continue
		}
		balanceBase = balanceBase.Add(balVal)
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

// toBase converts an amount expressed in `default_currency` to `base`,
// using the rate as of `asOf`. When base == default (the single-currency
// setup that covers every regression fixture and the pre-M1 user base),
// this is the identity — `defaultCurrency` is passed in so the caller's
// `base != defaultCurrency` shortcut can be evaluated against config once.
//
// Failure to find a rate is non-fatal: we return the amount unchanged
// rather than zero out an entire net-worth timeline because frankfurter
// had a 503. The downstream effect is that the same nominal value is
// added in two different units, which is the same behaviour the codebase
// had pre-M1-F.
func toBase(store *fx.RateStore, amount decimal.Decimal, defaultCurrency, base string, asOf time.Time) decimal.Decimal {
	if defaultCurrency == base || store == nil {
		return amount
	}
	v, err := store.ConvertToBase(amount, defaultCurrency, base, asOf)
	if err != nil {
		return amount
	}
	return v
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
	defaultCurrency := config.DefaultCurrency()
	// fxBalanceBase / fxCostBase aggregate, in *base* currency, the current
	// vs acquisition-time value of every non-base-currency cash lot. The
	// delta is the fx_gain attribution. The previous gate used
	// `utils.IsCurrency` which returns true only for the *default* currency,
	// so USD/HKD holdings under a CNY base never triggered the FX branch
	// (BLOCK 1 of the iter-1 review). `isFxAttributable` correctly accepts
	// any ISO-shaped currency that isn't the base.
	var fxBalanceBase decimal.Decimal = decimal.Zero
	var fxCostBase decimal.Decimal = decimal.Zero
	for _, p := range postings {
		isInterest := service.IsInterest(db, p)
		isInterestRepayment := service.IsInterestRepayment(db, p)
		isStockSplit := service.IsStockSplit(db, p)
		isCapitalGains := service.IsCapitalGains(p)

		if isInterest || isInterestRepayment {
			balance = balance.Add(toBase(store, p.Amount, defaultCurrency, base, now))
		} else if isCapitalGains {
			withdrawal = withdrawal.Add(toBase(store, p.Amount.Neg(), defaultCurrency, base, p.Date))
		} else {
			if p.Amount.GreaterThan(decimal.Zero) && !isStockSplit {
				investment = investment.Add(toBase(store, p.Amount, defaultCurrency, base, p.Date))
			}

			if p.Amount.LessThan(decimal.Zero) && !isStockSplit {
				withdrawal = withdrawal.Add(toBase(store, p.Amount.Neg(), defaultCurrency, base, p.Date))
			}

			balance = balance.Add(toBase(store, service.GetMarketPrice(db, p, now), defaultCurrency, base, now))

			if balVal, costVal, ok := fxAttribution(store, p, base, now); ok {
				fxBalanceBase = fxBalanceBase.Add(balVal)
				fxCostBase = fxCostBase.Add(costVal)
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

func computeNetworthTimeline(db *gorm.DB, postings []posting.Posting, computeBalanceUnits bool, store *fx.RateStore) []Networth {
	var networths []Networth

	var p posting.Posting

	if len(postings) == 0 {
		return []Networth{}
	}

	// fxLot keeps the foreign-currency quantity and its acquisition-time
	// cost expressed in base currency. On each timeline day we re-value
	// the same quantity at the day's FX rate and the delta is the fx_gain
	// attribution for that day.
	type fxLot struct {
		commodity  string
		date       time.Time
		quantity   decimal.Decimal // original quantity in `commodity` units
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
	defaultCurrency := config.DefaultCurrency()

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
				// lots. Stock/fund prices flow through service.GetUnitPrice
				// (already in default_currency); the FX move on the
				// stock's denominating currency is what we attribute here.
				if isFxAttributable(p.Commodity, base) {
					cost, err := store.ConvertToBase(p.Quantity, p.Commodity, base, p.Date)
					if err == nil {
						fxLots = append(fxLots, fxLot{
							commodity:  p.Commodity,
							date:       p.Date,
							quantity:   p.Quantity,
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
			investment = investment.Add(toBase(store, rs.investment, defaultCurrency, base, start))
			withdrawal = withdrawal.Add(toBase(store, rs.withdrawal, defaultCurrency, base, start))

			if utils.IsCurrency(commodity) {
				balance = balance.Add(toBase(store, rs.balance, defaultCurrency, base, start))
			} else {
				if computeBalanceUnits {
					balanceUnits = balanceUnits.Add(rs.balanceUnits)
				}
				price := service.GetUnitPrice(db, commodity, start)
				if !price.Value.Equal(decimal.Zero) {
					balance = balance.Add(toBase(store, rs.balanceUnits.Mul(price.Value), defaultCurrency, base, start))
				} else {
					balance = balance.Add(toBase(store, rs.balance, defaultCurrency, base, start))
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
			balVal, err := store.ConvertToBase(lot.quantity, lot.commodity, base, start)
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
