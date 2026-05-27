package server

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/fx"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
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
