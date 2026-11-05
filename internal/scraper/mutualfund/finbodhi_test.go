package mutualfund

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

type finbodhiRoundTripFunc func(*http.Request) (*http.Response, error)

func (f finbodhiRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func mockFinBodhiResponse(t *testing.T, status int, body string, check func(*http.Request)) {
	t.Helper()
	original := finbodhiClient
	finbodhiClient = &http.Client{Transport: finbodhiRoundTripFunc(func(req *http.Request) (*http.Response, error) {
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
	for _, search := range []string{"Nifty", "120716", " O'Brien ", ""} {
		t.Run(search, func(t *testing.T) {
			mockFinBodhiResponse(t, http.StatusOK, `{"data":[{"code":"120716","name":"UTI Nifty 50 Index Fund - Direct Plan - Growth"}]}`, func(req *http.Request) {
				sql := req.URL.Query().Get("query")
				require.Contains(t, sql, "nav_name as name, toString(code) as code")
				require.Contains(t, sql, "from scheme")
				require.Contains(t, sql, "splitByNonAlpha(lower({search:String}))")
				require.Contains(t, sql, "order by matched_tokens desc, nav_name, code limit 200")
				require.Equal(t, strings.TrimSpace(search), req.URL.Query().Get("param_search"))
				if strings.TrimSpace(search) != "" {
					require.Contains(t, sql, "where hasAnyTokens(lower(nav_name), lower({search:String})) or toString(code) = {search:String}")
					require.NotContains(t, sql, "'"+strings.TrimSpace(search)+"'")
				} else {
					require.NotContains(t, sql, "where")
				}
			})

			want := []price.AutoCompleteItem{{ID: "120716", Label: "UTI Nifty 50 Index Fund - Direct Plan - Growth — 120716"}}
			require.Equal(t, want, (&FinBodhiPriceProvider{}).AutoComplete(nil, "scheme", map[string]string{"scheme": search}))
		})
	}
}

func TestFinBodhiGetPrices(t *testing.T) {
	mockFinBodhiResponse(t, http.StatusOK, `{"data":[{"date":"2026-03-01","nav":123.123456789012345678},{"date":"2026-03-02","nav":"124.5"}]}`, func(req *http.Request) {
		user, password, ok := req.BasicAuth()
		require.True(t, ok)
		require.Equal(t, "play", user)
		require.Empty(t, password)
		require.Equal(t, "JSON", req.URL.Query().Get("default_format"))
		require.Contains(t, req.URL.Query().Get("query"), "from nav where scheme_code = {scheme_code:String}")
		require.Equal(t, "120716", req.URL.Query().Get("param_scheme_code"))
	})

	prices, err := (&FinBodhiPriceProvider{}).GetPrices("120716", "NIFTY")
	require.NoError(t, err)
	require.Len(t, prices, 2)
	require.Equal(t, "120716", prices[0].CommodityID)
	require.Equal(t, "NIFTY", prices[0].CommodityName)
	require.Equal(t, config.MutualFund, prices[0].CommodityType)
	require.Equal(t, "123.123456789012345678", prices[0].Value.String())
	require.Equal(t, "124.5", prices[1].Value.String())
	require.Equal(t, time.Date(2026, 3, 1, 0, 0, 0, 0, config.TimeZone()), prices[0].Date)
}

func TestFinBodhiErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"HTTP failure", http.StatusServiceUnavailable, "unavailable", "Service Unavailable"},
		{"invalid JSON", http.StatusOK, "invalid", "decode FinBodhi mutual fund response"},
		{"empty history", http.StatusOK, `{"data":[]}`, "no NAV history"},
		{"missing NAV", http.StatusOK, `{"data":[{"date":"2026-03-01","nav":null}]}`, "invalid NAV"},
		{"invalid date", http.StatusOK, `{"data":[{"date":"invalid","nav":10}]}`, "invalid NAV date"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mockFinBodhiResponse(t, tc.status, tc.body, nil)
			prices, err := (&FinBodhiPriceProvider{}).GetPrices("120716", "NIFTY")
			require.ErrorContains(t, err, tc.want)
			require.Nil(t, prices)
		})
	}
}

func TestFinBodhiSearchEmpty(t *testing.T) {
	mockFinBodhiResponse(t, http.StatusOK, `{"data":[]}`, nil)
	schemes, err := searchFinBodhi("unknown")
	require.NoError(t, err)
	require.Empty(t, schemes)
}
