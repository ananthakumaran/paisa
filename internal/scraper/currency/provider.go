package currency

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
)

// BaseURL of the keyless, Frankfurter compatible ECB reference rate
// endpoint. Overridden in tests.
var BaseURL = "https://allratestoday.com/api/compat/frankfurter"

// The endpoint caps a range at 366 days, so history is fetched one
// year at a time, walking backwards until a year has no rates.
const chunkDays = 365

// ECB reference rates start on 1999-01-04.
var earliest = time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)

type PriceProvider struct {
}

func (p *PriceProvider) Code() string {
	return "com-allratestoday"
}

func (p *PriceProvider) Label() string {
	return "AllRatesToday ECB Exchange Rates"
}

func (p *PriceProvider) Description() string {
	return "Supports daily European Central Bank reference rates for 30 currencies (history from 1999, no api key). The rate is converted to your default currency."
}

func (p *PriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "Currency", ID: "currency", Help: "Currency to fetch the exchange rate for. The price is the value of 1 unit in your default currency."},
	}
}

func (p *PriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	return []price.AutoCompleteItem{
		{Label: "Australian Dollar (AUD)", ID: "AUD"},
		{Label: "Brazilian Real (BRL)", ID: "BRL"},
		{Label: "Canadian Dollar (CAD)", ID: "CAD"},
		{Label: "Swiss Franc (CHF)", ID: "CHF"},
		{Label: "Chinese Yuan (CNY)", ID: "CNY"},
		{Label: "Czech Koruna (CZK)", ID: "CZK"},
		{Label: "Danish Krone (DKK)", ID: "DKK"},
		{Label: "Euro (EUR)", ID: "EUR"},
		{Label: "British Pound (GBP)", ID: "GBP"},
		{Label: "Hong Kong Dollar (HKD)", ID: "HKD"},
		{Label: "Hungarian Forint (HUF)", ID: "HUF"},
		{Label: "Indonesian Rupiah (IDR)", ID: "IDR"},
		{Label: "Israeli New Shekel (ILS)", ID: "ILS"},
		{Label: "Indian Rupee (INR)", ID: "INR"},
		{Label: "Icelandic Krona (ISK)", ID: "ISK"},
		{Label: "Japanese Yen (JPY)", ID: "JPY"},
		{Label: "South Korean Won (KRW)", ID: "KRW"},
		{Label: "Mexican Peso (MXN)", ID: "MXN"},
		{Label: "Malaysian Ringgit (MYR)", ID: "MYR"},
		{Label: "Norwegian Krone (NOK)", ID: "NOK"},
		{Label: "New Zealand Dollar (NZD)", ID: "NZD"},
		{Label: "Philippine Peso (PHP)", ID: "PHP"},
		{Label: "Polish Zloty (PLN)", ID: "PLN"},
		{Label: "Romanian Leu (RON)", ID: "RON"},
		{Label: "Swedish Krona (SEK)", ID: "SEK"},
		{Label: "Singapore Dollar (SGD)", ID: "SGD"},
		{Label: "Thai Baht (THB)", ID: "THB"},
		{Label: "Turkish Lira (TRY)", ID: "TRY"},
		{Label: "US Dollar (USD)", ID: "USD"},
		{Label: "South African Rand (ZAR)", ID: "ZAR"},
	}
}

func (p *PriceProvider) ClearCache(db *gorm.DB) {
}

func (p *PriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	return GetHistory(code, config.DefaultCurrency(), commodityName, time.Now())
}

type rangeResponse struct {
	Rates map[string]map[string]decimal.Decimal `json:"rates"`
}

// GetHistory returns the daily price of 1 unit of currency `from`
// expressed in `to`, up to `until`.
func GetHistory(from, to, commodityName string, until time.Time) ([]*price.Price, error) {
	from = strings.ToUpper(strings.TrimSpace(from))
	to = strings.ToUpper(strings.TrimSpace(to))
	if from == "" || to == "" {
		return nil, fmt.Errorf("Invalid currency pair: %q/%q", from, to)
	}
	if from == to {
		return nil, fmt.Errorf("Currency %s is the same as the default currency", from)
	}

	log.Infof("Fetching %s/%s exchange rate history from AllRatesToday", from, to)

	var prices []*price.Price
	end := until
	for !end.Before(earliest) {
		start := end.AddDate(0, 0, -chunkDays)
		if start.Before(earliest) {
			start = earliest
		}

		chunk, err := fetchRange(from, to, commodityName, start, end)
		if err != nil {
			return nil, err
		}
		if len(chunk) == 0 {
			break
		}
		prices = append(prices, chunk...)
		end = start.AddDate(0, 0, -1)
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("No exchange rate found for %s/%s", from, to)
	}

	sort.Slice(prices, func(i, j int) bool { return prices[i].Date.Before(prices[j].Date) })
	return prices, nil
}

func fetchRange(from, to, commodityName string, start, end time.Time) ([]*price.Price, error) {
	url := fmt.Sprintf("%s/%s..%s?base=%s&symbols=%s", BaseURL, start.Format("2006-01-02"), end.Format("2006-01-02"), from, to)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "paisa")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 404 means no published rates in the range.
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code: %d, body: %s", resp.StatusCode, string(respBytes))
	}

	var response rangeResponse
	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		return nil, err
	}

	var prices []*price.Price
	for date, rates := range response.Rates {
		value, ok := rates[to]
		if !ok || value.IsZero() {
			continue
		}

		dateTime, err := time.ParseInLocation("2006-01-02", date, config.TimeZone())
		if err != nil {
			return nil, err
		}

		prices = append(prices, &price.Price{
			Date:          dateTime,
			CommodityType: config.Unknown,
			CommodityID:   from,
			CommodityName: commodityName,
			Value:         value,
		})
	}
	return prices, nil
}
