// Package yahoo hosts Yahoo Finance-backed scrapers. This file holds the FX
// rate provider only; stock-quote support lives in
// `internal/scraper/stock/yahoo.go` and will continue to do so for the
// foreseeable future (M2-D owns the stock side).
package yahoo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/fx"
	"github.com/ananthakumaran/paisa/internal/model/price"
)

// init wires the M1-F FX subsystem into M2-D's stock scraper and replaces
// M2-D's own FX-history hook so that any commodity registered under
// provider "yahoo" with an FX-style symbol (e.g. USDCNY=X) is resolved
// through the configured FX provider chain rather than only Yahoo's chart
// endpoint.
//
// There are two hooks here because M2-D shipped two integration points:
//
//  1. fxLookupFn (per-timestamp lookup used by stock.go to convert each
//     foreign-currency stock quote into the user's default currency).
//  2. getFXPricesFn (full-history fetcher used by yahoo.PriceProvider
//     when GetPrices receives a FX-style symbol).
//
// Neither hook blocks on a network fetch in the common case: by the time
// the stock scraper runs in a server, the RateStore has been seeded from
// the prices table (see internal/server/networth.go::loadFxRatesFromDB)
// and from the FxProviders chain (see chooseFXPrices below). If the
// RateStore is empty (fresh `paisa update` before any FX commodity has
// been synced), both hooks return ok=false / fall back to Yahoo.
func init() {
	SetFXLookupFn(func(base, target string, asOf time.Time) (float64, bool) {
		store := fx.Store()
		if store == nil {
			return 0, false
		}
		rate, ok := store.Lookup(base, target, asOf)
		if !ok {
			return 0, false
		}
		v, _ := rate.Float64()
		if v == 0 {
			return 0, false
		}
		return v, true
	})

	// Replace yahoo.go's package-level getFXPricesFn with one that goes
	// through the configured FX provider chain (config.FxProviders()).
	// When the chain has data we use it; otherwise we fall back to the
	// original Yahoo behaviour so single-currency users see no change.
	getFXPricesFn = func(c *Client, code, commodityName string) ([]*price.Price, error) {
		if prices, ok := chooseFXPrices(code, commodityName); ok {
			return prices, nil
		}
		return getHistoryWithClient(c, code, commodityName)
	}

	// Self-register the Yahoo FX provider so the chain in chooseFXPrices
	// can find it under config.FxProviders() = [..., "yahoo-fx", ...].
	RegisterFxProvider("yahoo-fx", &FxPriceProvider{})
}

// chooseFXPrices iterates config.FxProviders() in order, returning the first
// non-empty result. We let the result through unchanged; the caller writes
// it to the prices table via the regular price.UpsertAllByTypeNameAndID
// path, so the rate store seeder picks it up on the next /api/networth
// request.
//
// Symbol shape handling: yahoo's FX symbols look like "USDCNY=X" (8 chars,
// trailing "=X"). Our FX providers expect "USDCNY" (no =X). We strip the
// suffix when present.
func chooseFXPrices(code, commodityName string) ([]*price.Price, bool) {
	pair := code
	if len(pair) == 8 && pair[6:] == "=X" {
		pair = pair[:6]
	}
	if len(pair) != 6 {
		return nil, false
	}
	providers := config.FxProviders()
	for _, providerCode := range providers {
		provider := lookupFxProvider(providerCode)
		if provider == nil {
			continue
		}
		prices, err := provider.GetPrices(pair, commodityName)
		if err != nil {
			log.Warnf("fx provider %s failed for %s: %v", providerCode, pair, err)
			continue
		}
		if len(prices) > 0 {
			log.Infof("fx provider %s returned %d points for %s", providerCode, len(prices), pair)
			return prices, true
		}
	}
	return nil, false
}

// fxProviderRegistry holds the FX-capable providers known to the yahoo
// package. We can't import the top-level scraper package (cycle), so each
// FX provider self-registers via RegisterFxProvider on init.
var (
	fxProviderRegistry   = map[string]price.PriceProvider{}
	fxProviderRegistryMu sync.Mutex
)

// RegisterFxProvider lets an FX-capable provider package register itself
// for use by the chained FX resolver. The yahoo FX provider self-registers
// in this file; the BoC provider registers via its package's init() (see
// internal/scraper/cn/boc/boc.go).
func RegisterFxProvider(code string, p price.PriceProvider) {
	fxProviderRegistryMu.Lock()
	defer fxProviderRegistryMu.Unlock()
	fxProviderRegistry[code] = p
}

func lookupFxProvider(code string) price.PriceProvider {
	fxProviderRegistryMu.Lock()
	defer fxProviderRegistryMu.Unlock()
	return fxProviderRegistry[code]
}

// FxSymbol returns Yahoo's ticker form for a base->target rate. For example,
// FxSymbol("USD", "CNY") -> "USDCNY=X".
func FxSymbol(base, target string) string {
	return fmt.Sprintf("%s%s=X", base, target)
}

// FxPriceProvider scrapes daily FX history from Yahoo's chart endpoint.
type FxPriceProvider struct{}

func (p *FxPriceProvider) Code() string  { return "yahoo-fx" }
func (p *FxPriceProvider) Label() string { return "Yahoo Finance FX" }
func (p *FxPriceProvider) Description() string {
	return "Daily foreign-exchange rates from Yahoo Finance. The currency-pair code uses the form USDCNY (no =X suffix)."
}

func (p *FxPriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "Currency Pair", ID: "pair", Help: "Source/target currency pair like USDCNY.", InputType: "text"},
	}
}

func (p *FxPriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	return []price.AutoCompleteItem{
		{Label: "USD/CNY", ID: "USDCNY"},
		{Label: "HKD/CNY", ID: "HKDCNY"},
		{Label: "EUR/CNY", ID: "EURCNY"},
		{Label: "USD/HKD", ID: "USDHKD"},
	}
}

func (p *FxPriceProvider) ClearCache(db *gorm.DB) {}

// GetPrices accepts a 6-letter currency pair like "USDCNY".
func (p *FxPriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	if len(code) != 6 {
		return nil, fmt.Errorf("yahoo fx: expected 6-letter currency pair, got %q", code)
	}
	base, target := code[:3], code[3:]
	return GetFxHistory(base, target)
}

// GetFxHistory fetches the FX time series from Yahoo for the requested pair.
func GetFxHistory(base, target string) ([]*price.Price, error) {
	log.Info("Fetching FX history from Yahoo: ", base, "->", target)
	url := fmt.Sprintf("https://query2.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=50y", FxSymbol(base, target))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Yahoo refuses to serve the chart endpoint without a real-looking
	// User-Agent. Reuse the first known-good agent from the stock package's
	// rotation list; we deliberately don't import it to keep package
	// boundaries clean.
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseFxChartResponse(body, base, target)
}

// fxChartResponse mirrors the slice of Yahoo's payload we care about. We keep
// it private and type-narrow to FX (the stock package has a wider version).
type fxChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency string `json:"currency"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Close []float64 `json:"close"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

func parseFxChartResponse(body []byte, base, target string) ([]*price.Price, error) {
	var parsed fxChartResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Chart.Result) == 0 || len(parsed.Chart.Result[0].Indicators.Quote) == 0 {
		return nil, fmt.Errorf("yahoo fx: empty result for %s->%s", base, target)
	}
	result := parsed.Chart.Result[0]
	closes := result.Indicators.Quote[0].Close
	commodityID := base + target
	prices := make([]*price.Price, 0, len(result.Timestamp))
	for i, ts := range result.Timestamp {
		if i >= len(closes) {
			break
		}
		// Yahoo emits null for non-trading days as 0; skip to avoid
		// poisoning the rate store.
		if closes[i] == 0 {
			continue
		}
		date := time.Unix(ts, 0).In(config.TimeZone())
		prices = append(prices, &price.Price{
			Date:          date,
			CommodityType: config.Unknown,
			CommodityID:   commodityID,
			CommodityName: commodityID,
			Value:         decimal.NewFromFloat(closes[i]),
		})
	}
	return prices, nil
}
