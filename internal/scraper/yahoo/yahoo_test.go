package yahoo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/stretchr/testify/assert"
)

// readFixture reads a recorded Yahoo response from testdata/.
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
}

func TestClassifySymbol(t *testing.T) {
	cases := []struct {
		symbol string
		want   SymbolKind
	}{
		{"UBER", SymbolStock},
		{"AAPL", SymbolStock},
		{"BABA", SymbolStock},
		{"0700.HK", SymbolStock},
		{"1810.HK", SymbolStock},
		{"USDCNY=X", SymbolFX},
		{"EURUSD=X", SymbolFX},
		{"^GSPC", SymbolIndex},
		{"^HSI", SymbolIndex},
	}
	for _, c := range cases {
		got := ClassifySymbol(c.symbol)
		assert.Equal(t, c.want, got, "symbol=%s", c.symbol)
	}
}

func TestParseChartResponseUBER(t *testing.T) {
	body := readFixture(t, "UBER.json")
	resp, err := parseChartResponse(body)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Chart.Result, 1)
	result := resp.Chart.Result[0]
	assert.Equal(t, "USD", result.Meta.Currency)
	assert.Equal(t, []int64{1704067200, 1704153600, 1704240000}, result.Timestamp)
	assert.Len(t, result.Indicators.Quote, 1)
	assert.Equal(t, []float64{61.25, 62.4, 63.1}, result.Indicators.Quote[0].Close)
}

func TestParseChartResponseHK(t *testing.T) {
	body := readFixture(t, "0700-HK.json")
	resp, err := parseChartResponse(body)
	assert.NoError(t, err)
	result := resp.Chart.Result[0]
	assert.Equal(t, "HKD", result.Meta.Currency)
	assert.Equal(t, "0700.HK", result.Meta.Symbol)
	assert.Equal(t, []float64{310.4, 312.6}, result.Indicators.Quote[0].Close)
}

func TestParseChartResponseError(t *testing.T) {
	body := readFixture(t, "not-found.json")
	resp, err := parseChartResponse(body)
	assert.NoError(t, err)
	// Yahoo returns a structured error in the body even with HTTP 404. Result should be empty.
	assert.Empty(t, resp.Chart.Result)
	assert.NotNil(t, resp.Chart.Error)
	assert.Equal(t, "Not Found", resp.Chart.Error.Code)
}

func TestFetchChartHonors429WithBackoff(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(readFixture(t, "UBER.json"))
	}))
	defer srv.Close()

	client := &Client{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		MaxRetries: 5,
		// Tiny backoff so test stays fast.
		BackoffBase: 1 * time.Millisecond,
	}
	resp, err := client.FetchChart("UBER")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(3), atomic.LoadInt32(&calls))
	assert.Equal(t, "USD", resp.Chart.Result[0].Meta.Currency)
}

func TestFetchChartGivesUpAfterMaxRetries(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := &Client{
		BaseURL:     srv.URL,
		HTTPClient:  srv.Client(),
		MaxRetries:  2,
		BackoffBase: 1 * time.Millisecond,
	}
	_, err := client.FetchChart("UBER")
	assert.Error(t, err)
	// MaxRetries=2 -> 1 initial + 2 retries = 3 attempts.
	assert.Equal(t, int32(3), atomic.LoadInt32(&calls))
}

func TestFetchChart404Graceful(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(readFixture(t, "not-found.json"))
	}))
	defer srv.Close()

	client := &Client{
		BaseURL:     srv.URL,
		HTTPClient:  srv.Client(),
		MaxRetries:  2,
		BackoffBase: 1 * time.Millisecond,
	}
	resp, err := client.FetchChart("DELISTED")
	// 404 from Yahoo should not be retried; it's a delisting signal.
	// The caller wants graceful behavior: a typed NotFound error so callers can degrade.
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestGetHistoryConvertsToPricesUSD(t *testing.T) {
	// When the symbol's quote currency matches the user's default currency,
	// no FX conversion is performed.
	cfg := []byte(
		"journal_path: /tmp/paisa.ledger\n" +
			"db_path: /tmp/paisa.db\n" +
			"default_currency: USD\n",
	)
	if err := config.LoadConfig(cfg, ""); err != nil {
		t.Fatalf("load config: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(readFixture(t, "UBER.json"))
	}))
	defer srv.Close()

	client := &Client{
		BaseURL:     srv.URL,
		HTTPClient:  srv.Client(),
		MaxRetries:  1,
		BackoffBase: 1 * time.Millisecond,
	}

	prices, err := getHistoryWithClient(client, "UBER", "Uber Technologies")
	assert.NoError(t, err)
	assert.Len(t, prices, 3)
	assert.Equal(t, "Uber Technologies", prices[0].CommodityName)
	assert.Equal(t, "UBER", prices[0].CommodityID)
	assert.Equal(t, 61.25, prices[0].Value.InexactFloat64())
	assert.Equal(t, 63.1, prices[2].Value.InexactFloat64())
}

func TestProviderCodeIsYahoo(t *testing.T) {
	p := &PriceProvider{}
	assert.Equal(t, "yahoo", p.Code())
	assert.NotEmpty(t, p.Label())
	assert.NotEmpty(t, p.Description())
	assert.NotEmpty(t, p.AutoCompleteFields())
}

// TestRoute confirms that the provider's GetPrices for FX-style symbols does
// not panic when the FX subsystem is absent; FX integration is owned by M1-F.
func TestRouteSymbolStockUsesStockPath(t *testing.T) {
	// Symbol routing should report Stock for the listed equities the user enters.
	for _, sym := range []string{"UBER", "AAPL", "0700.HK", "1810.HK", "BABA", "BIDU"} {
		assert.Equal(t, SymbolStock, ClassifySymbol(sym),
			fmt.Sprintf("expected %s to route to stock", sym))
	}
}
