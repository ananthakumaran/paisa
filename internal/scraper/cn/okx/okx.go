// Package okx implements a price provider that pulls historical daily
// closing prices for crypto pairs from OKX's public market data API.
//
// API: https://www.okx.com/api/v5/market/history-candles
// Auth: not required for the public endpoint.
// Pagination: descending by time; the `before` query parameter takes the
// oldest timestamp (ms) from the previous page to walk further back.
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
	// defaultEndpoint is the public OKX market data base URL. The
	// `history-candles` resource is appended in fetchAll so tests can
	// point at a local httptest server.
	defaultEndpoint = "https://www.okx.com/api/v5/market/history-candles"

	// pageLimit is the per-page size requested from OKX. OKX accepts up
	// to 100 here.
	pageLimit = 100

	// maxPages is a hard upper bound on how many pages we will fetch in
	// a single GetPrices call. At pageLimit=100 daily candles per page
	// this covers ~ 27 years of history, which is well beyond any
	// crypto's lifetime. It exists to guarantee termination if the
	// upstream returns degenerate (non-advancing) responses.
	maxPages = 100
)

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

// parseCandles converts one page of the OKX response into price records.
// Order is preserved — OKX returns descending by timestamp, callers that
// care about chronological order should sort downstream.
func parseCandles(body []byte, code string, commodityName string) ([]*price.Price, error) {
	var resp okxResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("okx: decode response: %w", err)
	}
	if resp.Code != "0" {
		return nil, fmt.Errorf("okx: api error code=%s msg=%q", resp.Code, resp.Msg)
	}

	prices := make([]*price.Price, 0, len(resp.Data))
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

		date := time.Unix(tsMs/1000, 0).In(config.TimeZone())
		prices = append(prices, &price.Price{
			Date:          date,
			CommodityType: config.Unknown,
			CommodityID:   code,
			CommodityName: commodityName,
			Value:         closeVal,
		})
	}
	return prices, nil
}

// fetchAll walks OKX's pagination until the server returns an empty page
// or pagination stops advancing. The `endpoint` parameter is the full URL
// for the history-candles resource (so tests can substitute httptest URLs).
func fetchAll(endpoint string, instID string, commodityName string) ([]*price.Price, error) {
	var (
		all    []*price.Price
		before string // empty on the first request
	)

	for page := 0; page < maxPages; page++ {
		body, err := getPage(endpoint, instID, before)
		if err != nil {
			return nil, err
		}
		prices, err := parseCandles(body, instID, commodityName)
		if err != nil {
			return nil, err
		}
		if len(prices) == 0 {
			// Empty page: we've walked off the end of history.
			break
		}

		// OKX returns descending order; the oldest timestamp on this
		// page is the last element. Use it as the `before` cursor for
		// the next page.
		oldest := prices[len(prices)-1]
		oldestMs := strconv.FormatInt(oldest.Date.Unix()*1000, 10)

		all = append(all, prices...)

		// Degenerate-response guard: if the oldest timestamp didn't
		// advance, stop to avoid an infinite loop.
		if oldestMs == before {
			log.Warnf("okx: pagination stalled at before=%s, terminating", before)
			break
		}
		before = oldestMs
	}

	return all, nil
}

// getPage performs a single GET against the OKX candles endpoint.
func getPage(endpoint string, instID string, before string) ([]byte, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("okx: invalid endpoint %q: %w", endpoint, err)
	}
	q := u.Query()
	q.Set("instId", instID)
	q.Set("bar", "1D")
	q.Set("limit", strconv.Itoa(pageLimit))
	if before != "" {
		q.Set("before", before)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
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
