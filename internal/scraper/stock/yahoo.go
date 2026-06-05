package stock

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/google/btree"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/utils"
)

var UserAgents = []string{
	// Chrome
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",

	// # Firefox
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.7; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (X11; Linux i686; rv:135.0) Gecko/20100101 Firefox/135.0",

	// # Safari
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_7_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Safari/605.1.15",

	// # Edge
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/131.0.2903.86",
}

type UserAgent struct {
	sync.Once
	name string
}

var agent UserAgent

func selectAgent() {
	agent.name = UserAgents[rand.Intn(len(UserAgents))]
}

type Quote struct {
	Close []float64
}

type Indicators struct {
	Quote []Quote
}

type Meta struct {
	Currency string
}

type Result struct {
	Timestamp  []int64
	Indicators Indicators
	Meta       Meta
}

type Chart struct {
	Result []Result
}
type Response struct {
	Chart Chart
}

type ExchangePrice struct {
	Timestamp int64
	Close     float64
}

func (p ExchangePrice) Less(o btree.Item) bool {
	return p.Timestamp < (o.(ExchangePrice).Timestamp)
}

func normalizeYahooCurrency(currency string) (string, float64) {
	switch currency {
	case "GBp", "GBX":
		return "GBP", 0.01
	default:
		return currency, 1
	}
}

func normalizeYahooPrice(value float64, currency string) (float64, string) {
	currency, scale := normalizeYahooCurrency(currency)
	return value * scale, currency
}

func exchangeRateAt(exchangePrice *btree.BTree, timestamp int64) (float64, error) {
	if exchangePrice == nil {
		return 0, fmt.Errorf("exchange price not found for timestamp %d", timestamp)
	}

	price := utils.BTreeDescendFirstLessOrEqual(exchangePrice, ExchangePrice{Timestamp: timestamp})
	if price.Timestamp == 0 || price.Close == 0 {
		return 0, fmt.Errorf("exchange price not found for timestamp %d", timestamp)
	}

	return price.Close, nil
}

func GetHistory(ticker string, commodityName string) ([]*price.Price, error) {
	log.Info("Fetching stock price history from Yahoo")
	response, err := getTicker(ticker)
	if err != nil {
		return nil, err
	}

	var prices []*price.Price
	if len(response.Chart.Result) == 0 {
		return nil, fmt.Errorf("missing yahoo chart result")
	}

	result := response.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("missing yahoo quote data")
	}

	quoteCurrency, _ := normalizeYahooCurrency(result.Meta.Currency)
	needExchangePrice := false
	var exchangePrice *btree.BTree

	if !utils.IsCurrency(quoteCurrency) {
		needExchangePrice = true
		exchangeResponse, err := getTicker(fmt.Sprintf("%s%s=X", quoteCurrency, config.DefaultCurrency()))
		if err != nil {
			return nil, err
		}

		if len(exchangeResponse.Chart.Result) == 0 {
			return nil, fmt.Errorf("missing yahoo exchange chart result")
		}

		exchangeResult := exchangeResponse.Chart.Result[0]
		if len(exchangeResult.Indicators.Quote) == 0 {
			return nil, fmt.Errorf("missing yahoo exchange quote data")
		}

		exchangePrice = btree.New(2)
		exchangeCloses := exchangeResult.Indicators.Quote[0].Close
		for i, t := range exchangeResult.Timestamp {
			if i >= len(exchangeCloses) {
				return nil, fmt.Errorf("missing yahoo exchange close price for timestamp %d", t)
			}

			close := exchangeCloses[i]
			if close == 0 {
				continue
			}
			exchangePrice.ReplaceOrInsert(ExchangePrice{Timestamp: t, Close: close})
		}
	}

	closes := result.Indicators.Quote[0].Close
	for i, timestamp := range result.Timestamp {
		if i >= len(closes) {
			return nil, fmt.Errorf("missing yahoo close price for timestamp %d", timestamp)
		}

		date := time.Unix(timestamp, 0)
		value, _ := normalizeYahooPrice(closes[i], result.Meta.Currency)

		if needExchangePrice {
			rate, err := exchangeRateAt(exchangePrice, timestamp)
			if err != nil {
				return nil, err
			}
			value = value * rate
		}

		price := price.Price{Date: date, CommodityType: config.Stock, CommodityID: ticker, CommodityName: commodityName, Value: decimal.NewFromFloat(value)}

		prices = append(prices, &price)
	}
	return prices, nil
}

func getTicker(ticker string) (*Response, error) {
	url := fmt.Sprintf("https://query2.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=50y", ticker)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	agent.Do(func() { selectAgent() })
	req.Header.Add("User-Agent", agent.name)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response Response
	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

type YahooPriceProvider struct {
}

func (p *YahooPriceProvider) Code() string {
	return "com-yahoo"
}

func (p *YahooPriceProvider) Label() string {
	return "Yahoo Finance"
}

func (p *YahooPriceProvider) Description() string {
	return "Supports a large set of stocks, ETFs, mutual funds, currencies, bonds, commodities, and cryptocurrencies. The stock price will be automatically converted to your default currency using the yahoo exchange rate."
}

func (p *YahooPriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "Ticker", ID: "ticker", Help: "Stock ticker symbol, can be located on Yahoo's website. For example, AAPL is the ticker symbol for Apple Inc. (AAPL)", InputType: "text"},
	}
}

func (p *YahooPriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	return []price.AutoCompleteItem{}
}

func (p *YahooPriceProvider) ClearCache(db *gorm.DB) {
}

func (p *YahooPriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	return GetHistory(code, commodityName)
}
