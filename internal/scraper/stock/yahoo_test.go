package stock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ananthakumaran/paisa/internal/config"
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

func TestGetHistorySkipsNullCloses(t *testing.T) {
	baseCurrency := config.DefaultCurrency()
	foreignCurrency := "IDR"
	if foreignCurrency == baseCurrency {
		foreignCurrency = "USD"
	}

	for _, tc := range []struct {
		name       string
		currency   string
		timestamps string
		closes     string
		exchange   bool
		wantValues []float64
	}{
		{
			name:       "stock quotes in default currency",
			currency:   baseCurrency,
			timestamps: "[100,200,300,400,500]",
			closes:     "[null,10,null,20,null]",
			wantValues: []float64{10, 20},
		},
		{
			name:       "missing exchange rate uses preceding valid rate",
			currency:   foreignCurrency,
			timestamps: "[200,400]",
			closes:     "[10,20]",
			exchange:   true,
			wantValues: []float64{5, 5},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stockBody := fmt.Sprintf(`{"chart":{"result":[{"timestamp":%s,"meta":{"currency":%q},"indicators":{"quote":[{"close":%s}]}}]}}`, tc.timestamps, tc.currency, tc.closes)
			exchangeBody := `{"chart":{"result":[{"timestamp":[100,200,300,500],"indicators":{"quote":[{"close":[0.5,null,0.25,2]}]}}]}}`
			requests := 0
			originalClient := http.DefaultClient
			http.DefaultClient = &http.Client{Transport: yahooRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				requests++
				var body string
				switch req.URL.Path {
				case "/v8/finance/chart/STOCK":
					body = stockBody
				case "/v8/finance/chart/" + foreignCurrency + baseCurrency + "=X":
					require.True(t, tc.exchange, "default-currency stock must not request FX")
					body = exchangeBody
				default:
					t.Fatalf("unexpected Yahoo request: %s", req.URL)
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			})}
			t.Cleanup(func() { http.DefaultClient = originalClient })

			prices, err := GetHistory("STOCK", "Stock")
			require.NoError(t, err)
			require.Len(t, prices, 2)
			for i, p := range prices {
				assert.Equal(t, int64((i+1)*200), p.Date.Unix())
				assert.Equal(t, tc.wantValues[i], p.Value.InexactFloat64())
			}
			if tc.exchange {
				assert.Equal(t, 2, requests)
			} else {
				assert.Equal(t, 1, requests)
			}
		})
	}
}
