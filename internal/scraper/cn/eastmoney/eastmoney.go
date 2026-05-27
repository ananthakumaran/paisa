// Package eastmoney implements a paisa price provider for Chinese A-share
// historical quotes served by Eastmoney (东方财富).
//
// Data source:
//
//	https://push2his.eastmoney.com/api/qt/stock/kline/get
//	  ?secid=<market>.<code>
//	  &fields1=f1,f2,f3,f4,f5
//	  &fields2=f51,f52,f53,f54,f55,f56,f57,f58
//	  &klt=101&fqt=1&beg=0&end=20500101
//
// Market prefixes:
//
//	1   Shanghai (沪) — codes starting with 6
//	0   Shenzhen (深) — codes starting with 0 or 3
//	116 Hong Kong   (5-digit)
//	105 Nasdaq      (ticker)
//	106 NYSE        (ticker)
//
// Users can write either a bare A-share code ("600000", "000001") and let
// the scraper auto-resolve the prefix, or an explicit "market.code" string
// for any other market.
package eastmoney

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
)

// baseURL is a package-level variable so tests can swap it for an httptest
// server. The default points at Eastmoney's public historical kline endpoint.
var baseURL = "https://push2his.eastmoney.com/api/qt/stock/kline/get"

// cacheTTL governs how long a successful kline fetch is reused within a
// single process. It protects the upstream from being hammered when many
// commodities share the provider or the user runs `paisa update` repeatedly.
const cacheTTL = 30 * time.Minute

type cacheEntry struct {
	at     time.Time
	prices []*price.Price
}

var (
	cacheMu sync.Mutex
	cache   = map[string]cacheEntry{}
)

func clearCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = map[string]cacheEntry{}
}

func cacheGet(key string) ([]*price.Price, bool) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	e, ok := cache[key]
	if !ok {
		return nil, false
	}
	if time.Since(e.at) > cacheTTL {
		delete(cache, key)
		return nil, false
	}
	return e.prices, true
}

func cacheSet(key string, prices []*price.Price) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache[key] = cacheEntry{at: time.Now(), prices: prices}
}

// resolveSecid returns the Eastmoney secid for a user-provided code.
// Bare 6-digit A-share codes are auto-prefixed; strings that already contain
// a "." are treated as an explicit market.code and returned unchanged.
func resolveSecid(code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", fmt.Errorf("eastmoney: empty code")
	}
	// Explicit "market.code" passes through.
	if strings.Contains(code, ".") {
		return code, nil
	}
	// Auto-detect A-share market by leading digit. Only 6-digit codes are
	// auto-resolved; anything else must be specified explicitly.
	if len(code) != 6 {
		return "", fmt.Errorf("eastmoney: cannot auto-resolve secid for %q; expected 6-digit A-share code or explicit market.code", code)
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("eastmoney: cannot auto-resolve secid for %q; expected 6-digit A-share code or explicit market.code", code)
		}
	}
	switch code[0] {
	case '6':
		return "1." + code, nil // Shanghai
	case '0', '3':
		return "0." + code, nil // Shenzhen (main + ChiNext)
	default:
		return "", fmt.Errorf("eastmoney: cannot auto-resolve secid for %q; leading digit %c is not a known A-share prefix", code, code[0])
	}
}

type klineResponse struct {
	Data *struct {
		Code   string   `json:"code"`
		Name   string   `json:"name"`
		Klines []string `json:"klines"`
	} `json:"data"`
}

// parseKlines decodes an Eastmoney kline response and returns a slice of
// daily closing prices. Malformed rows are skipped with a warning rather
// than failing the whole sync — eastmoney occasionally returns short or
// empty rows for non-trading days, halts or delisted tickers.
func parseKlines(body []byte, code string, commodityName string) ([]*price.Price, error) {
	var resp klineResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("eastmoney: decode response: %w", err)
	}
	if resp.Data == nil || len(resp.Data.Klines) == 0 {
		return nil, nil
	}

	prices := make([]*price.Price, 0, len(resp.Data.Klines))
	for _, row := range resp.Data.Klines {
		parts := strings.Split(row, ",")
		// Need at least date,open,close.
		if len(parts) < 3 {
			log.Warnf("eastmoney: skipping malformed kline row for %s: %q", code, row)
			continue
		}
		date, err := time.ParseInLocation("2006-01-02", parts[0], config.TimeZone())
		if err != nil {
			log.Warnf("eastmoney: skipping unparseable date in kline for %s: %q", code, row)
			continue
		}
		// Column index 2 is the close price.
		closeVal, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			log.Warnf("eastmoney: skipping non-numeric close in kline for %s: %q", code, row)
			continue
		}
		prices = append(prices, &price.Price{
			Date:          date,
			CommodityType: config.Stock,
			CommodityID:   code,
			CommodityName: commodityName,
			Value:         decimal.NewFromFloat(closeVal),
		})
	}
	return prices, nil
}

func fetchKlines(secid string) ([]byte, error) {
	url := fmt.Sprintf(
		"%s?secid=%s&fields1=f1,f2,f3,f4,f5&fields2=f51,f52,f53,f54,f55,f56,f57,f58&klt=101&fqt=1&beg=0&end=20500101",
		baseURL, secid,
	)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Eastmoney serves the public endpoint without auth, but a UA helps avoid
	// being rate-limited as an automated client.
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; paisa-eastmoney/1.0)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("eastmoney: HTTP %d for secid=%s", resp.StatusCode, secid)
	}

	return io.ReadAll(resp.Body)
}

// PriceProvider is the paisa price provider for Eastmoney historical kline data.
type PriceProvider struct{}

func (p *PriceProvider) Code() string {
	return "cn-eastmoney"
}

func (p *PriceProvider) Label() string {
	return "Eastmoney (东方财富) A-Share"
}

func (p *PriceProvider) Description() string {
	return "Fetches historical daily close prices for Chinese A-shares (Shanghai / Shenzhen) from Eastmoney. " +
		"For Shanghai stocks (6xxxxx) and Shenzhen stocks (0xxxxx / 3xxxxx), use just the 6-digit code; " +
		"for Hong Kong (116), Nasdaq (105) or NYSE (106) use the explicit \"market.code\" form."
}

func (p *PriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{
			Label:     "Code",
			ID:        "code",
			Help:      "A-share code (e.g. 600000, 000001, 300750) or explicit market.code (e.g. 116.00700, 105.AAPL)",
			InputType: "text",
		},
	}
}

func (p *PriceProvider) AutoComplete(_ *gorm.DB, _ string, _ map[string]string) []price.AutoCompleteItem {
	return []price.AutoCompleteItem{}
}

func (p *PriceProvider) ClearCache(_ *gorm.DB) {
	clearCache()
}

// GetPrices fetches the full available daily close history for code from
// Eastmoney. Errors fetching or parsing degrade gracefully: non-trading
// days, halts and delistings come back as an empty slice rather than a
// hard error, so a single bad ticker does not abort the whole sync.
func (p *PriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	secid, err := resolveSecid(code)
	if err != nil {
		return nil, err
	}

	if cached, ok := cacheGet(secid); ok {
		log.Debugf("eastmoney: using cached prices for secid=%s", secid)
		return cached, nil
	}

	log.Infof("Fetching A-share price history from Eastmoney (secid=%s)", secid)
	body, err := fetchKlines(secid)
	if err != nil {
		return nil, err
	}

	prices, err := parseKlines(body, code, commodityName)
	if err != nil {
		return nil, err
	}

	cacheSet(secid, prices)
	return prices, nil
}
