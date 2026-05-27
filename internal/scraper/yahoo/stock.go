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

// fxLookupFn returns a single FX rate for base->target as of `asOf`, used by
// the stock scraper to convert a stock's quote currency into the user's
// default currency. The default implementation is nil (no override), in which
// case the stock scraper falls back to Yahoo's own FX chart endpoint. M1-F
// (issue #11) replaces this variable from fx.go's init() so the conversion
// goes through the configured FX provider chain (BoC -> Yahoo-fx -> USD
// pivot).
//
// Why a per-timestamp lookup rather than the [] series the rest of stock.go
// uses? Because the M1-F RateStore has effectively-continuous data (it
// re-uses the most recent prior rate), so passing each stock-quote
// timestamp through gets us a more accurate conversion than the monthly
// or weekly sampling we'd otherwise need to synthesise a tree from.
var fxLookupFn func(base, target string, asOf time.Time) (float64, bool)

// SetFXLookupFn installs a per-timestamp FX resolver. fx.go calls this from
// its init() so the stock scraper picks up the M1-F provider chain without a
// hard import dependency (which would cause a cycle: stock -> fx -> price ->
// stock via the price interface).
func SetFXLookupFn(fn func(base, target string, asOf time.Time) (float64, bool)) {
	fxLookupFn = fn
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

	base := result.Meta.Currency
	target := config.DefaultCurrency()

	// useM1FFx: when the M1-F FX subsystem has data for this pair, we skip
	// the Yahoo FX fetch entirely and resolve each stock-quote timestamp
	// through fxLookupFn below. Only fall back to Yahoo's FX series when
	// the hook is unset or returns no rate.
	useM1FFx := false
	if needExchangePrice && fxLookupFn != nil {
		if _, ok := fxLookupFn(base, target, time.Now()); ok {
			useM1FFx = true
			log.Infof("yahoo: stock fx for %s->%s sourced from M1-F provider chain", base, target)
		}
	}

	var exchangeTree *btree.BTree
	if needExchangePrice && !useM1FFx {
		fxSymbol := fmt.Sprintf("%s%s=X", base, target)
		fxResp, fxErr := c.FetchChart(fxSymbol)
		// Graceful FX degrade: if Yahoo's FX series cannot be fetched (404,
		// timeout, rate limit, empty), we log a warning and return the stock
		// prices in their original currency. M1-F's FX subsystem can perform
		// the conversion downstream; failing here would break the whole
		// commodity sync just because the FX endpoint blipped.
		switch {
		case fxErr != nil:
			log.Warnf("yahoo: fx fetch %s failed (%v); returning %s in %s",
				fxSymbol, fxErr, ticker, result.Meta.Currency)
			needExchangePrice = false
		case len(fxResp.Chart.Result) == 0:
			log.Warnf("yahoo: fx fetch %s returned empty; returning %s in %s",
				fxSymbol, ticker, result.Meta.Currency)
			needExchangePrice = false
		default:
			fxResult := fxResp.Chart.Result[0]
			if len(fxResult.Indicators.Quote) == 0 {
				log.Warnf("yahoo: fx fetch %s has no quote series; returning %s in %s",
					fxSymbol, ticker, result.Meta.Currency)
				needExchangePrice = false
			} else {
				exchangeTree = btree.New(2)
				for i, t := range fxResult.Timestamp {
					if i >= len(fxResult.Indicators.Quote[0].Close) {
						break
					}
					exchangeTree.ReplaceOrInsert(exchangePoint{Timestamp: t, Close: fxResult.Indicators.Quote[0].Close[i]})
				}
				if exchangeTree.Len() == 0 {
					log.Warnf("yahoo: fx fetch %s had no usable points; returning %s in %s",
						fxSymbol, ticker, result.Meta.Currency)
					needExchangePrice = false
					exchangeTree = nil
				}
			}
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

		if needExchangePrice && useM1FFx {
			// Per-timestamp lookup through the M1-F provider chain.
			asOf := time.Unix(ts, 0)
			if rate, ok := fxLookupFn(base, target, asOf); ok && rate > 0 {
				value = value * rate
			}
		} else if needExchangePrice && exchangeTree != nil {
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
