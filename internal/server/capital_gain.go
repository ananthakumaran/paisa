package server

import (
	c "github.com/ananthakumaran/paisa/internal/model/commodity"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/tax"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type FYCapitalGain struct {
	Gain          float64 `json:"gain"`
	TaxableGain   float64 `json:"taxable_gain"`
	ShortTermTax  float64 `json:"short_term_tax"`
	LongTermTax   float64 `json:"long_term_tax"`
	Units         float64 `json:"units"`
	PurchasePrice float64 `json:"purchase_price"`
	SellPrice     float64 `json:"sell_price"`
}

type CapitalGain struct {
	Account     string                   `json:"account"`
	TaxCategory string                   `json:"tax_category"`
	FY          map[string]FYCapitalGain `json:"fy"`
}

func GetCapitalGains(db *gorm.DB) gin.H {
	commodities := lo.Filter(c.All(), func(c c.Commodity, _ int) bool {
		return c.Harvest > 0
	})
	postings := query.Init(db).Like("Assets:%").Commodities(lo.Map(commodities, func(c c.Commodity, _ int) string { return c.Name })).All()
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	capitalGains := lo.MapValues(byAccount, func(postings []posting.Posting, account string) CapitalGain {
		return computeCapitalGains(db, account, c.FindByName(postings[0].Commodity), postings)
	})
	return gin.H{"capital_gains": capitalGains}
}

func computeCapitalGains(db *gorm.DB, account string, commodity c.Commodity, postings []posting.Posting) CapitalGain {
	capitalGain := CapitalGain{Account: account, TaxCategory: string(commodity.TaxCategory), FY: make(map[string]FYCapitalGain)}
	var available []posting.Posting
	for _, p := range postings {
		if p.Quantity > 0 {
			available = append(available, p)
		} else {
			quantity := -p.Quantity
			taxableGain := 0.0
			shortTermTax := 0.0
			longTermTax := 0.0
			purchasePrice := 0.0
			for quantity > 0 && len(available) > 0 {
				first := available[0]
				q := 0.0
				if first.Quantity > quantity {
					first.AddQuantity(-quantity)
					q = quantity
					available[0] = first
					quantity = 0
				} else {
					quantity -= first.Quantity
					q = first.Quantity
					available = available[1:]
				}

				purchasePrice += q * first.Price()
				t, s, l := tax.Calculate(db, q, commodity, first.Price(), first.Date, p.Price(), p.Date)
				taxableGain += t
				shortTermTax += s
				longTermTax += l

			}
			fy := utils.FY(p.Date)
			fyCapitalGain := capitalGain.FY[fy]
			fyCapitalGain.Gain += (-p.Amount - purchasePrice)
			fyCapitalGain.TaxableGain += taxableGain
			fyCapitalGain.LongTermTax += longTermTax
			fyCapitalGain.ShortTermTax += shortTermTax
			fyCapitalGain.Units += -p.Quantity
			fyCapitalGain.PurchasePrice += purchasePrice
			fyCapitalGain.SellPrice += -p.Amount

			capitalGain.FY[fy] = fyCapitalGain

		}
	}

	return capitalGain
}
