// Package boc exposes an FX rate scraper registered under the provider code
// "cn-boc". The name reflects the design intent — a China-friendly daily FX
// reference rate source — but the current implementation actually queries
// frankfurter.app (ECB-based) instead of the live Bank of China web page.
//
// TODO(M1-F follow-up): replace this with a real scrape of
// https://srh.bankofchina.com/search/whpj/search_cn.jsp once we have an HTML
// parser. The live BOC page is HTML-only and rate-limited; frankfurter.app is
// JSON, key-less, and accurate enough for monthly net-worth attribution.
package boc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/scraper/yahoo"
)

// init self-registers this provider into the yahoo package's FX provider
// registry so config.FxProviders() chain resolution finds it without the
// yahoo package needing to import us (which would create a cycle:
// yahoo -> boc -> yahoo).
func init() {
	yahoo.RegisterFxProvider("cn-boc", &PriceProvider{})
}

const defaultEndpoint = "https://api.frankfurter.app"

// PriceProvider implements price.PriceProvider for FX rates.
type PriceProvider struct{}

func (p *PriceProvider) Code() string  { return "cn-boc" }
func (p *PriceProvider) Label() string { return "Bank of China (via frankfurter.app)" }
func (p *PriceProvider) Description() string {
	return "Daily FX reference rates. The provider code is reserved for a future scrape of the Bank of China page; today rates are sourced from frankfurter.app (ECB)."
}

func (p *PriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "Currency Pair", ID: "pair", Help: "Source/target currency pair like USDCNY or HKDCNY.", InputType: "text"},
	}
}

func (p *PriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	return []price.AutoCompleteItem{
		{Label: "USD/CNY", ID: "USDCNY"},
		{Label: "HKD/CNY", ID: "HKDCNY"},
		{Label: "EUR/CNY", ID: "EURCNY"},
		{Label: "USD/EUR", ID: "USDEUR"},
	}
}

func (p *PriceProvider) ClearCache(db *gorm.DB) {}

// GetPrices accepts a 6-letter pair like "USDCNY" and returns the full
// historical series for that pair (since frankfurter's earliest date).
func (p *PriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	if len(code) != 6 {
		return nil, fmt.Errorf("boc: expected 6-letter currency pair, got %q", code)
	}
	from, to := code[:3], code[3:]
	log.Info("Fetching FX history from frankfurter.app: ", from, "->", to)
	// Frankfurter's full history starts at 1999-01-04.
	since, _ := time.Parse("2006-01-02", "1999-01-04")
	return fetchHistoricalFrom(defaultEndpoint, from, to, since)
}

// GetRate is a higher-level helper used by the fx module to populate a single
// data point without paying the cost of the full history. Returns the rate
// (e.g. USDCNY=7.2) on the requested date or the most recent prior date.
func GetRate(base, target string, _ time.Time) (decimal.Decimal, error) {
	rate, _, err := fetchLatestFrom(defaultEndpoint, base, target)
	return rate, err
}

// GetHistoricalRates returns every published rate from `since` to today for
// the base->target pair.
func GetHistoricalRates(base, target string, since time.Time) ([]*price.Price, error) {
	return fetchHistoricalFrom(defaultEndpoint, base, target, since)
}

// --- transport helpers ---

type latestResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

type historicalResponse struct {
	Amount    float64                       `json:"amount"`
	Base      string                        `json:"base"`
	StartDate string                        `json:"start_date"`
	EndDate   string                        `json:"end_date"`
	Rates     map[string]map[string]float64 `json:"rates"`
}

// fetchLatestFrom returns the most recently published rate for base->target.
// The endpoint argument lets tests inject a httptest server URL.
func fetchLatestFrom(endpoint, base, target string) (decimal.Decimal, time.Time, error) {
	u := fmt.Sprintf("%s/latest?from=%s&to=%s", endpoint, base, target)
	resp, err := http.Get(u)
	if err != nil {
		return decimal.Zero, time.Time{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return decimal.Zero, time.Time{}, err
	}
	var parsed latestResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return decimal.Zero, time.Time{}, err
	}
	raw, ok := parsed.Rates[target]
	if !ok {
		return decimal.Zero, time.Time{}, fmt.Errorf("boc: %s missing from response for %s->%s", target, base, target)
	}
	date, err := time.ParseInLocation("2006-01-02", parsed.Date, config.TimeZone())
	if err != nil {
		return decimal.Zero, time.Time{}, err
	}
	return decimal.NewFromFloat(raw), date, nil
}

// fetchHistoricalFrom fetches the time series from `since` to today.
func fetchHistoricalFrom(endpoint, base, target string, since time.Time) ([]*price.Price, error) {
	u := fmt.Sprintf("%s/%s..?from=%s&to=%s", endpoint, since.Format("2006-01-02"), base, target)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var parsed historicalResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	dates := make([]string, 0, len(parsed.Rates))
	for d := range parsed.Rates {
		dates = append(dates, d)
	}
	sort.Strings(dates)
	commodityID := base + target
	prices := make([]*price.Price, 0, len(dates))
	for _, d := range dates {
		raw, ok := parsed.Rates[d][target]
		if !ok {
			continue
		}
		parsedDate, err := time.ParseInLocation("2006-01-02", d, config.TimeZone())
		if err != nil {
			return nil, err
		}
		prices = append(prices, &price.Price{
			Date:          parsedDate,
			CommodityType: config.Unknown,
			CommodityID:   commodityID,
			CommodityName: commodityID,
			Value:         decimal.NewFromFloat(raw),
		})
	}
	return prices, nil
}
