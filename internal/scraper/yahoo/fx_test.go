package yahoo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFxSymbol returns the yahoo-style ticker for a base->target pair.
func TestFxSymbol(t *testing.T) {
	assert.Equal(t, "USDCNY=X", FxSymbol("USD", "CNY"))
	assert.Equal(t, "HKDCNY=X", FxSymbol("HKD", "CNY"))
	assert.Equal(t, "EURCNY=X", FxSymbol("EUR", "CNY"))
}

// TestFxProviderCode confirms the FX provider registers under the expected key.
func TestFxProviderCode(t *testing.T) {
	p := &FxPriceProvider{}
	assert.Equal(t, "yahoo-fx", p.Code())
}

// TestParseChartResponse exercises the lightweight chart-response parser used
// for FX symbols. The Yahoo /v8/finance/chart payload is a deep nested struct;
// the parser is shared with the stock scraper but kept independently testable
// here so we don't reach across packages in tests.
func TestParseChartResponse(t *testing.T) {
	body := []byte(`{
		"chart":{
			"result":[{
				"meta":{"currency":"CNY"},
				"timestamp":[1717200000,1717286400],
				"indicators":{"quote":[{"close":[7.20,7.21]}]}
			}]
		}
	}`)
	prices, err := parseFxChartResponse(body, "USD", "CNY")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(prices))
	assert.Equal(t, "7.2", prices[0].Value.String())
	assert.Equal(t, "7.21", prices[1].Value.String())
	assert.Equal(t, "USDCNY", prices[0].CommodityID)
}
