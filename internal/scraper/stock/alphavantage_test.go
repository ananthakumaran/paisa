package stock

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAlphaVantageReturnsUnconvertedPrices(t *testing.T) {
	requests := 0
	originalClient := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: yahooRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		require.Equal(t, "TIME_SERIES_DAILY", req.URL.Query().Get("function"))
		require.Equal(t, "AAPL", req.URL.Query().Get("symbol"))
		require.Equal(t, "test-key", req.URL.Query().Get("apikey"))
		body := `{"Time Series (Daily)":{"2024-01-01":{"4. close":"100.25"},"2024-01-02":{"4. close":"0"}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	t.Cleanup(func() { http.DefaultClient = originalClient })

	prices, err := getHistory("test-key:AAPL:USD", "APPLE")
	require.NoError(t, err)
	require.Len(t, prices, 2)
	values := map[string]string{}
	for _, p := range prices {
		values[p.Date.Format("2006-01-02")] = p.Value.String()
		require.Equal(t, "test-key:AAPL:USD", p.CommodityID)
		require.Equal(t, "APPLE", p.CommodityName)
	}
	require.Equal(t, map[string]string{"2024-01-01": "100.25", "2024-01-02": "0"}, values)
	require.Equal(t, 1, requests)
	_, err = getHistory("test-key:AAPL", "APPLE")
	require.ErrorContains(t, err, "Invalid code")
	require.Equal(t, 1, requests)
}
