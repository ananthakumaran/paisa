package stock

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/btree"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/utils"
)

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

func GetHistory(ticker string, commodityName string) ([]*price.Price, error) {
	log.Info("Fetching stock price history from Yahoo")
	response, err := getTicker(ticker)
	if err != nil {
		return nil, err
	}

	var prices []*price.Price
	result := response.Chart.Result[0]
	needExchangePrice := false
	var exchangePrice *btree.BTree

	if !utils.IsCurrency(result.Meta.Currency) {
		needExchangePrice = true
		exchangeResponse, err := getTicker(fmt.Sprintf("%s%s=X", result.Meta.Currency, config.DefaultCurrency()))
		if err != nil {
			return nil, err
		}

		exchangeResult := exchangeResponse.Chart.Result[0]

		exchangePrice = btree.New(2)
		for i, t := range exchangeResult.Timestamp {
			exchangePrice.ReplaceOrInsert(ExchangePrice{Timestamp: t, Close: exchangeResult.Indicators.Quote[0].Close[i]})
		}
	}

	for i, timestamp := range result.Timestamp {
		date := time.Unix(timestamp, 0)
		value := result.Indicators.Quote[0].Close[i]

		if needExchangePrice {
			exchangePrice := utils.BTreeDescendFirstLessOrEqual(exchangePrice, ExchangePrice{Timestamp: timestamp})
			value = value * exchangePrice.Close
		}

		price := price.Price{Date: date, CommodityType: config.Stock, CommodityID: ticker, CommodityName: commodityName, Value: decimal.NewFromFloat(value)}

		prices = append(prices, &price)
	}
	return prices, nil
}

func getTicker(ticker string) (*Response, error) {
	url := fmt.Sprintf("https://query2.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=50y", ticker)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
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
