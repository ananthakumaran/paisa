// Package okx implements a price provider that pulls historical daily
// closing prices for crypto pairs from OKX's public market data API.
//
// API: https://www.okx.com/api/v5/market/history-candles
// Auth: not required for the public endpoint.
//
// Pagination (per OKX v5 docs): the response is sorted DESCENDING by ts.
//   - `before=<ts_ms>` returns candles strictly NEWER than ts (forward walk).
//   - `after=<ts_ms>`  returns candles strictly OLDER than ts (backward walk).
//
// For historical backfill we walk BACKWARDS in time, so we use `after`,
// seeded with the oldest ts from the previous page.
//
// OKX returns prices denominated in the quote currency of the instrument
// (e.g. USDT for BTC-USDT). Conversion to the user's default currency is
// handled downstream by the FX layer.
package okx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
)

const (
	// defaultEndpoint is the public OKX `history-candles` URL. Tests
	// substitute an httptest URL via fetchAll's endpoint argument.
	defaultEndpoint = "https://www.okx.com/api/v5/market/history-candles"

	// pageLimit is the per-page size requested from OKX. OKX accepts up
	// to 100 here.
	pageLimit = 100

	// maxPages is a hard upper bound on how many pages we will fetch in
	// a single GetPrices call. At pageLimit=100 daily candles per page
	// this covers ~27 years of history, well beyond any crypto's
	// lifetime. It exists to guarantee termination if the upstream
	// returns degenerate (non-advancing) responses.
	maxPages = 100

	// httpTimeout caps a single HTTP request. OKX international routes
	// can be slow from some regions; without a timeout the loop could
	// stall indefinitely on a single page.
	httpTimeout = 30 * time.Second
)

// httpClient is shared across fetches so we get a single connection pool
// plus a sane request timeout.
var httpClient = &http.Client{Timeout: httpTimeout}

// okxResponse models the JSON envelope OKX returns from the candles
// endpoint. Each candle is encoded as a heterogeneous string array:
//
//	[ts_ms, open, high, low, close, vol, volCcy, volCcyQuote, confirm]
//
// We only consume index 0 (timestamp) and index 4 (close).
type okxResponse struct {
	Code string     `json:"code"`
	Msg  string     `json:"msg"`
	Data [][]string `json:"data"`
}

// parsedPage is the structured form of one OKX response page. `oldestMs`
// is the raw ts_ms string of the oldest candle on the page (the last row
// since OKX returns descending order), preserved verbatim so the cursor
// keeps full millisecond precision when fed back as `after=` on the next
// request.
type parsedPage struct {
	prices   []*price.Price
	oldestMs string
}

// parseCandles converts one page of the OKX response into price records
// plus the raw oldest-ts cursor string. Order is preserved — OKX returns
// descending by timestamp; callers that care about chronological order
// should sort downstream.
func parseCandles(body []byte, code string, commodityName string) (*parsedPage, error) {
	var resp okxResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("okx: decode response: %w", err)
	}
	if resp.Code != "0" {
		return nil, fmt.Errorf("okx: api error code=%s msg=%q", resp.Code, resp.Msg)
	}

	page := &parsedPage{prices: make([]*price.Price, 0, len(resp.Data))}
	for _, row := range resp.Data {
		if len(row) < 5 {
			return nil, fmt.Errorf("okx: malformed candle row, expected >= 5 fields, got %d", len(row))
		}
		tsMs, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("okx: invalid timestamp %q: %w", row[0], err)
		}
		closeVal, err := decimal.NewFromString(row[4])
		if err != nil {
			return nil, fmt.Errorf("okx: invalid close %q: %w", row[4], err)
		}

		// Crypto markets trade 24/7 and OKX defines daily candles on a
		// UTC midnight boundary. Anchor the Price.Date in UTC so the
		// day doesn't shift under negative timezones.
		date := time.UnixMilli(tsMs).UTC()
		page.prices = append(page.prices, &price.Price{
			Date:          date,
			CommodityType: config.Unknown,
			CommodityID:   code,
			CommodityName: commodityName,
			Value:         closeVal,
		})
	}
	if n := len(resp.Data); n > 0 {
		page.oldestMs = resp.Data[n-1][0]
	}
	return page, nil
}

// fetchAll walks OKX's pagination backwards in time until the server
// returns an empty page or pagination stops advancing. The `endpoint`
// parameter is the full URL for the history-candles resource (so tests
// can substitute httptest URLs).
func fetchAll(endpoint string, instID string, commodityName string) ([]*price.Price, error) {
	var (
		all   []*price.Price
		after string // empty on the first request; "" means "latest"
	)

	for page := 0; page < maxPages; page++ {
		body, err := getPage(endpoint, instID, after)
		if err != nil {
			return nil, err
		}
		parsed, err := parseCandles(body, instID, commodityName)
		if err != nil {
			return nil, err
		}
		if len(parsed.prices) == 0 {
			// Empty page: we've walked off the end of history.
			break
		}

		all = append(all, parsed.prices...)

		// Degenerate-response guard: if the oldest timestamp didn't
		// advance, stop to avoid an infinite loop.
		if parsed.oldestMs == after {
			log.Warnf("okx: pagination stalled at after=%s, terminating", after)
			break
		}
		after = parsed.oldestMs
	}

	return all, nil
}

// getPage performs a single GET against the OKX candles endpoint.
//
// `after` is the cursor in OKX's pagination contract: the server returns
// candles strictly OLDER than this timestamp (ms). Pass an empty string
// on the first request to get the most recent `pageLimit` candles.
func getPage(endpoint string, instID string, after string) ([]byte, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("okx: invalid endpoint %q: %w", endpoint, err)
	}
	q := u.Query()
	q.Set("instId", instID)
	q.Set("bar", "1D")
	q.Set("limit", strconv.Itoa(pageLimit))
	if after != "" {
		q.Set("after", after)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("okx: GET %s: %w", u.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("okx: GET %s returned %s", u.String(), resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// PriceProvider is the cn-okx price provider implementation.
type PriceProvider struct{}

func (p *PriceProvider) Code() string  { return "cn-okx" }
func (p *PriceProvider) Label() string { return "OKX Crypto" }
func (p *PriceProvider) Description() string {
	return "Fetches daily closing prices for crypto pairs (e.g. BTC-USDT, ETH-USDT) from OKX's public market data API. Prices are denominated in the quote currency of the instrument; conversion to the default currency is handled by the FX layer."
}

func (p *PriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{
			Label:     "Instrument",
			ID:        "instId",
			Help:      "OKX instrument id, e.g. BTC-USDT, ETH-USDT, BTC-USD",
			InputType: "text",
		},
	}
}

func (p *PriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	return []price.AutoCompleteItem{}
}

func (p *PriceProvider) ClearCache(db *gorm.DB) {}

func (p *PriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	log.Info("Fetching crypto price history from OKX for ", code)
	return fetchAll(defaultEndpoint, code, commodityName)
}
