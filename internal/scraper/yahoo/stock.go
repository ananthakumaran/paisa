package yahoo

import (
	"fmt"
	"time"

	"github.com/google/btree"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/utils"
)

// exchangePoint is a sortable rate point used to look up an FX rate for a
// particular trading-day timestamp.
type exchangePoint struct {
	Timestamp int64
	Close     float64
}

func (p exchangePoint) Less(o btree.Item) bool {
	return p.Timestamp < (o.(exchangePoint).Timestamp)
}

// GetHistory fetches the daily price history for a stock / ETF / index from
// Yahoo Finance and converts the values into the user's default currency
// (when needed) using Yahoo FX rates.
//
// The function delegates the actual HTTP work to DefaultClient(); tests
// inject a custom client via getHistoryWithClient.
func GetHistory(ticker string, commodityName string) ([]*price.Price, error) {
	return getHistoryWithClient(DefaultClient(), ticker, commodityName)
}

func getHistoryWithClient(c *Client, ticker string, commodityName string) ([]*price.Price, error) {
	log.Infof("yahoo: fetching stock history for %s", ticker)
	resp, err := c.FetchChart(ticker)
	if err != nil {
		return nil, err
	}
	if len(resp.Chart.Result) == 0 {
		if resp.Chart.Error != nil {
			return nil, fmt.Errorf("yahoo: %s: %s", resp.Chart.Error.Code, resp.Chart.Error.Description)
		}
		return nil, fmt.Errorf("yahoo: empty result for %s", ticker)
	}

	result := resp.Chart.Result[0]
	needExchangePrice := !utils.IsCurrency(result.Meta.Currency) && result.Meta.Currency != ""

	var exchangeTree *btree.BTree
	if needExchangePrice {
		fxSymbol := fmt.Sprintf("%s%s=X", result.Meta.Currency, config.DefaultCurrency())
		fxResp, fxErr := c.FetchChart(fxSymbol)
		if fxErr != nil {
			return nil, fmt.Errorf("yahoo: fetch fx %s: %w", fxSymbol, fxErr)
		}
		if len(fxResp.Chart.Result) == 0 {
			return nil, fmt.Errorf("yahoo: empty fx result for %s", fxSymbol)
		}
		fxResult := fxResp.Chart.Result[0]
		exchangeTree = btree.New(2)
		for i, t := range fxResult.Timestamp {
			if i >= len(fxResult.Indicators.Quote[0].Close) {
				break
			}
			exchangeTree.ReplaceOrInsert(exchangePoint{Timestamp: t, Close: fxResult.Indicators.Quote[0].Close[i]})
		}
	}

	var prices []*price.Price
	if len(result.Indicators.Quote) == 0 {
		return prices, nil
	}
	closes := result.Indicators.Quote[0].Close
	for i, ts := range result.Timestamp {
		if i >= len(closes) {
			break
		}
		value := closes[i]
		// Yahoo emits null in the closes array when a holiday lands inside a
		// session window; the JSON decoder turns these into the zero value, so
		// we skip them rather than emit a 0-priced point.
		if value == 0 {
			continue
		}

		if needExchangePrice && exchangeTree != nil {
			fx := utils.BTreeDescendFirstLessOrEqual(exchangeTree, exchangePoint{Timestamp: ts})
			if fx.Close > 0 {
				value = value * fx.Close
			}
		}

		date := time.Unix(ts, 0)
		prices = append(prices, &price.Price{
			Date:          date,
			CommodityType: config.Stock,
			CommodityID:   ticker,
			CommodityName: commodityName,
			Value:         decimal.NewFromFloat(value),
		})
	}
	return prices, nil
}
