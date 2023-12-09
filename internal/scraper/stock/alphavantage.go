package stock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/google/btree"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Match struct {
	Symbol   string `json:"1. symbol"`
	Name     string `json:"2. name"`
	Type     string `json:"3. type"`
	Region   string `json:"4. region"`
	Currency string `json:"8. currency"`
}

type ErrorResponse struct {
	Information string `json:"Information"`
}

type SearchResponse struct {
	BestMatches []Match `json:"bestMatches"`
}

type DailyPrice struct {
	Close string `json:"4. close"`
}

type FXSeriesDailyReponse struct {
	TimeSeriesFX map[string]DailyPrice `json:"Time Series FX (Daily)"`
}

type TimeSeriesDailyReponse struct {
	TimeSeriesDaily map[string]DailyPrice `json:"Time Series (Daily)"`
}

type AlphaVantageExchangePrice struct {
	Date  time.Time
	Close decimal.Decimal
}

func (p AlphaVantageExchangePrice) Less(o btree.Item) bool {
	return p.Date.Before(o.(AlphaVantageExchangePrice).Date)
}

func fetch[R any](url string, response *R) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Unexpected status code: %d, body: %s", resp.StatusCode, string(respBytes))
	}

	var errorResponse ErrorResponse
	err = json.Unmarshal(respBytes, &errorResponse)
	if err != nil {
		return err
	}

	if errorResponse.Information != "" {
		return fmt.Errorf("Error response: %s", errorResponse.Information)
	}

	err = json.Unmarshal(respBytes, response)
	if err != nil {
		return err
	}

	return nil
}

func getHistory(code, commodityName string) ([]*price.Price, error) {
	parts := strings.Split(code, ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("Invalid code: %s", code)
	}
	apiKey, ticker, currency := parts[0], parts[1], parts[2]

	log.Info("Fetching stock price history from Alpha Vantage")
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=%s&outputsize=full&apikey=%s", ticker, apiKey)
	var response TimeSeriesDailyReponse
	err := fetch(url, &response)
	if err != nil {
		return nil, err
	}

	var exchangePrice *btree.BTree
	needExchangePrice := false
	if !utils.IsCurrency(currency) {
		needExchangePrice = true
		log.Info("Fetching exchange rate from Alpha Vantage")
		url = fmt.Sprintf("https://www.alphavantage.co/query?function=FX_DAILY&from_symbol=%s&to_symbol=%s&outputsize=full&apikey=%s", currency, config.DefaultCurrency(), apiKey)
		var response FXSeriesDailyReponse
		err = fetch(url, &response)
		if err != nil {
			return nil, err
		}

		exchangePrice = btree.New(2)
		for date, value := range response.TimeSeriesFX {
			dateTime, err := time.Parse("2006-01-02", date)
			if err != nil {
				return nil, err
			}
			value, err := decimal.NewFromString(value.Close)
			if err != nil {
				return nil, err
			}

			exchangePrice.ReplaceOrInsert(AlphaVantageExchangePrice{Date: dateTime, Close: value})
		}
	}

	var prices []*price.Price
	for date, value := range response.TimeSeriesDaily {
		dateTime, err := time.Parse("2006-01-02", date)
		if err != nil {
			return nil, err
		}
		value, err := decimal.NewFromString(value.Close)
		if err != nil {
			return nil, err
		}

		if needExchangePrice {
			exchangePrice := utils.BTreeDescendFirstLessOrEqual(exchangePrice, AlphaVantageExchangePrice{Date: dateTime})
			value = value.Mul(exchangePrice.Close)
		}

		if value.IsZero() {
			continue
		}

		prices = append(prices, &price.Price{
			Date:          dateTime,
			CommodityType: config.Stock,
			CommodityID:   code,
			CommodityName: commodityName,
			Value:         value,
		})
	}

	return prices, nil
}

func searchTicker(apiKey, ticker string) (*SearchResponse, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=SYMBOL_SEARCH&keywords=%s&apikey=%s", ticker, apiKey)
	var response SearchResponse
	err := fetch(url, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

type AlphaVantagePriceProvider struct {
}

func (p *AlphaVantagePriceProvider) Code() string {
	return "co-alphavantage"
}

func (p *AlphaVantagePriceProvider) Label() string {
	return "Alpha Vantage"
}

func (p *AlphaVantagePriceProvider) Description() string {
	return "Supports 100,000+ stocks, ETFs, mutual funds. The stock price will be automatically converted to your default currency using the exchange rate."
}

func (p *AlphaVantagePriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "Api Key", ID: "apikey", Help: "Alpha Vantage provides <a href='https://www.alphavantage.co/support/#api-key' target='_blank'>free api key</a> with 25 requests per day limit.", InputType: "text"},
		{Label: "Ticker", ID: "ticker", Help: "Type the name and wait for the dropdown to appear. The api key is required to fetch the list of tickers."},
	}
}

func (p *AlphaVantagePriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	apiKey := filter["apikey"]
	ticker := filter["ticker"]
	if apiKey == "" || ticker == "" {
		return []price.AutoCompleteItem{}
	}

	response, err := searchTicker(apiKey, ticker)
	if err != nil {
		log.Error(err)
		return []price.AutoCompleteItem{}
	}

	return lo.Map(response.BestMatches, func(match Match, _ int) price.AutoCompleteItem {
		return price.AutoCompleteItem{
			Label: match.Name + " (" + match.Region + ", " + match.Type + ", " + match.Currency + ", " + match.Symbol + ")",
			ID:    strings.Join([]string{apiKey, match.Symbol, match.Currency}, ":"),
		}
	})

}

func (p *AlphaVantagePriceProvider) ClearCache(db *gorm.DB) {
}

func (p *AlphaVantagePriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	return getHistory(code, commodityName)
}
