package model

import (
	"fmt"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/google/btree"
	"github.com/shopspring/decimal"
)

// priceHistory indexes one commodity's already normalized prices for this fetch.
type priceHistory struct {
	tree *btree.BTree
}

func newPriceHistory(prices []*price.Price) priceHistory {
	history := priceHistory{tree: btree.New(2)}
	for _, sample := range prices {
		history.tree.ReplaceOrInsert(*sample)
	}
	return history
}

func (h priceHistory) lookup(date time.Time) (decimal.Decimal, bool) {
	value := decimal.Zero
	found := false
	if h.tree != nil {
		h.tree.DescendLessOrEqual(price.Price{Date: date}, func(item btree.Item) bool {
			value = item.(price.Price).Value
			found = true
			return false
		})
	}
	return value, found
}

type priceCoverageGap struct {
	Count int
	From  time.Time
	To    time.Time
}

func (g priceCoverageGap) String() string {
	if g.Count == 0 {
		return ""
	}
	return fmt.Sprintf("skipped %d prices without historical conversion rates from %s to %s", g.Count, g.From.Format(time.RFC3339), g.To.Format(time.RFC3339))
}

// normalizeCommodityPrices converts raw samples without mutating provider output.
// Samples before the dependency's history are omitted and reported as a gap.
func normalizeCommodityPrices(commodity config.Commodity, raw []*price.Price, exchange priceHistory, defaultCurrency string) ([]*price.Price, priceCoverageGap, error) {
	var gap priceCoverageGap
	if len(raw) == 0 {
		return nil, gap, fmt.Errorf("%s: no prices fetched", commodity.Name)
	}

	quote := commodity.Price.QuoteCommodity(defaultCurrency)
	prices := make([]*price.Price, 0, len(raw))
	for _, sample := range raw {
		value := sample.Value
		if quote != defaultCurrency {
			rate, found := exchange.lookup(sample.Date)
			if !found {
				if gap.Count == 0 || sample.Date.Before(gap.From) {
					gap.From = sample.Date
				}
				if gap.Count == 0 || sample.Date.After(gap.To) {
					gap.To = sample.Date
				}
				gap.Count++
				continue
			}
			value = value.Mul(rate)
		}
		prices = append(prices, &price.Price{
			Date:          sample.Date,
			CommodityType: commodity.Type,
			CommodityID:   commodity.Price.Code,
			CommodityName: commodity.Name,
			Value:         value,
		})
	}
	if len(prices) == 0 {
		return nil, gap, fmt.Errorf("%s: no historical %s conversion rates available for fetched prices", commodity.Name, quote)
	}
	return prices, gap, nil
}
