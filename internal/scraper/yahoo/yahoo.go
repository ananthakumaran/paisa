// Package yahoo implements the unified Yahoo Finance price provider.
//
// The package is split across files so different responsibilities can be
// developed in parallel without merge conflicts:
//
//   - yahoo.go (this file): provider entry point, symbol classification, HTTP
//     client with backoff/429/404 handling, response parsing.
//   - stock.go: stock / ETF / index price history fetcher.
//   - fx.go:    FX rate fetcher (owned by issue #11 / M1-F).
//
// The exported provider is registered under the code "yahoo". GetPrices routes
// to the appropriate subsystem based on the input symbol.
package yahoo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	"github.com/ananthakumaran/paisa/internal/model/price"
	log "github.com/sirupsen/logrus"
)

// ErrNotFound is returned when Yahoo reports the symbol is delisted / missing
// (HTTP 404). Callers should treat this as a non-fatal condition.
var ErrNotFound = errors.New("yahoo: symbol not found")

// defaultBaseURL is the public Yahoo Finance chart endpoint.
const defaultBaseURL = "https://query1.finance.yahoo.com/v8/finance/chart"

// userAgents is a small pool of common UA strings; Yahoo throttles aggressively
// when it sees a default Go UA, so we rotate per request via a monotonically
// increasing counter (one round-robin across the pool).
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_7_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Safari/605.1.15",
}

// uaCounter rotates through userAgents one entry per request. Seeded with a
// random offset so concurrent processes don't all start on UA[0].
var uaCounter atomic.Uint64

func init() {
	uaCounter.Store(uint64(rand.Intn(len(userAgents))))
}

func pickUserAgent() string {
	n := uaCounter.Add(1)
	return userAgents[int(n%uint64(len(userAgents)))]
}

// SymbolKind disambiguates the type of Yahoo Finance symbol being requested.
type SymbolKind int

const (
	SymbolStock SymbolKind = iota
	SymbolFX
	SymbolIndex
)

var (
	fxSymbolRegexp    = regexp.MustCompile(`=X$`)
	indexSymbolRegexp = regexp.MustCompile(`^\^`)
)

// ClassifySymbol returns the kind of Yahoo symbol the input represents. The
// classification is purely syntactic and does not contact Yahoo.
func ClassifySymbol(symbol string) SymbolKind {
	s := strings.TrimSpace(symbol)
	switch {
	case fxSymbolRegexp.MatchString(s):
		return SymbolFX
	case indexSymbolRegexp.MatchString(s):
		return SymbolIndex
	default:
		return SymbolStock
	}
}

// Quote mirrors the inner quote array entry in a Yahoo chart response.
type Quote struct {
	Close []float64 `json:"close"`
}

// Indicators wraps the list of quote series; we only use Close prices.
type Indicators struct {
	Quote []Quote `json:"quote"`
}

// Meta carries currency/symbol metadata which is required for currency
// conversion at the caller level.
type Meta struct {
	Currency string `json:"currency"`
	Symbol   string `json:"symbol"`
}

// Result is a single chart series.
type Result struct {
	Timestamp  []int64    `json:"timestamp"`
	Indicators Indicators `json:"indicators"`
	Meta       Meta       `json:"meta"`
}

// ChartError is the error payload Yahoo returns when the symbol is missing.
type ChartError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// Chart is the outer wrapper of a Yahoo chart response.
type Chart struct {
	Result []Result    `json:"result"`
	Error  *ChartError `json:"error"`
}

// Response is the full JSON response from the chart endpoint.
type Response struct {
	Chart Chart `json:"chart"`
}

// Client is a thin Yahoo Finance HTTP client with retry / backoff for 429s
// and a typed error for 404s.
type Client struct {
	BaseURL     string
	HTTPClient  *http.Client
	MaxRetries  int
	BackoffBase time.Duration
}

// DefaultClient returns a Client with sensible production defaults.
func DefaultClient() *Client {
	return &Client{
		BaseURL:     defaultBaseURL,
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
		MaxRetries:  4,
		BackoffBase: 500 * time.Millisecond,
	}
}

// FetchChart requests the daily chart for symbol and returns the parsed body.
// Implements:
//   - exponential backoff on HTTP 429 (Retry-After respected when present)
//   - a graceful ErrNotFound on HTTP 404
//   - a fixed range of ~50 years to mirror the existing com-yahoo provider
func (c *Client) FetchChart(symbol string) (*Response, error) {
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	if c.BaseURL == "" {
		c.BaseURL = defaultBaseURL
	}
	if c.BackoffBase <= 0 {
		c.BackoffBase = 500 * time.Millisecond
	}

	url := fmt.Sprintf("%s/%s?interval=1d&range=50y", c.BaseURL, symbol)

	attempts := c.MaxRetries + 1
	var lastErr error
	for i := 0; i < attempts; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", pickUserAgent())
		req.Header.Set("Accept", "application/json")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			c.sleepBackoff(i, 0)
			continue
		}

		switch {
		case resp.StatusCode == http.StatusNotFound:
			resp.Body.Close()
			return nil, ErrNotFound
		case resp.StatusCode == http.StatusTooManyRequests:
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
			resp.Body.Close()
			lastErr = fmt.Errorf("yahoo: rate limited (429) for %s", symbol)
			if i == attempts-1 {
				break
			}
			c.sleepBackoff(i, retryAfter)
			continue
		case resp.StatusCode >= 500:
			resp.Body.Close()
			lastErr = fmt.Errorf("yahoo: server error %d for %s", resp.StatusCode, symbol)
			c.sleepBackoff(i, 0)
			continue
		case resp.StatusCode != http.StatusOK:
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("yahoo: unexpected status %d for %s: %s", resp.StatusCode, symbol, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			c.sleepBackoff(i, 0)
			continue
		}

		return parseChartResponse(body)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("yahoo: failed to fetch %s", symbol)
	}
	return nil, lastErr
}

func (c *Client) sleepBackoff(attempt int, retryAfter time.Duration) {
	if retryAfter > 0 {
		time.Sleep(retryAfter)
		return
	}
	// 2^attempt * BackoffBase with a small jitter, capped at 30s.
	delay := c.BackoffBase << attempt
	max := 30 * time.Second
	if delay > max || delay <= 0 {
		delay = max
	}
	jitter := time.Duration(rand.Int63n(int64(c.BackoffBase + 1)))
	time.Sleep(delay + jitter)
}

func parseRetryAfter(h string) time.Duration {
	if h == "" {
		return 0
	}
	// Yahoo sends seconds.
	var secs int
	if _, err := fmt.Sscanf(h, "%d", &secs); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return 0
}

// parseChartResponse unmarshals a Yahoo chart response body.
func parseChartResponse(body []byte) (*Response, error) {
	var resp Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// NewClient is the package-level constructor used by PriceProvider to obtain
// a Client. Tests substitute this variable to point at a httptest server.
// M1-F can also override it from fx.go if FX requests need different defaults.
var NewClient = func() *Client { return DefaultClient() }

// PriceProvider implements price.PriceProvider for Yahoo Finance under the
// canonical code "yahoo". It routes incoming symbols to the stock or FX
// fetcher based on ClassifySymbol; FX support is provided by fx.go (owned by
// M1-F) and gracefully degrades when not compiled in.
type PriceProvider struct {
	// client is lazily initialised so tests can inject a fake; if nil, the
	// package-level NewClient is used.
	client *Client
	once   sync.Once
}

func (p *PriceProvider) Code() string  { return "yahoo" }
func (p *PriceProvider) Label() string { return "Yahoo Finance" }

func (p *PriceProvider) Description() string {
	return "Supports HK / US / China-ADR stocks, ETFs, indices and FX rates via Yahoo Finance. " +
		"Use ticker symbols as shown on finance.yahoo.com, e.g. AAPL, UBER, 0700.HK, 1810.HK, BABA, BIDU."
}

func (p *PriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{
			Label:     "Ticker",
			ID:        "ticker",
			Help:      "Yahoo Finance ticker symbol. Examples: AAPL, UBER, 0700.HK, 1810.HK, BABA, USDCNY=X",
			InputType: "text",
		},
	}
}

func (p *PriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	return []price.AutoCompleteItem{}
}

func (p *PriceProvider) ClearCache(db *gorm.DB) {}

// getClient returns the Client for this provider, constructing it lazily via
// the package-level NewClient on first use.
func (p *PriceProvider) getClient() *Client {
	p.once.Do(func() {
		if p.client == nil {
			p.client = NewClient()
		}
	})
	return p.client
}

// getFXPricesFn is the package-level FX fetcher. The default implementation
// treats FX symbols (e.g. USDCNY=X) as raw chart series and returns them
// without currency conversion. M1-F (issue #11) is expected to replace this
// variable from fx.go (via init()) with a richer implementation that
// understands base/quote currencies, caching and the user's default currency.
var getFXPricesFn = func(c *Client, code, commodityName string) ([]*price.Price, error) {
	return getHistoryWithClient(c, code, commodityName)
}

// GetPrices routes the request to the stock or FX backend depending on the
// shape of the symbol. Unknown / 404 / timeout failures are logged and degrade
// to an empty slice + nil error so the parent sync loop keeps any previously
// cached data and continues on.
func (p *PriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	kind := ClassifySymbol(code)
	switch kind {
	case SymbolFX:
		prices, err := getFXPricesFn(p.getClient(), code, commodityName)
		if errors.Is(err, ErrNotFound) {
			log.Warnf("yahoo: %s not found / delisted; using cached prices", code)
			return nil, nil
		}
		return prices, err
	default:
		// Stocks, ETFs and indices all use the same chart endpoint.
		prices, err := getHistoryWithClient(p.getClient(), code, commodityName)
		if errors.Is(err, ErrNotFound) {
			log.Warnf("yahoo: %s not found / delisted; using cached prices", code)
			return nil, nil
		}
		return prices, err
	}
}
