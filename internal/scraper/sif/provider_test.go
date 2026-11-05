package sif

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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func mockResponse(t *testing.T, status int, body string, check func(*http.Request)) {
	t.Helper()
	original := client
	client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
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
	t.Cleanup(func() { client = original })
}

func TestGetPrices(t *testing.T) {
	mockResponse(t, http.StatusOK, `{"data":[{"date":"2026-03-01","nav":10.123456789012345678}]}`, func(req *http.Request) {
		user, password, ok := req.BasicAuth()
		require.True(t, ok)
		require.Equal(t, "play", user)
		require.Empty(t, password)
		require.Equal(t, "JSON", req.URL.Query().Get("default_format"))
		require.Contains(t, req.URL.Query().Get("query"), "scheme_id = {scheme_id:String}")
		require.Equal(t, "001234", req.URL.Query().Get("param_scheme_id"))
	})

	prices, err := (&PriceProvider{}).GetPrices("001234", "MY_SIF")
	require.NoError(t, err)
	require.Len(t, prices, 1)
	require.Equal(t, "001234", prices[0].CommodityID)
	require.Equal(t, "MY_SIF", prices[0].CommodityName)
	require.Equal(t, config.SIF, prices[0].CommodityType)
	require.Equal(t, "10.123456789012345678", prices[0].Value.String())
	require.Equal(t, time.Date(2026, 3, 1, 0, 0, 0, 0, config.TimeZone()), prices[0].Date)
}

func TestGetNavErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"HTTP failure", http.StatusServiceUnavailable, "unavailable", "Service Unavailable"},
		{"invalid JSON", http.StatusOK, "invalid", "decode FinBodhi SIF response"},
		{"empty history", http.StatusOK, `{"data":[]}`, "no NAV history"},
		{"missing NAV", http.StatusOK, `{"data":[{"date":"2026-03-01","nav":null}]}`, "invalid NAV"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mockResponse(t, tc.status, tc.body, nil)
			prices, err := GetNav("1234", "MY_SIF")
			require.ErrorContains(t, err, tc.want)
			require.Nil(t, prices)
		})
	}
}

func TestAutoComplete(t *testing.T) {
	for _, search := range []string{"Alpha", "001234", " O'Brien ", ""} {
		t.Run(search, func(t *testing.T) {
			mockResponse(t, http.StatusOK, `{"data":[{"id":"001234","nav_name":"Alpha Direct Growth"}]}`, func(req *http.Request) {
				sql := req.URL.Query().Get("query")
				require.Contains(t, sql, "from sif_scheme")
				require.Contains(t, sql, "splitByNonAlpha(lower({search:String}))")
				require.Contains(t, sql, "order by matched_tokens desc, nav_name, id limit 200")
				require.Equal(t, strings.TrimSpace(search), req.URL.Query().Get("param_search"))
				if strings.TrimSpace(search) != "" {
					require.Contains(t, sql, "where hasAnyTokens(lower(nav_name), lower({search:String})) or id = {search:String}")
					require.NotContains(t, sql, "'"+strings.TrimSpace(search)+"'")
				} else {
					require.NotContains(t, sql, "where")
				}
			})

			want := []price.AutoCompleteItem{{ID: "001234", Label: "Alpha Direct Growth — 001234"}}
			require.Equal(t, want, (&PriceProvider{}).AutoComplete(nil, "scheme", map[string]string{"scheme": search}))
		})
	}
}

func TestGetSchemesEmpty(t *testing.T) {
	mockResponse(t, http.StatusOK, `{"data":[]}`, nil)
	schemes, err := GetSchemes("unknown")
	require.NoError(t, err)
	require.Empty(t, schemes)
}
