package tax

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/model/cii"
	c "github.com/ananthakumaran/paisa/internal/model/commodity"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"gorm.io/gorm"
)

var EQUITY_GRANDFATHER_DATE, _ = time.Parse("2006-01-02", "2018-02-01")
var CII_START_DATE, _ = time.Parse("2006-01-02", "2001-03-31")
var ONE_YEAR = time.Hour * 24 * 365
var THREE_YEAR = ONE_YEAR * 3

func Calculate(db *gorm.DB, quantity float64, commodity c.Commodity, purchasePrice float64, purchaseDate time.Time, sellPrice float64, sellDate time.Time) (float64, float64, float64, float64) {
	dateDiff := sellDate.Sub(purchaseDate)
	gain := sellPrice*quantity - purchasePrice*quantity

	if commodity.TaxCategory == c.Equity && sellDate.Before(EQUITY_GRANDFATHER_DATE) {
		return gain, 0, 0, 0
	}

	if commodity.TaxCategory == c.Equity && purchaseDate.Before(EQUITY_GRANDFATHER_DATE) {
		purchasePrice = service.GetUnitPrice(db, commodity.Name, EQUITY_GRANDFATHER_DATE).Value
	}

	if commodity.TaxCategory == c.Debt && purchaseDate.After(CII_START_DATE) && dateDiff > THREE_YEAR {
		purchasePrice = (purchasePrice * float64(cii.GetIndex(db, utils.FY(sellDate)))) / float64(cii.GetIndex(db, utils.FY(purchaseDate)))
	}

	taxable := sellPrice*quantity - purchasePrice*quantity
	shortTerm := 0.0
	longTerm := 0.0

	if commodity.TaxCategory == c.Equity {
		if dateDiff > ONE_YEAR {
			longTerm = taxable * 0.10
		} else {
			shortTerm = taxable * 0.15
		}

	}

	if commodity.TaxCategory == c.Debt {
		if dateDiff > THREE_YEAR {
			longTerm = taxable * 0.20
		} else {
			shortTerm = taxable * 0.30
		}
	}

	return gain, taxable, shortTerm, longTerm
}
