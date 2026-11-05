package stock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYahooResultUnmarshalJSON(t *testing.T) {
	for _, tc := range []struct {
		name       string
		timestamps string
		closes     string
		wantDates  []int64
		wantCloses []float64
	}{
		{
			name:       "preserve numeric zero",
			timestamps: "[100,200,300]",
			closes:     "[10,0,20]",
			wantDates:  []int64{100, 200, 300},
			wantCloses: []float64{10, 0, 20},
		},
		{
			name:       "drop leading middle and trailing nulls",
			timestamps: "[100,200,300,400,500]",
			closes:     "[null,10,null,20,null]",
			wantDates:  []int64{200, 400},
			wantCloses: []float64{10, 20},
		},
		{
			name:       "all null",
			timestamps: "[100,200]",
			closes:     "[null,null]",
			wantDates:  []int64{},
		},
		{
			name:       "empty history",
			timestamps: "[]",
			closes:     "[]",
			wantDates:  []int64{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := fmt.Sprintf(`{"timestamp":%s,"meta":{"currency":"INR"},"indicators":{"quote":[{"close":%s}]}}`, tc.timestamps, tc.closes)
			var result Result
			require.NoError(t, json.Unmarshal([]byte(data), &result))
			assert.Equal(t, "INR", result.Meta.Currency)
			assert.Equal(t, tc.wantDates, result.Timestamp)
			require.Len(t, result.Indicators.Quote, 1)
			require.Len(t, result.Indicators.Quote[0].Close, len(tc.wantCloses))
			for i, close := range tc.wantCloses {
				assert.Equal(t, close, result.Indicators.Quote[0].Close[i])
			}
		})
	}
}

type yahooRoundTripFunc func(*http.Request) (*http.Response, error)

func (f yahooRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGetHistoryReturnsUnconvertedPrices(t *testing.T) {
	for _, currency := range []string{"INR", "USD", "IDR", "₹", ""} {
		t.Run("currency="+currency, func(t *testing.T) {
			body := fmt.Sprintf(`{"chart":{"result":[{"timestamp":[100,200,300,400,500],"meta":{"currency":%q},"indicators":{"quote":[{"close":[null,0,null,20,null]}]}}]}}`, currency)
			requests := 0
			originalClient := http.DefaultClient
			http.DefaultClient = &http.Client{Transport: yahooRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				requests++
				require.Equal(t, "/v8/finance/chart/STOCK", req.URL.Path)
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			})}
			t.Cleanup(func() { http.DefaultClient = originalClient })

			prices, err := GetHistory("STOCK", "Stock")
			require.NoError(t, err)
			require.Len(t, prices, 2)
			assert.Equal(t, int64(200), prices[0].Date.Unix())
			assert.True(t, prices[0].Value.IsZero())
			assert.Equal(t, int64(400), prices[1].Date.Unix())
			assert.Equal(t, float64(20), prices[1].Value.InexactFloat64())
			assert.Equal(t, "Stock", prices[1].CommodityName)
			assert.Equal(t, 1, requests)
		})
	}
}
