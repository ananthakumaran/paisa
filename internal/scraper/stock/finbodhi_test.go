package stock

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/stretchr/testify/require"
)

func mockFinBodhiResponse(t *testing.T, status int, body string, check func(*http.Request)) {
	t.Helper()
	original := finbodhiClient
	finbodhiClient = &http.Client{Transport: yahooRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if check != nil {
			check(req)
		}
		return &http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})}
	t.Cleanup(func() { finbodhiClient = original })
}

func TestFinBodhiAutoComplete(t *testing.T) {
	for _, search := range []string{"Reliance", "RELIANCE", "INE002A01018", " O'Brien ", ""} {
		t.Run(search, func(t *testing.T) {
			mockFinBodhiResponse(t, http.StatusOK, `{"data":[{"isin":"INE002A01018","name":"Reliance Industries Ltd","ticker":"RELIANCE","exchange":"nse"}]}`, func(req *http.Request) {
				sql := req.URL.Query().Get("query")
				require.Contains(t, sql, "notEmpty(trimBoth(name))")
				require.Contains(t, sql, "exchange in ('bse', 'nse', 'gift', 'NMS', 'NYQ', 'PCX', 'BTS', 'ASE')")
				require.Contains(t, sql, "type in ('equity', 'etf')")
				require.Contains(t, sql, "isin in (select isin from isins)")
				require.Contains(t, sql, "order by name, isin, exchange limit 200")
				require.Equal(t, strings.TrimSpace(search), req.URL.Query().Get("param_search"))
				if strings.TrimSpace(search) != "" {
					require.Contains(t, sql, "name ilike {pattern:String} or ticker ilike {pattern:String} or isin = {search:String}")
					require.Equal(t, "%"+strings.TrimSpace(search)+"%", req.URL.Query().Get("param_pattern"))
					require.NotContains(t, sql, "'"+strings.TrimSpace(search)+"'")
				} else {
					require.NotContains(t, sql, "ilike")
				}
			})

			want := []price.AutoCompleteItem{{ID: "INE002A01018", Label: "Reliance Industries Ltd (RELIANCE, nse) — INE002A01018"}}
			require.Equal(t, want, (&FinBodhiPriceProvider{}).AutoComplete(nil, "security", map[string]string{"security": search}))
		})
	}
}

func TestFinBodhiGetPrices(t *testing.T) {
	for _, isin := range []string{"INE002A01018", "US0378331005"} {
		t.Run(isin, func(t *testing.T) {
			mockFinBodhiResponse(t, http.StatusOK, `{"data":[{"date":"2026-03-01","close":123.123456789012345678},{"date":"2026-03-02","close":0}]}`, func(req *http.Request) {
				user, password, ok := req.BasicAuth()
				require.True(t, ok)
				require.Equal(t, "play", user)
				require.Empty(t, password)
				require.Equal(t, "JSON", req.URL.Query().Get("default_format"))
				require.Contains(t, req.URL.Query().Get("query"), "toDate(date) as date, close from price where isin = {isin:String}")
				require.Equal(t, isin, req.URL.Query().Get("param_isin"))
			})

			prices, err := (&FinBodhiPriceProvider{}).GetPrices(isin, "MY_STOCK")
			require.NoError(t, err)
			require.Len(t, prices, 2)
			require.Equal(t, isin, prices[0].CommodityID)
			require.Equal(t, "MY_STOCK", prices[0].CommodityName)
			require.Equal(t, config.Stock, prices[0].CommodityType)
			require.Equal(t, "123.123456789012345678", prices[0].Value.String())
			require.True(t, prices[1].Value.IsZero())
			require.Equal(t, time.Date(2026, 3, 1, 0, 0, 0, 0, config.TimeZone()), prices[0].Date)
		})
	}
}

func TestFinBodhiErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"HTTP failure", http.StatusServiceUnavailable, "unavailable", "Service Unavailable"},
		{"invalid JSON", http.StatusOK, "invalid", "decode FinBodhi stock response"},
		{"empty history", http.StatusOK, `{"data":[]}`, "no price history"},
		{"missing close", http.StatusOK, `{"data":[{"date":"2026-03-01","close":null}]}`, "invalid closing price"},
		{"invalid date", http.StatusOK, `{"data":[{"date":"invalid","close":10}]}`, "invalid price date"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mockFinBodhiResponse(t, tc.status, tc.body, nil)
			prices, err := (&FinBodhiPriceProvider{}).GetPrices("INE002A01018", "RELIANCE")
			require.ErrorContains(t, err, tc.want)
			require.Nil(t, prices)
		})
	}
}
