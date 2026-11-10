package server

import (
	"github.com/ananthakumaran/paisa/internal/accounting"
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
	_, sales := accounting.TaxFIFO(db, postings)
	for _, sale := range sales {
		p := sale.Sell
		totalTax := taxation.Tax{}
		purchasePrice := decimal.Zero
		postingPairs := make([]PostingPair, 0)
		for _, purchase := range sale.Purchases {
			q := purchase.Quantity
			purchasePrice = purchasePrice.Add(purchase.Amount)
			tax := taxation.Calculate(db, q, commodity, purchase.Price(), purchase.Date, p.Price(), p.Date)
			totalTax = taxation.Add(totalTax, tax)
			postingPair := PostingPair{Purchase: purchase, Sell: p.WithQuantity(q.Neg()), Tax: tax}
			postingPairs = append(postingPairs, postingPair)
		}
		fy := utils.FY(p.Date)
		fyCapitalGain := capitalGain.FY[fy]
		fyCapitalGain.Tax = taxation.Add(fyCapitalGain.Tax, totalTax)
		fyCapitalGain.Units = fyCapitalGain.Units.Add(p.Quantity.Neg())
		fyCapitalGain.PurchasePrice = fyCapitalGain.PurchasePrice.Add(purchasePrice)
		fyCapitalGain.SellPrice = fyCapitalGain.SellPrice.Add(p.Amount.Neg())
		fyCapitalGain.PostingPairs = append(fyCapitalGain.PostingPairs, postingPairs...)

		capitalGain.FY[fy] = fyCapitalGain
	}

	return capitalGain
}
