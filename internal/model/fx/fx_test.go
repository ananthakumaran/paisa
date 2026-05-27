package fx

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// TestConvertToBase_SameCurrency: when the source currency matches the base,
// no conversion is applied (rate = 1).
func TestConvertToBase_SameCurrency(t *testing.T) {
	store := NewRateStore()
	amount := decimal.NewFromInt(1000)
	got, err := store.ConvertToBase(amount, "CNY", "CNY", time.Now())
	assert.NoError(t, err)
	assert.True(t, got.Equal(amount))
}

// TestConvertToBase_DirectRate: a single direct rate in the store is used to
// convert. 100 USD * 7.2 = 720 CNY.
func TestConvertToBase_DirectRate(t *testing.T) {
	d := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	store := NewRateStore()
	store.Put("USD", "CNY", d, decimal.NewFromFloat(7.2))

	got, err := store.ConvertToBase(decimal.NewFromInt(100), "USD", "CNY", d)
	assert.NoError(t, err)
	assert.True(t, got.Equal(decimal.NewFromFloat(720)),
		"expected 720, got %s", got.String())
}

// TestConvertToBase_AsOfBefore: looks up the latest rate as of the given date,
// not requiring an exact day match.
func TestConvertToBase_AsOfBefore(t *testing.T) {
	d1 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	d3 := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	d5 := time.Date(2024, 6, 5, 0, 0, 0, 0, time.UTC)

	store := NewRateStore()
	store.Put("USD", "CNY", d1, decimal.NewFromFloat(7.10))
	store.Put("USD", "CNY", d3, decimal.NewFromFloat(7.20))
	store.Put("USD", "CNY", d5, decimal.NewFromFloat(7.30))

	// As of June 4: should use June 3 rate (most-recent-prior).
	d4 := time.Date(2024, 6, 4, 0, 0, 0, 0, time.UTC)
	got, err := store.ConvertToBase(decimal.NewFromInt(100), "USD", "CNY", d4)
	assert.NoError(t, err)
	assert.True(t, got.Equal(decimal.NewFromFloat(720)),
		"expected 720 (using June 3 rate), got %s", got.String())
}

// TestConvertToBase_InversePivot: if only the reverse pair is known, the
// store should compute base = 1 / inverse. e.g. given CNY->USD = 0.1389,
// converting USD->CNY uses ~7.199.
func TestConvertToBase_InverseRate(t *testing.T) {
	d := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	store := NewRateStore()
	store.Put("CNY", "USD", d, decimal.NewFromFloat(0.1389))

	got, err := store.ConvertToBase(decimal.NewFromInt(100), "USD", "CNY", d)
	assert.NoError(t, err)
	// 100 / 0.1389 = 719.94...
	expected := decimal.NewFromInt(100).Div(decimal.NewFromFloat(0.1389))
	assert.True(t, got.Sub(expected).Abs().LessThan(decimal.NewFromFloat(0.0001)),
		"expected ~%s, got %s", expected.String(), got.String())
}

// TestConvertToBase_USDPivot: HKD->CNY through USD when no direct rate is known.
// HKD->USD = 0.128, USD->CNY = 7.2  => HKD->CNY = 0.9216
func TestConvertToBase_USDPivot(t *testing.T) {
	d := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	store := NewRateStore()
	store.Put("HKD", "USD", d, decimal.NewFromFloat(0.128))
	store.Put("USD", "CNY", d, decimal.NewFromFloat(7.2))

	got, err := store.ConvertToBase(decimal.NewFromInt(1000), "HKD", "CNY", d)
	assert.NoError(t, err)
	// 1000 * 0.128 * 7.2 = 921.6
	assert.True(t, got.Sub(decimal.NewFromFloat(921.6)).Abs().LessThan(decimal.NewFromFloat(0.001)),
		"expected ~921.6, got %s", got.String())
}

// TestConvertToBase_Missing: returns an error if no path can be resolved.
func TestConvertToBase_Missing(t *testing.T) {
	d := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	store := NewRateStore()
	_, err := store.ConvertToBase(decimal.NewFromInt(100), "USD", "CNY", d)
	assert.Error(t, err)
}

// TestIsKnownCurrency_Shape: accepts 3 uppercase letters and rejects everything
// else (tickers, lowercase, empty, fund codes).
func TestIsKnownCurrency_Shape(t *testing.T) {
	cases := map[string]bool{
		"USD":      true,
		"CNY":      true,
		"HKD":      true,
		"EUR":      true,
		"INR":      true,
		"":         false,
		"usd":      false,
		"AAPL":     false,
		"600000":   false,
		"AB":       false,
		"US D":     false,
		"BTC-USDT": false,
	}
	for in, want := range cases {
		assert.Equal(t, want, IsKnownCurrency(in), "IsKnownCurrency(%q)", in)
	}
}

// TestPut_DedupSameDay: re-Put on the same (from,to,date) replaces the value
// in place rather than appending. Without this guarantee the series would
// grow unbounded across `paisa update` invocations because the store is
// process-cached (see Store()).
func TestPut_DedupSameDay(t *testing.T) {
	d := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	store := NewRateStore()
	store.Put("USD", "CNY", d, decimal.NewFromFloat(7.10))
	store.Put("USD", "CNY", d, decimal.NewFromFloat(7.20))
	store.Put("USD", "CNY", d.Add(2*time.Hour), decimal.NewFromFloat(7.30))

	// Three Puts, but only one logical day -> series length should be 1 and
	// hold the most recent value.
	got, ok := store.Lookup("USD", "CNY", d)
	assert.True(t, ok)
	assert.True(t, got.Equal(decimal.NewFromFloat(7.30)),
		"expected last value 7.30, got %s", got.String())
}

// TestStaleLookup: when asOf < earliest known date, we extrapolate backwards
// and signal stale=true so callers can warn but not break.
func TestStaleLookup(t *testing.T) {
	d1 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	d0 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	store := NewRateStore()
	store.Put("USD", "CNY", d1, decimal.NewFromFloat(7.20))

	got, ok, stale := store.directLookupWithStale("USD", "CNY", d0)
	assert.True(t, ok)
	assert.True(t, stale, "should mark stale when asOf precedes first datapoint")
	assert.True(t, got.Equal(decimal.NewFromFloat(7.20)))
}
