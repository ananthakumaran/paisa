package boc

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// TestParseLatest verifies parsing of the frankfurter.app /latest payload
// for a single base->target conversion.
func TestParseLatest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/latest", r.URL.Path)
		assert.Equal(t, "USD", r.URL.Query().Get("from"))
		assert.Equal(t, "CNY", r.URL.Query().Get("to"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"amount":1.0,
			"base":"USD",
			"date":"2024-06-03",
			"rates":{"CNY":7.2345}
		}`))
	}))
	defer srv.Close()

	rate, date, err := fetchLatestFrom(srv.URL, "USD", "CNY")
	assert.NoError(t, err)
	assert.True(t, rate.Equal(decimal.NewFromFloat(7.2345)),
		"expected 7.2345, got %s", rate.String())
	assert.Equal(t, "2024-06-03", date.Format("2006-01-02"))
}

// TestParseHistorical verifies parsing of the frankfurter.app date series
// (used for historical FX rate lookup since a fixed start).
func TestParseHistorical(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/2024-06-01..", r.URL.Path)
		assert.Equal(t, "USD", r.URL.Query().Get("from"))
		assert.Equal(t, "CNY", r.URL.Query().Get("to"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"amount":1.0,
			"base":"USD",
			"start_date":"2024-06-01",
			"end_date":"2024-06-03",
			"rates":{
				"2024-06-01":{"CNY":7.20},
				"2024-06-03":{"CNY":7.21}
			}
		}`))
	}))
	defer srv.Close()

	since, _ := time.Parse("2006-01-02", "2024-06-01")
	prices, err := fetchHistoricalFrom(srv.URL, "USD", "CNY", since)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(prices))
	// rates sorted by date ascending
	assert.Equal(t, "2024-06-01", prices[0].Date.Format("2006-01-02"))
	assert.True(t, prices[0].Value.Equal(decimal.NewFromFloat(7.20)))
	assert.Equal(t, "2024-06-03", prices[1].Date.Format("2006-01-02"))
	assert.True(t, prices[1].Value.Equal(decimal.NewFromFloat(7.21)))
	assert.Equal(t, "USDCNY", prices[0].CommodityID)
	assert.Equal(t, "USDCNY", prices[0].CommodityName)
}

// TestProviderCode confirms the provider registers under the expected key.
func TestProviderCode(t *testing.T) {
	p := &PriceProvider{}
	assert.Equal(t, "cn-boc", p.Code())
}
