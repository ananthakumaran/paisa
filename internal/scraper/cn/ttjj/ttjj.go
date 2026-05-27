package ttjj

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
)

// PriceProvider scrapes mutual fund NAV history from 天天基金 (Eastmoney).
//
// Data source: https://fund.eastmoney.com/pingzhongdata/<code>.js — a JS
// file containing several `var Foo = ...;` declarations. The full NAV
// history is exposed as `Data_netWorthTrend` (an array of objects with
// `x` = ms timestamp and `y` = unit NAV).
type PriceProvider struct{}

func (p *PriceProvider) Code() string {
	return "cn-ttjj"
}

func (p *PriceProvider) Label() string {
	return "天天基金 (Eastmoney)"
}

func (p *PriceProvider) Description() string {
	return "Supports mainland China mutual funds via 天天基金 (Eastmoney) historical NAV."
}

func (p *PriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "Fund Code", ID: "code", Help: "Eastmoney fund code, e.g. 000311"},
	}
}

func (p *PriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	return []price.AutoCompleteItem{}
}

func (p *PriceProvider) ClearCache(db *gorm.DB) {
}

func (p *PriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	log.Info("Fetching mutual fund NAV history from 天天基金 (Eastmoney): ", code)
	url := fmt.Sprintf("https://fund.eastmoney.com/pingzhongdata/%s.js", code)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseNetWorthTrend(body, code, commodityName)
}

// netWorthTrendRegex matches `Data_netWorthTrend = [ ... ];`. The `(?s)`
// flag lets `.` span newlines (Eastmoney sometimes minifies, sometimes
// not). We grab the array body lazily so trailing JS code is ignored.
var netWorthTrendRegex = regexp.MustCompile(`(?s)Data_netWorthTrend\s*=\s*(\[.*?\])\s*;`)

// shanghaiLoc is fixed UTC+8 — Eastmoney encodes Chinese trading days as
// midnight in Asia/Shanghai. Using a fixed offset (instead of
// `time.LoadLocation("Asia/Shanghai")`) avoids depending on the host's
// tzdata being installed.
var shanghaiLoc = time.FixedZone("CST", 8*60*60)

type netWorthEntry struct {
	X int64           `json:"x"`
	Y decimal.Decimal `json:"y"`
}

func parseNetWorthTrend(body []byte, code string, commodityName string) ([]*price.Price, error) {
	match := netWorthTrendRegex.FindSubmatch(body)
	if match == nil {
		return nil, fmt.Errorf("ttjj: Data_netWorthTrend not found for code %s (delisted or invalid code?)", code)
	}

	var entries []netWorthEntry
	if err := json.Unmarshal(match[1], &entries); err != nil {
		return nil, fmt.Errorf("ttjj: failed to parse Data_netWorthTrend for %s: %w", code, err)
	}

	prices := make([]*price.Price, 0, len(entries))
	for _, e := range entries {
		// Truncate to the Shanghai calendar date so the stored date is
		// the trading day rather than a UTC midnight that drifts.
		t := time.UnixMilli(e.X).In(shanghaiLoc)
		date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, config.TimeZone())
		prices = append(prices, &price.Price{
			Date:          date,
			CommodityType: config.MutualFund,
			CommodityID:   code,
			CommodityName: commodityName,
			Value:         e.Y,
		})
	}
	return prices, nil
}
