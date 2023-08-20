package server

import (
	"github.com/ananthakumaran/paisa/internal/config"
	c "github.com/ananthakumaran/paisa/internal/model/commodity"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/taxation"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PostingPair struct {
	Purchase posting.Posting `json:"purchase"`
	Sell     posting.Posting `json:"sell"`
	Tax      taxation.Tax    `json:"tax"`
}

type FYCapitalGain struct {
	Units         decimal.Decimal `json:"units"`
	PurchasePrice decimal.Decimal `json:"purchase_price"`
	SellPrice     decimal.Decimal `json:"sell_price"`
	Tax           taxation.Tax    `json:"tax"`
	PostingPairs  []PostingPair   `json:"posting_pairs"`
}

type CapitalGain struct {
	Account     string                   `json:"account"`
	TaxCategory string                   `json:"tax_category"`
	FY          map[string]FYCapitalGain `json:"fy"`
}

func GetCapitalGains(db *gorm.DB) gin.H {
	commodities := lo.Filter(c.All(), func(c config.Commodity, _ int) bool {
		return (c.Type == config.MutualFund || c.Type == config.Stock) &&
			(c.TaxCategory == config.Debt || c.TaxCategory == config.Equity || c.TaxCategory == config.Equity65 || c.TaxCategory == config.Equity35 || c.TaxCategory == config.UnlistedEquity)
	})
	postings := query.Init(db).Like("Assets:%").Commodities(commodities).All()
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	capitalGains := lo.MapValues(byAccount, func(postings []posting.Posting, account string) CapitalGain {
		return computeCapitalGains(db, account, c.FindByName(postings[0].Commodity), postings)
	})
	return gin.H{"capital_gains": capitalGains}
}

func computeCapitalGains(db *gorm.DB, account string, commodity config.Commodity, postings []posting.Posting) CapitalGain {
	capitalGain := CapitalGain{Account: account, TaxCategory: string(commodity.TaxCategory), FY: make(map[string]FYCapitalGain)}
	var available []posting.Posting
	for _, p := range postings {
		if p.Quantity.GreaterThan(decimal.Zero) {
			available = append(available, p)
		} else {
			quantity := p.Quantity.Neg()
			totalTax := taxation.Tax{}
			purchasePrice := decimal.Zero
			postingPairs := make([]PostingPair, 0)
			for quantity.GreaterThan(decimal.Zero) && len(available) > 0 {
				first := available[0]
				q := decimal.Zero

				if first.Quantity.GreaterThan(quantity) {
					first.AddQuantity(quantity.Neg())
					q = quantity
					available[0] = first
					quantity = decimal.Zero
				} else {
					quantity = quantity.Sub(first.Quantity)
					q = first.Quantity
					available = available[1:]
				}

				purchasePrice = purchasePrice.Add(q.Mul(first.Price()))
				tax := taxation.Calculate(db, q, commodity, first.Price(), first.Date, p.Price(), p.Date)
				totalTax = taxation.Add(totalTax, tax)
				postingPair := PostingPair{Purchase: first.WithQuantity(q), Sell: p.WithQuantity(q.Neg()), Tax: tax}
				postingPairs = append(postingPairs, postingPair)

			}
			fy := utils.FY(p.Date)
			fyCapitalGain := capitalGain.FY[fy]
			fyCapitalGain.Tax = taxation.Add(fyCapitalGain.Tax, totalTax)
			fyCapitalGain.Units.Add(p.Quantity.Neg())
			fyCapitalGain.PurchasePrice.Add(purchasePrice)
			fyCapitalGain.SellPrice.Add(p.Amount.Neg())
			fyCapitalGain.PostingPairs = append(fyCapitalGain.PostingPairs, postingPairs...)

			capitalGain.FY[fy] = fyCapitalGain

		}
	}

	return capitalGain
}
