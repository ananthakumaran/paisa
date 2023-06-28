package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/accounting"
	c "github.com/ananthakumaran/paisa/internal/model/commodity"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/taxation"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type HarvestBreakdown struct {
	Units             float64      `json:"units"`
	PurchaseDate      time.Time    `json:"purchase_date"`
	PurchasePrice     float64      `json:"purchase_price"`
	CurrentPrice      float64      `json:"current_price"`
	PurchaseUnitPrice float64      `json:"purchase_unit_price"`
	Tax               taxation.Tax `json:"tax"`
}

type Harvestable struct {
	Account               string             `json:"account"`
	TaxCategory           string             `json:"tax_category"`
	TotalUnits            float64            `json:"total_units"`
	HarvestableUnits      float64            `json:"harvestable_units"`
	UnrealizedGain        float64            `json:"unrealized_gain"`
	TaxableUnrealizedGain float64            `json:"taxable_unrealized_gain"`
	HarvestBreakdown      []HarvestBreakdown `json:"harvest_breakdown"`
	CurrentUnitPrice      float64            `json:"current_unit_price"`
	CurrentUnitDate       time.Time          `json:"current_unit_date"`
}

func GetHarvest(db *gorm.DB) gin.H {
	commodities := lo.Filter(c.All(), func(c c.Commodity, _ int) bool {
		return c.Harvest > 0
	})
	postings := query.Init(db).Like("Assets:%").Commodities(commodities).All()
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	harvestables := lo.MapValues(byAccount, func(postings []posting.Posting, account string) Harvestable {
		return computeHarvestable(db, account, c.FindByName(postings[0].Commodity), postings)
	})
	return gin.H{"harvestables": harvestables}
}

func computeHarvestable(db *gorm.DB, account string, commodity c.Commodity, postings []posting.Posting) Harvestable {
	available := accounting.FIFO(postings)

	today := time.Now()
	currentPrice := service.GetUnitPrice(db, commodity.Name, today)

	harvestable := Harvestable{Account: account, TaxCategory: string(commodity.TaxCategory), HarvestBreakdown: []HarvestBreakdown{}, CurrentUnitPrice: currentPrice.Value, CurrentUnitDate: currentPrice.Date}
	cutoff := time.Now().AddDate(0, 0, -commodity.Harvest)
	for _, p := range available {
		harvestable.TotalUnits += p.Quantity
		if p.Date.Before(cutoff) {
			tax := taxation.Calculate(db, p.Quantity, commodity, p.Price(), p.Date, currentPrice.Value, currentPrice.Date)
			harvestable.HarvestableUnits += p.Quantity
			harvestable.UnrealizedGain += tax.Gain
			harvestable.TaxableUnrealizedGain += tax.Taxable
			harvestable.HarvestBreakdown = append(harvestable.HarvestBreakdown, HarvestBreakdown{
				Units:             p.Quantity,
				PurchaseDate:      p.Date,
				PurchasePrice:     p.Amount,
				CurrentPrice:      currentPrice.Value * p.Quantity,
				PurchaseUnitPrice: p.Price(),
				Tax:               tax,
			})
		}
	}
	return harvestable
}
