package currency

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withServer(t *testing.T, handler http.HandlerFunc) {
	server := httptest.NewServer(handler)
	original := BaseURL
	BaseURL = server.URL
	t.Cleanup(func() {
		BaseURL = original
		server.Close()
	})
}

func TestGetHistoryWalksBackUntilEmpty(t *testing.T) {
	until := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	var ranges []string

	withServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "USD", r.URL.Query().Get("base"))
		assert.Equal(t, "INR", r.URL.Query().Get("symbols"))
		dateRange := strings.TrimPrefix(r.URL.Path, "/")
		ranges = append(ranges, dateRange)

		switch len(ranges) {
		case 1:
			fmt.Fprint(w, `{"amount":1,"base":"USD","rates":{"2026-10-30":{"INR":94.955245},"2026-10-29":{"INR":95.269093}}}`)
		case 2:
			// a date where the symbol is missing must be skipped
			fmt.Fprint(w, `{"amount":1,"base":"USD","rates":{"2025-01-02":{"INR":85.5},"2025-01-03":{}}}`)
		default:
			http.Error(w, `{"error":"No published rates in that range"}`, http.StatusNotFound)
		}
	})

	prices, err := GetHistory("usd", "INR", "USD", until)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"2025-10-31..2026-10-31",
		"2024-10-30..2025-10-30",
		"2023-10-30..2024-10-29",
	}, ranges)

	require.Len(t, prices, 3)
	assert.Equal(t, "2025-01-02", prices[0].Date.Format("2006-01-02"))
	assert.Equal(t, "85.5", prices[0].Value.String())
	assert.Equal(t, "2026-10-29", prices[1].Date.Format("2006-01-02"))
	assert.Equal(t, "2026-10-30", prices[2].Date.Format("2006-01-02"))
	assert.Equal(t, "94.955245", prices[2].Value.String())
	for _, p := range prices {
		assert.Equal(t, "USD", p.CommodityID)
		assert.Equal(t, "USD", p.CommodityName)
	}
}

func TestGetHistoryStopsAtEarliest(t *testing.T) {
	requests := 0
	withServer(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		fmt.Fprint(w, `{"rates":{"2000-06-01":{"INR":45}}}`)
	})

	// 2000-01-02..2001-01-01, then 1999-01-01..2000-01-01, then stop
	prices, err := GetHistory("USD", "INR", "USD", time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	assert.Equal(t, 2, requests)
	assert.Len(t, prices, 2)
}

func TestGetHistoryErrors(t *testing.T) {
	until := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)

	t.Run("server error", func(t *testing.T) {
		withServer(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "boom", http.StatusInternalServerError)
		})
		_, err := GetHistory("USD", "INR", "USD", until)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("unsupported currency", func(t *testing.T) {
		withServer(t, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"rates":{}}`)
		})
		_, err := GetHistory("ARS", "INR", "ARS", until)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ARS/INR")
	})

	t.Run("same currency", func(t *testing.T) {
		withServer(t, func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("unexpected request: %s", r.URL)
		})
		_, err := GetHistory("INR", "INR", "INR", until)
		require.Error(t, err)
	})
}
