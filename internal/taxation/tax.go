package taxation

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/model/cii"
	c "github.com/ananthakumaran/paisa/internal/model/commodity"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"gorm.io/gorm"
)

var EQUITY_GRANDFATHER_DATE, _ = time.Parse("2006-01-02", "2018-02-01")
var DEBT_INDEXATION_REVOCATION_DATE, _ = time.Parse("2006-01-02", "2023-04-01")
var CII_START_DATE, _ = time.Parse("2006-01-02", "2001-03-31")
var ONE_YEAR = time.Hour * 24 * 365
var THREE_YEAR = ONE_YEAR * 3
var TWO_YEAR = ONE_YEAR * 2

type Tax struct {
	Gain      float64 `json:"gain"`
	Taxable   float64 `json:"taxable"`
	Slab      float64 `json:"slab"`
	LongTerm  float64 `json:"long_term"`
	ShortTerm float64 `json:"short_term"`
}

func Add(a, b Tax) Tax {
	return Tax{Gain: a.Gain + b.Gain, Taxable: a.Taxable + b.Taxable, LongTerm: a.LongTerm + b.LongTerm, ShortTerm: a.ShortTerm + b.ShortTerm, Slab: a.Slab + b.Slab}
}

func Calculate(db *gorm.DB, quantity float64, commodity c.Commodity, purchasePrice float64, purchaseDate time.Time, sellPrice float64, sellDate time.Time) Tax {
	dateDiff := sellDate.Sub(purchaseDate)
	gain := sellPrice*quantity - purchasePrice*quantity

	if (commodity.TaxCategory == c.Equity || commodity.TaxCategory == c.Equity65) && sellDate.Before(EQUITY_GRANDFATHER_DATE) {
		return Tax{Gain: gain, Taxable: 0, ShortTerm: 0, LongTerm: 0, Slab: 0}
	}

	if (commodity.TaxCategory == c.Equity || commodity.TaxCategory == c.Equity65) && purchaseDate.Before(EQUITY_GRANDFATHER_DATE) {
		purchasePrice = service.GetUnitPrice(db, commodity.Name, EQUITY_GRANDFATHER_DATE).Value
	}

	if commodity.TaxCategory == c.Debt && purchaseDate.After(CII_START_DATE) && dateDiff > THREE_YEAR {
		purchasePrice = (purchasePrice * float64(cii.GetIndex(db, utils.FY(sellDate)))) / float64(cii.GetIndex(db, utils.FY(purchaseDate)))
	}

	if commodity.TaxCategory == c.UnlistedEquity && purchaseDate.After(CII_START_DATE) && dateDiff > TWO_YEAR {
		purchasePrice = (purchasePrice * float64(cii.GetIndex(db, utils.FY(sellDate)))) / float64(cii.GetIndex(db, utils.FY(purchaseDate)))
	}

	taxable := sellPrice*quantity - purchasePrice*quantity
	shortTerm := 0.0
	longTerm := 0.0
	slab := 0.0

	if commodity.TaxCategory == c.Equity || commodity.TaxCategory == c.Equity65 {
		if dateDiff > ONE_YEAR {
			longTerm = taxable * 0.10
		} else {
			shortTerm = taxable * 0.15
		}

	}

	if commodity.TaxCategory == c.Debt {
		if dateDiff > THREE_YEAR && purchaseDate.Before(DEBT_INDEXATION_REVOCATION_DATE) {
			longTerm = taxable * 0.20
		} else {
			slab = taxable
		}
	}

	if commodity.TaxCategory == c.Equity35 {
		if dateDiff > THREE_YEAR {
			longTerm = taxable * 0.20
		} else {
			slab = taxable
		}
	}

	if commodity.TaxCategory == c.UnlistedEquity {
		if dateDiff > TWO_YEAR {
			longTerm = taxable * 0.20
		} else {
			slab = taxable
		}
	}

	return Tax{Gain: gain, Taxable: taxable, ShortTerm: shortTerm, LongTerm: longTerm, Slab: slab}
}
