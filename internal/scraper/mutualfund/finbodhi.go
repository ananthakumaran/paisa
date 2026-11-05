package mutualfund

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

type finbodhiScheme struct {
	Code string `json:"code"`
	Name string `json:"name"`
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
		return fmt.Errorf("FinBodhi mutual fund request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("FinBodhi mutual fund: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode FinBodhi mutual fund response: %w", err)
	}
	return nil
}

func searchFinBodhi(search string) ([]finbodhiScheme, error) {
	search = strings.TrimSpace(search)
	sql := `select nav_name as name, toString(code) as code,
		length(arrayIntersect(splitByNonAlpha(lower(nav_name)), splitByNonAlpha(lower({search:String})))) as matched_tokens
		from scheme`
	if search != "" {
		sql += " where hasAnyTokens(lower(nav_name), lower({search:String})) or toString(code) = {search:String}"
	}
	sql += " order by matched_tokens desc, nav_name, code limit 200"
	var result struct {
		Data []finbodhiScheme `json:"data"`
	}
	err := queryFinBodhi(sql, url.Values{"param_search": {search}}, &result)
	return result.Data, err
}

func getFinBodhiNav(schemeCode, commodityName string) ([]*price.Price, error) {
	if strings.TrimSpace(schemeCode) == "" {
		return nil, fmt.Errorf("mutual fund scheme code is required")
	}
	log.Info("Fetching mutual fund NAV history from FinBodhi")
	var result struct {
		Data []struct {
			Date string           `json:"date"`
			NAV  *decimal.Decimal `json:"nav"`
		} `json:"data"`
	}
	err := queryFinBodhi(
		"select date, nav from nav where scheme_code = {scheme_code:String} order by date",
		url.Values{"param_scheme_code": {schemeCode}},
		&result,
	)
	if err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no NAV history found for mutual fund scheme %q", schemeCode)
	}

	prices := make([]*price.Price, 0, len(result.Data))
	for _, item := range result.Data {
		date, err := time.ParseInLocation("2006-01-02", item.Date, config.TimeZone())
		if err != nil {
			return nil, fmt.Errorf("mutual fund %s: invalid NAV date: %w", schemeCode, err)
		}
		if item.NAV == nil || !item.NAV.IsPositive() {
			return nil, fmt.Errorf("mutual fund %s: invalid NAV on %s", schemeCode, item.Date)
		}
		prices = append(prices, &price.Price{
			Date:          date,
			CommodityType: config.MutualFund,
			CommodityID:   schemeCode,
			CommodityName: commodityName,
			Value:         *item.NAV,
		})
	}
	return prices, nil
}

type FinBodhiPriceProvider struct {
}

func (p *FinBodhiPriceProvider) Code() string {
	return "com-finbodhi-mutualfund"
}

func (p *FinBodhiPriceProvider) Label() string {
	return "FinBodhi Mutual Funds India"
}

func (p *FinBodhiPriceProvider) Description() string {
	return "Supports mutual funds in India."
}

func (p *FinBodhiPriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "Fund scheme", ID: "scheme", Help: "Search by fund name or scheme code."},
	}
}

func (p *FinBodhiPriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	completions := []price.AutoCompleteItem{}
	if field != "scheme" {
		return completions
	}
	schemes, err := searchFinBodhi(filter["scheme"])
	if err != nil {
		log.Error("Failed to search FinBodhi mutual funds: ", err)
		return completions
	}
	for _, scheme := range schemes {
		completions = append(completions, price.AutoCompleteItem{
			ID:    scheme.Code,
			Label: scheme.Name + " — " + scheme.Code,
		})
	}
	return completions
}

func (p *FinBodhiPriceProvider) ClearCache(db *gorm.DB) {
}

func (p *FinBodhiPriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	return getFinBodhiNav(code, commodityName)
}
