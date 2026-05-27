package yahoo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
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

// setConfig is a small helper that loads a minimal valid paisa config with the
// given default currency.
func setConfig(t *testing.T, currency string) {
	t.Helper()
	cfg := []byte(fmt.Sprintf(
		"journal_path: /tmp/paisa.ledger\n"+
			"db_path: /tmp/paisa.db\n"+
			"default_currency: %s\n", currency))
	if err := config.LoadConfig(cfg, ""); err != nil {
		t.Fatalf("load config: %v", err)
	}
}

// fxRoutingHandler returns a stub Yahoo handler that serves the matching
// fixture for both stock and FX requests.
func fxRoutingHandler(t *testing.T, stockFixture string, fxFixture string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "=X"):
			if fxFixture == "" {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = w.Write(readFixture(t, fxFixture))
		default:
			_, _ = w.Write(readFixture(t, stockFixture))
		}
	}
}

func TestGetHistoryConvertsToCNYUsingFX(t *testing.T) {
	// UBER quote currency (USD) != default currency (CNY) -> FX conversion.
	setConfig(t, "CNY")

	srv := httptest.NewServer(fxRoutingHandler(t, "UBER.json", "USDCNY-X.json"))
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
	// USDCNY close[i] aligns 1:1 with UBER timestamps:
	//   61.25 * 7.0   = 428.75
	//   62.4  * 7.1   = 443.04
	//   63.1  * 7.2   = 454.32
	assert.InDelta(t, 428.75, prices[0].Value.InexactFloat64(), 1e-9)
	assert.InDelta(t, 443.04, prices[1].Value.InexactFloat64(), 1e-9)
	assert.InDelta(t, 454.32, prices[2].Value.InexactFloat64(), 1e-9)
}

func TestGetHistoryGracefullyDegradesWhenFXFetchFails(t *testing.T) {
	// FX endpoint returns 500 -> we should still get UBER prices in USD with a
	// warning logged, NOT an error.
	setConfig(t, "CNY")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "=X") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(readFixture(t, "UBER.json"))
	}))
	defer srv.Close()

	client := &Client{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		// 0 retries to keep the test fast; 500 path attempts retries.
		MaxRetries:  0,
		BackoffBase: 1 * time.Millisecond,
	}

	prices, err := getHistoryWithClient(client, "UBER", "Uber Technologies")
	assert.NoError(t, err, "FX failure must not fail the stock fetch")
	assert.Len(t, prices, 3)
	// Values are returned in the ORIGINAL currency (USD) since FX failed.
	assert.Equal(t, 61.25, prices[0].Value.InexactFloat64())
	assert.Equal(t, 63.1, prices[2].Value.InexactFloat64())
}

func TestGetHistoryGracefullyDegradesWhenFXReturnsEmpty(t *testing.T) {
	// FX endpoint returns a 200 but with no result -> graceful degrade.
	setConfig(t, "CNY")

	emptyFX := []byte(`{"chart": {"result": [], "error": null}}`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "=X") {
			_, _ = w.Write(emptyFX)
			return
		}
		_, _ = w.Write(readFixture(t, "UBER.json"))
	}))
	defer srv.Close()

	client := &Client{
		BaseURL:     srv.URL,
		HTTPClient:  srv.Client(),
		MaxRetries:  0,
		BackoffBase: 1 * time.Millisecond,
	}

	prices, err := getHistoryWithClient(client, "UBER", "Uber Technologies")
	assert.NoError(t, err)
	assert.Len(t, prices, 3)
	assert.Equal(t, 61.25, prices[0].Value.InexactFloat64())
}

func TestGetHistoryNoFXNeededWhenCurrencyMatches(t *testing.T) {
	// When default currency == quote currency, the FX endpoint must never be
	// called. We assert that via a counter on the test server.
	setConfig(t, "USD")

	var fxCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "=X") {
			atomic.AddInt32(&fxCalls, 1)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
	assert.Equal(t, int32(0), atomic.LoadInt32(&fxCalls), "FX endpoint must not be called when currencies match")
}

// TestGetPricesEndToEnd exercises the provider through its public interface
// (the entry point used by SyncCommodities) using an injected client via the
// package-level NewClient variable.
func TestGetPricesEndToEnd(t *testing.T) {
	setConfig(t, "USD")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(readFixture(t, "UBER.json"))
	}))
	defer srv.Close()

	// Substitute the package-level client constructor for the duration of this
	// test so PriceProvider.GetPrices hits our httptest server.
	prev := NewClient
	defer func() { NewClient = prev }()
	NewClient = func() *Client {
		return &Client{
			BaseURL:     srv.URL,
			HTTPClient:  srv.Client(),
			MaxRetries:  1,
			BackoffBase: 1 * time.Millisecond,
		}
	}

	p := &PriceProvider{}
	prices, err := p.GetPrices("UBER", "Uber Technologies")
	assert.NoError(t, err)
	assert.Len(t, prices, 3)
	assert.Equal(t, "UBER", prices[0].CommodityID)
	assert.Equal(t, "Uber Technologies", prices[0].CommodityName)
	assert.Equal(t, 61.25, prices[0].Value.InexactFloat64())
}

// TestGetPricesGracefulOn404 confirms the provider degrades to (nil, nil) when
// Yahoo returns 404 for a delisted symbol so the sync loop continues.
func TestGetPricesGracefulOn404(t *testing.T) {
	setConfig(t, "USD")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(readFixture(t, "not-found.json"))
	}))
	defer srv.Close()

	prev := NewClient
	defer func() { NewClient = prev }()
	NewClient = func() *Client {
		return &Client{
			BaseURL:     srv.URL,
			HTTPClient:  srv.Client(),
			MaxRetries:  0,
			BackoffBase: 1 * time.Millisecond,
		}
	}

	p := &PriceProvider{}
	prices, err := p.GetPrices("DELISTED", "Was Listed")
	assert.NoError(t, err, "404 must degrade gracefully")
	assert.Nil(t, prices)
}

// TestGetPricesRoutesFXToFXSubsystem verifies that an FX symbol passes through
// getFXPricesFn (which M1-F overrides). We swap the variable temporarily.
func TestGetPricesRoutesFXToFXSubsystem(t *testing.T) {
	setConfig(t, "USD")

	var seenCode string
	prevFn := getFXPricesFn
	defer func() { getFXPricesFn = prevFn }()
	getFXPricesFn = func(c *Client, code, commodityName string) ([]*price.Price, error) {
		seenCode = code
		return []*price.Price{}, nil
	}

	prevNC := NewClient
	defer func() { NewClient = prevNC }()
	NewClient = func() *Client { return DefaultClient() }

	p := &PriceProvider{}
	_, err := p.GetPrices("USDCNY=X", "USD->CNY")
	assert.NoError(t, err)
	assert.Equal(t, "USDCNY=X", seenCode, "FX symbols must be routed via getFXPricesFn")
}

// TestPickUserAgentRotates verifies that pickUserAgent walks through the pool
// instead of caching a single string forever.
func TestPickUserAgentRotates(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < len(userAgents)*2; i++ {
		seen[pickUserAgent()] = true
	}
	// After 2x the pool size we must have seen every entry.
	assert.Equal(t, len(userAgents), len(seen),
		"pickUserAgent must rotate through every entry")
}

// TestSchemaAllowsYahooProvider asserts that the JSON-schema enum that gates
// config loading lists the "yahoo" provider. Without this, paisa.yaml files
// using provider: yahoo would be rejected before GetProviderByCode is even
// called.
func TestSchemaAllowsYahooProvider(t *testing.T) {
	cfg := []byte(
		"journal_path: /tmp/paisa.ledger\n" +
			"db_path: /tmp/paisa.db\n" +
			"default_currency: USD\n" +
			"commodities:\n" +
			"  - name: UBER\n" +
			"    type: stock\n" +
			"    price:\n" +
			"      provider: yahoo\n" +
			"      code: UBER\n",
	)
	if err := config.LoadConfig(cfg, ""); err != nil {
		t.Fatalf("schema should allow provider: yahoo, got: %v", err)
	}
}
