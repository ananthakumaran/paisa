package stock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var finbodhiClient = &http.Client{Timeout: 30 * time.Second}

type finbodhiSecurity struct {
	ISIN     string `json:"isin"`
	Name     string `json:"name"`
	Ticker   string `json:"ticker"`
	Exchange string `json:"exchange"`
}

func queryFinBodhi(sql string, params url.Values, result interface{}) error {
	params.Set("default_format", "JSON")
	params.Set("query", sql)
	req, err := http.NewRequest(http.MethodGet, "https://mutualfund.finbodhi.com?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth("play", "")

	resp, err := finbodhiClient.Do(req)
	if err != nil {
		return fmt.Errorf("FinBodhi stock request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("FinBodhi stock: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode FinBodhi stock response: %w", err)
	}
	return nil
}

func searchFinBodhi(search string) ([]finbodhiSecurity, error) {
	search = strings.TrimSpace(search)
	sql := `select isin, name, ticker, exchange from security
		where notEmpty(trimBoth(name))
		and exchange in ('bse', 'nse', 'gift', 'NMS', 'NYQ', 'PCX', 'BTS', 'ASE')
		and type in ('equity', 'etf')
		and isin in (select isin from isins)`
	if search != "" {
		sql += " and (name ilike {pattern:String} or ticker ilike {pattern:String} or isin = {search:String})"
	}
	sql += " order by name, isin, exchange limit 200"
	var result struct {
		Data []finbodhiSecurity `json:"data"`
	}
	err := queryFinBodhi(sql, url.Values{"param_search": {search}, "param_pattern": {"%" + search + "%"}}, &result)
	return result.Data, err
}

func getFinBodhiHistory(isin, commodityName string) ([]*price.Price, error) {
	if strings.TrimSpace(isin) == "" {
		return nil, fmt.Errorf("stock ISIN is required")
	}
	log.Info("Fetching stock price history from FinBodhi")
	var result struct {
		Data []struct {
			Date  string           `json:"date"`
			Close *decimal.Decimal `json:"close"`
		} `json:"data"`
	}
	err := queryFinBodhi(
		"select toDate(date) as date, close from price where isin = {isin:String} order by date",
		url.Values{"param_isin": {isin}},
		&result,
	)
	if err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no price history found for ISIN %q", isin)
	}

	prices := make([]*price.Price, 0, len(result.Data))
	for _, item := range result.Data {
		date, err := time.ParseInLocation("2006-01-02", item.Date, config.TimeZone())
		if err != nil {
			return nil, fmt.Errorf("stock %s: invalid price date: %w", isin, err)
		}
		if item.Close == nil || item.Close.IsNegative() {
			return nil, fmt.Errorf("stock %s: invalid closing price on %s", isin, item.Date)
		}
		prices = append(prices, &price.Price{
			Date:          date,
			CommodityType: config.Stock,
			CommodityID:   isin,
			CommodityName: commodityName,
			Value:         *item.Close,
		})
	}
	return prices, nil
}

type FinBodhiPriceProvider struct {
}

func (p *FinBodhiPriceProvider) Code() string {
	return "com-finbodhi-stock"
}

func (p *FinBodhiPriceProvider) Label() string {
	return "FinBodhi Stocks"
}

func (p *FinBodhiPriceProvider) Description() string {
	return "Supports Indian and US stocks, ETFs, and GIFT City funds."
}

func (p *FinBodhiPriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "Stock or ETF", ID: "security", Help: "Search by name, ticker or ISIN."},
	}
}

func (p *FinBodhiPriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	completions := []price.AutoCompleteItem{}
	if field != "security" {
		return completions
	}
	securities, err := searchFinBodhi(filter["security"])
	if err != nil {
		log.Error("Failed to search FinBodhi stocks: ", err)
		return completions
	}
	for _, security := range securities {
		listing := security.Exchange
		if security.Ticker != "" {
			listing = security.Ticker + ", " + listing
		}
		completions = append(completions, price.AutoCompleteItem{
			ID:    security.ISIN,
			Label: security.Name + " (" + listing + ") — " + security.ISIN,
		})
	}
	return completions
}

func (p *FinBodhiPriceProvider) ClearCache(db *gorm.DB) {
}

func (p *FinBodhiPriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	return getFinBodhiHistory(code, commodityName)
}
