package server

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/cii"
	"github.com/ananthakumaran/paisa/internal/model/fx"
	"github.com/ananthakumaran/paisa/internal/model/portfolio"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func makeFxPosting(commodity string, amountInCommodity int64, date time.Time) posting.Posting {
	return posting.Posting{
		Date:      date,
		Account:   "Assets:Investments:IBKR",
		Commodity: commodity,
		Quantity:  decimal.NewFromInt(amountInCommodity),
		Amount:    decimal.NewFromInt(amountInCommodity),
	}
}

// TestComputeFxGain_NoMovement: when the FX rate is unchanged between
// acquisition and "now", fx_gain = 0 and the entire gain is market_gain.
func TestComputeFxGain_NoMovement(t *testing.T) {
	acqDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	store := fx.NewRateStore()
	store.Put("USD", "CNY", acqDate, decimal.NewFromFloat(7.20))
	store.Put("USD", "CNY", now, decimal.NewFromFloat(7.20))

	// 100 USD acquired @ 7.20 = 720 CNY cost basis
	// Today: still 100 USD @ 7.20 = 720 CNY market value
	p := makeFxPosting("USD", 100, acqDate)
	balanceBase, costBase, err := computeFxAttribution(store, []posting.Posting{p}, "CNY", now)
	assert.NoError(t, err)
	assert.True(t, balanceBase.Equal(decimal.NewFromFloat(720)),
		"expected balance 720, got %s", balanceBase.String())
	assert.True(t, costBase.Equal(decimal.NewFromFloat(720)),
		"expected cost 720, got %s", costBase.String())
	fxGain := balanceBase.Sub(costBase)
	assert.True(t, fxGain.IsZero(), "expected fx_gain 0, got %s", fxGain.String())
}

// TestComputeFxGain_RateAppreciation: USD strengthened against CNY, so the
// same 100 USD is worth more CNY today => fx_gain > 0.
func TestComputeFxGain_RateAppreciation(t *testing.T) {
	acqDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	store := fx.NewRateStore()
	store.Put("USD", "CNY", acqDate, decimal.NewFromFloat(7.00))
	store.Put("USD", "CNY", now, decimal.NewFromFloat(7.20))

	p := makeFxPosting("USD", 100, acqDate)
	balanceBase, costBase, err := computeFxAttribution(store, []posting.Posting{p}, "CNY", now)
	assert.NoError(t, err)
	// Balance: 100 * 7.20 = 720
	// Cost basis at acquisition: 100 * 7.00 = 700
	// fx_gain = 720 - 700 = 20
	assert.True(t, balanceBase.Equal(decimal.NewFromFloat(720)),
		"expected balance 720, got %s", balanceBase.String())
	assert.True(t, costBase.Equal(decimal.NewFromFloat(700)),
		"expected cost 700, got %s", costBase.String())
	fxGain := balanceBase.Sub(costBase)
	assert.True(t, fxGain.Equal(decimal.NewFromInt(20)),
		"expected fx_gain 20, got %s", fxGain.String())
}

// TestComputeFxGain_BaseCurrencyOnly: postings in the base currency contribute
// no fx_gain (rate is always 1).
func TestComputeFxGain_BaseCurrencyOnly(t *testing.T) {
	acqDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	store := fx.NewRateStore()

	p := makeFxPosting("CNY", 1000, acqDate)
	balanceBase, costBase, err := computeFxAttribution(store, []posting.Posting{p}, "CNY", now)
	assert.NoError(t, err)
	assert.True(t, balanceBase.Equal(decimal.NewFromInt(1000)))
	assert.True(t, costBase.Equal(decimal.NewFromInt(1000)))
}

// inMemoryDB returns a gorm.DB backed by an in-memory SQLite, with all the
// tables computeNetworth's transitive dependencies touch already migrated.
func inMemoryDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&posting.Posting{}))
	assert.NoError(t, db.AutoMigrate(&price.Price{}))
	assert.NoError(t, db.AutoMigrate(&portfolio.Portfolio{}))
	assert.NoError(t, db.AutoMigrate(&cii.CII{}))
	return db
}

// TestComputeNetworth_USDUnderCNYBase exercises the end-to-end production
// path (BLOCK 1 from the iter-1 review): the old code gated FX attribution on
// `utils.IsCurrency(commodity)` which only matched the default_currency,
// so USD/HKD holdings under a CNY base never produced a non-zero fxGain.
//
// Scenario:
//   - base_currency = default_currency = CNY (matches the user's mydata)
//   - one 100 USD checking deposit acquired on 2024-01-01 at USDCNY=7.00
//   - "now" = 2024-06-01 with USDCNY=7.20
//
// Expected: fxGain = 100 * (7.20 - 7.00) = 20 CNY, marketGain ≈ 0.
func TestComputeNetworth_USDUnderCNYBase(t *testing.T) {
	acqDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	// Pin config — BaseCurrency() falls back to DefaultCurrency when unset,
	// which is what mydata/paisa.yaml does too.
	prev := config.SetConfigForTest(config.Config{DefaultCurrency: "CNY", BaseCurrency: "CNY", TimeZone: "UTC", Locale: "en-US"})
	defer config.SetConfigForTest(prev)

	// Pin "now" so utils.EndOfToday() is deterministic.
	utils.SetNow("2024-06-01")
	defer utils.ResetNow()

	store := fx.NewRateStore()
	store.Put("USD", "CNY", acqDate, decimal.NewFromFloat(7.00))
	store.Put("USD", "CNY", now, decimal.NewFromFloat(7.20))

	db := inMemoryDB(t)
	// Seed a non-FX USD price so service.GetUnitPrice resolves. In a real
	// journal these come from the user's `P USD = X CNY` lines or from a
	// ledger price commodity tied to a non-FX provider; for this test we
	// just insert the two relevant dates directly.
	db.Create(&price.Price{Date: acqDate, CommodityType: config.Stock, CommodityID: "USD", CommodityName: "USD", Value: decimal.NewFromFloat(7.00)})
	db.Create(&price.Price{Date: now, CommodityType: config.Stock, CommodityID: "USD", CommodityName: "USD", Value: decimal.NewFromFloat(7.20)})

	// Clear caches so the new in-memory DB is used.
	service.ClearInterestCache()
	service.ClearPriceCache()
	transaction.ClearCache()

	// One USD deposit. p.Amount is in default_currency (CNY) — the ledger
	// CLI applies market() at parse time, see ledger.go's CSV format.
	// p.Quantity is the original USD amount.
	p := posting.Posting{
		Date:          acqDate,
		Account:       "Assets:Checking",
		Commodity:     "USD",
		Quantity:      decimal.NewFromInt(100),
		Amount:        decimal.NewFromInt(700), // 100 USD @ 7.00 CNY = 700 CNY at acquisition
		TransactionID: "tx-1",
	}
	out := computeNetworth(db, []posting.Posting{p}, store)

	// FX gain: 100 * 7.20 - 100 * 7.00 = 20 CNY.
	assert.True(t, out.FxGainAmount.Equal(decimal.NewFromInt(20)),
		"expected fxGain=20, got %s", out.FxGainAmount.String())
	// Sanity: total gain should equal market + fx.
	sum := out.MarketGainAmount.Add(out.FxGainAmount)
	assert.True(t, out.GainAmount.Equal(sum),
		"gain (%s) != market (%s) + fx (%s)",
		out.GainAmount, out.MarketGainAmount, out.FxGainAmount)
	assert.False(t, out.FxGainAmount.IsZero(),
		"BLOCK 1 regression: fxGain must be non-zero for USD holding under CNY base")
}

// TestComputeNetworth_DefaultDifferentFromBase exercises BLOCK 2 of the iter-1
// review: when default_currency != base_currency the old code mixed
// default-currency `balance` with base-currency `fxGain` in the subtraction
// `marketGain := gain.Sub(fxGain)`. The new code uses toBase() to keep
// every accumulator in base currency.
//
// Scenario: default_currency=INR (ledger output in INR), base_currency=CNY.
// One INR-denominated deposit. With INR->CNY = 0.084 today, the 100,000 INR
// balance translates to 8,400 CNY. fxGain should be 0 (no FX move on the
// base-side currency), and marketGain should be 0 (no market move).
func TestComputeNetworth_DefaultDifferentFromBase(t *testing.T) {
	acqDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	prev := config.SetConfigForTest(config.Config{DefaultCurrency: "INR", BaseCurrency: "CNY", TimeZone: "UTC", Locale: "en-US"})
	defer config.SetConfigForTest(prev)

	utils.SetNow("2024-06-01")
	defer utils.ResetNow()
	now := utils.EndOfToday()

	store := fx.NewRateStore()
	store.Put("INR", "CNY", acqDate, decimal.NewFromFloat(0.084))
	store.Put("INR", "CNY", now, decimal.NewFromFloat(0.084))

	db := inMemoryDB(t)
	service.ClearInterestCache()
	service.ClearPriceCache()
	transaction.ClearCache()

	p := posting.Posting{
		Date:          acqDate,
		Account:       "Assets:Checking",
		Commodity:     "INR",
		Quantity:      decimal.NewFromInt(100000),
		Amount:        decimal.NewFromInt(100000), // already in default_currency (INR)
		TransactionID: "tx-1",
	}
	out := computeNetworth(db, []posting.Posting{p}, store)

	// Balance: 100000 INR -> 100000 * 0.084 = 8400 CNY.
	assert.True(t, out.BalanceAmount.Equal(decimal.NewFromInt(8400)),
		"expected balance=8400 CNY, got %s", out.BalanceAmount.String())
	// Investment: also converted (INR->CNY at acq date).
	assert.True(t, out.InvestmentAmount.Equal(decimal.NewFromInt(8400)),
		"expected investment=8400 CNY, got %s", out.InvestmentAmount.String())
	// Both balance and investment in CNY => gain = 0.
	assert.True(t, out.GainAmount.IsZero(),
		"expected gain=0, got %s", out.GainAmount.String())
	// No FX move (rate constant) => fxGain = 0.
	assert.True(t, out.FxGainAmount.IsZero(),
		"expected fxGain=0, got %s", out.FxGainAmount.String())
	// marketGain = gain - fxGain = 0.
	assert.True(t, out.MarketGainAmount.IsZero(),
		"expected marketGain=0, got %s", out.MarketGainAmount.String())
}
