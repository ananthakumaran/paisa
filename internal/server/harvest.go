package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	c "github.com/ananthakumaran/paisa/internal/model/commodity"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/taxation"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type HarvestBreakdown struct {
	Units             decimal.Decimal `json:"units"`
	PurchaseDate      time.Time       `json:"purchase_date"`
	PurchasePrice     decimal.Decimal `json:"purchase_price"`
	CurrentPrice      decimal.Decimal `json:"current_price"`
	PurchaseUnitPrice decimal.Decimal `json:"purchase_unit_price"`
	Tax               taxation.Tax    `json:"tax"`
}

type Harvestable struct {
	Account               string             `json:"account"`
	TaxCategory           string             `json:"tax_category"`
	TotalUnits            decimal.Decimal    `json:"total_units"`
	HarvestableUnits      decimal.Decimal    `json:"harvestable_units"`
	UnrealizedGain        decimal.Decimal    `json:"unrealized_gain"`
	TaxableUnrealizedGain decimal.Decimal    `json:"taxable_unrealized_gain"`
	HarvestBreakdown      []HarvestBreakdown `json:"harvest_breakdown"`
	CurrentUnitPrice      decimal.Decimal    `json:"current_unit_price"`
	CurrentUnitDate       time.Time          `json:"current_unit_date"`
}

func GetHarvest(db *gorm.DB) gin.H {
	commodities := lo.Filter(c.All(), func(c config.Commodity, _ int) bool {
		return c.Harvest > 0
	})
	postings := query.Init(db).Like("Assets:%").Commodities(commodities).All()
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	harvestables := lo.MapValues(byAccount, func(postings []posting.Posting, account string) Harvestable {
		return computeHarvestable(db, account, c.FindByName(postings[0].Commodity), postings)
	})
	return gin.H{"harvestables": harvestables}
}

func computeHarvestable(db *gorm.DB, account string, commodity config.Commodity, postings []posting.Posting) Harvestable {
	available := accounting.FIFO(postings)

	today := utils.EndOfToday()
	currentPrice := service.GetUnitPrice(db, commodity.Name, today)

	harvestable := Harvestable{Account: account, TaxCategory: string(commodity.TaxCategory), HarvestBreakdown: []HarvestBreakdown{}, CurrentUnitPrice: currentPrice.Value, CurrentUnitDate: currentPrice.Date}
	cutoff := utils.Now().AddDate(0, 0, -commodity.Harvest)
	for _, p := range available {
		harvestable.TotalUnits = harvestable.TotalUnits.Add(p.Quantity)
		if p.Date.Before(cutoff) {
			tax := taxation.Calculate(db, p.Quantity, commodity, p.Price(), p.Date, currentPrice.Value, currentPrice.Date)
			harvestable.HarvestableUnits = harvestable.HarvestableUnits.Add(p.Quantity)
			harvestable.UnrealizedGain = harvestable.UnrealizedGain.Add(tax.Gain)
			harvestable.TaxableUnrealizedGain = harvestable.TaxableUnrealizedGain.Add(tax.Taxable)
			harvestable.HarvestBreakdown = append(harvestable.HarvestBreakdown, HarvestBreakdown{
				Units:             p.Quantity,
				PurchaseDate:      p.Date,
				PurchasePrice:     p.Amount,
				CurrentPrice:      currentPrice.Value.Mul(p.Quantity),
				PurchaseUnitPrice: p.Price(),
				Tax:               tax,
			})
		}
	}
	return harvestable
}
