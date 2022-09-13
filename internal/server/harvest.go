package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/model/cii"
	c "github.com/ananthakumaran/paisa/internal/model/commodity"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type FYCapitalGain struct {
	Gain          float64 `json:"gain"`
	Units         float64 `json:"units"`
	PurchasePrice float64 `json:"purchase_price"`
	SellPrice     float64 `json:"sell_price"`
}

type HarvestBreakdown struct {
	Units                 float64   `json:"units"`
	PurchaseDate          time.Time `json:"purchase_date"`
	PurchasePrice         float64   `json:"purchase_price"`
	CurrentPrice          float64   `json:"current_price"`
	PurchaseUnitPrice     float64   `json:"purchase_unit_price"`
	UnrealizedGain        float64   `json:"unrealized_gain"`
	TaxableUnrealizedGain float64   `json:"taxable_unrealized_gain"`
}

type Harvestable struct {
	TotalUnits            float64            `json:"total_units"`
	HarvestableUnits      float64            `json:"harvestable_units"`
	UnrealizedGain        float64            `json:"unrealized_gain"`
	TaxableUnrealizedGain float64            `json:"taxable_unrealized_gain"`
	HarvestBreakdown      []HarvestBreakdown `json:"harvest_breakdown"`
	CurrentUnitPrice      float64            `json:"current_unit_price"`
	GrandfatherUnitPrice  float64            `json:"grandfather_unit_price"`
	CurrentUnitDate       time.Time          `json:"current_unit_date"`
}

type CapitalGain struct {
	Account     string                   `json:"account"`
	TaxCategory string                   `json:"tax_category"`
	FY          map[string]FYCapitalGain `json:"fy"`
	Harvestable Harvestable              `json:"harvestable"`
}

var EQUITY_GRANDFATHER_DATE, _ = time.Parse("2006-01-02", "2018-02-01")
var CII_START_DATE, _ = time.Parse("2006-01-02", "2001-03-31")

func GetHarvest(db *gorm.DB) gin.H {
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
			purchasePrice := 0.0
			for quantity > 0 && len(available) > 0 {
				first := available[0]
				if first.Quantity > quantity {
					purchasePrice += quantity * first.Price()
					first.AddQuantity(-quantity)
					quantity = 0
				} else {
					purchasePrice += quantity * first.Price()
					quantity -= first.Quantity
					available = available[1:]
				}
			}
			fy := utils.FY(p.Date)
			fyCapitalGain := capitalGain.FY[fy]
			fyCapitalGain.Gain += (-p.Amount - purchasePrice)
			fyCapitalGain.Units += -p.Quantity
			fyCapitalGain.PurchasePrice += purchasePrice
			fyCapitalGain.SellPrice += -p.Amount
			capitalGain.FY[fy] = fyCapitalGain

		}
	}

	grandfather := false
	if commodity.TaxCategory == c.Equity {
		grandfather = true
	}

	today := time.Now()
	currentPrice := service.GetUnitPrice(db, commodity.Name, today)
	grandfatherUnitPrice := 0.0
	if grandfather {
		grandfatherPrice := service.GetUnitPrice(db, commodity.Name, EQUITY_GRANDFATHER_DATE)
		grandfatherUnitPrice = grandfatherPrice.Value

	}

	harvestable := Harvestable{HarvestBreakdown: []HarvestBreakdown{}, CurrentUnitPrice: currentPrice.Value, CurrentUnitDate: currentPrice.Date, GrandfatherUnitPrice: grandfatherUnitPrice}
	cutoff := time.Now().AddDate(0, 0, -commodity.Harvest)
	for _, p := range available {
		harvestable.TotalUnits += p.Quantity
		if p.Date.Before(cutoff) {
			gain := currentPrice.Value*p.Quantity - p.Amount
			taxableGain := gain
			if grandfather && p.Date.Before(EQUITY_GRANDFATHER_DATE) {
				taxableGain = grandfatherUnitPrice*p.Quantity - p.Amount
			}

			if commodity.TaxCategory == c.Debt && p.Date.After(CII_START_DATE) {
				taxableGain = currentPrice.Value*p.Quantity - (p.Amount*float64(cii.GetIndex(db, utils.FY(today))))/float64(cii.GetIndex(db, utils.FY(p.Date)))
			}
			harvestable.HarvestableUnits += p.Quantity
			harvestable.UnrealizedGain += gain
			harvestable.TaxableUnrealizedGain += taxableGain
			harvestable.HarvestBreakdown = append(harvestable.HarvestBreakdown, HarvestBreakdown{
				Units:                 p.Quantity,
				PurchaseDate:          p.Date,
				PurchasePrice:         p.Amount,
				CurrentPrice:          currentPrice.Value * p.Quantity,
				PurchaseUnitPrice:     p.Price(),
				UnrealizedGain:        gain,
				TaxableUnrealizedGain: taxableGain,
			})
		}
	}
	capitalGain.Harvestable = harvestable
	return capitalGain
}
