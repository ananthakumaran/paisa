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
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
)

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
