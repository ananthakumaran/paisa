package sif

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
)

type Scheme struct {
	ID      string `json:"id"`
	NAVName string `json:"nav_name"`
}

var client = &http.Client{Timeout: 30 * time.Second}

func query(sql string, params url.Values, result interface{}) error {
	if params == nil {
		params = url.Values{}
	}
	params.Set("default_format", "JSON")
	params.Set("query", sql)
	req, err := http.NewRequest(http.MethodGet, "https://mutualfund.finbodhi.com?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth("play", "")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("FinBodhi SIF request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("FinBodhi SIF: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode FinBodhi SIF response: %w", err)
	}
	return nil
}

func GetSchemes(search string) ([]Scheme, error) {
	log.Info("Searching SIF schemes on FinBodhi")
	var result struct {
		Data []Scheme `json:"data"`
	}
	search = strings.TrimSpace(search)
	sql := `select toString(id) as id, nav_name,
		length(arrayIntersect(splitByNonAlpha(lower(nav_name)), splitByNonAlpha(lower({search:String})))) as matched_tokens
		from sif_scheme`
	if search != "" {
		sql += " where hasAnyTokens(lower(nav_name), lower({search:String})) or id = {search:String}"
	}
	sql += " order by matched_tokens desc, nav_name, id limit 200"
	err := query(sql, url.Values{"param_search": {search}}, &result)
	if err != nil {
		return nil, err
	}
	for _, item := range result.Data {
		if item.ID == "" || item.NAVName == "" {
			return nil, fmt.Errorf("FinBodhi SIF catalogue contains a missing number or name")
		}
	}
	return result.Data, nil
}

func GetNav(schemeID string, commodityName string) ([]*price.Price, error) {
	if strings.TrimSpace(schemeID) == "" {
		return nil, fmt.Errorf("SIF number is required")
	}
	log.Info("Fetching SIF NAV history from FinBodhi")
	var result struct {
		Data []struct {
			Date string           `json:"date"`
			NAV  *decimal.Decimal `json:"nav"`
		} `json:"data"`
	}
	err := query(
		"select date, nav from sif_nav where scheme_id = {scheme_id:String} order by date",
		url.Values{"param_scheme_id": {schemeID}},
		&result,
	)
	if err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no NAV history found for SIF number %q", schemeID)
	}

	prices := make([]*price.Price, 0, len(result.Data))
	for _, item := range result.Data {
		date, err := time.ParseInLocation("2006-01-02", item.Date, config.TimeZone())
		if err != nil {
			return nil, fmt.Errorf("SIF %s: invalid NAV date: %w", schemeID, err)
		}
		if item.NAV == nil || !item.NAV.IsPositive() {
			return nil, fmt.Errorf("SIF %s: invalid NAV on %s", schemeID, item.Date)
		}
		prices = append(prices, &price.Price{
			Date:          date,
			CommodityType: config.SIF,
			CommodityID:   schemeID,
			CommodityName: commodityName,
			Value:         *item.NAV,
		})
	}
	return prices, nil
}
