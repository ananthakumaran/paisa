package taxation

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/cii"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var EQUITY_GRANDFATHER_DATE, DEBT_INDEXATION_REVOCATION_DATE, CII_START_DATE time.Time
var ONE_YEAR = time.Hour * 24 * 365
var THREE_YEAR = ONE_YEAR * 3
var TWO_YEAR = ONE_YEAR * 2

func init() {
	EQUITY_GRANDFATHER_DATE, _ = time.ParseInLocation("2006-01-02", "2018-02-01", config.TimeZone())
	DEBT_INDEXATION_REVOCATION_DATE, _ = time.ParseInLocation("2006-01-02", "2023-04-01", config.TimeZone())
	CII_START_DATE, _ = time.ParseInLocation("2006-01-02", "2001-03-31", config.TimeZone())
}

type Tax struct {
	Gain      decimal.Decimal `json:"gain"`
	Taxable   decimal.Decimal `json:"taxable"`
	Slab      decimal.Decimal `json:"slab"`
	LongTerm  decimal.Decimal `json:"long_term"`
	ShortTerm decimal.Decimal `json:"short_term"`
}

func Add(a, b Tax) Tax {
	return Tax{Gain: a.Gain.Add(b.Gain), Taxable: a.Taxable.Add(b.Taxable), LongTerm: a.LongTerm.Add(b.LongTerm), ShortTerm: a.ShortTerm.Add(b.ShortTerm), Slab: a.Slab.Add(b.Slab)}
}

func Calculate(db *gorm.DB, quantity decimal.Decimal, commodity config.Commodity, purchasePrice decimal.Decimal, purchaseDate time.Time, sellPrice decimal.Decimal, sellDate time.Time) Tax {

	dateDiff := sellDate.Sub(purchaseDate)
	gain := sellPrice.Mul(quantity).Sub(purchasePrice.Mul(quantity))

	if (commodity.TaxCategory == config.Equity || commodity.TaxCategory == config.Equity65) && sellDate.Before(EQUITY_GRANDFATHER_DATE) {
		return Tax{Gain: gain, Taxable: decimal.Zero, ShortTerm: decimal.Zero, LongTerm: decimal.Zero, Slab: decimal.Zero}
	}

	if (commodity.TaxCategory == config.Equity || commodity.TaxCategory == config.Equity65) && purchaseDate.Before(EQUITY_GRANDFATHER_DATE) {
		purchasePrice = service.GetUnitPrice(db, commodity.Name, EQUITY_GRANDFATHER_DATE).Value
	}

	if commodity.TaxCategory == config.Debt && purchaseDate.After(CII_START_DATE) && dateDiff > THREE_YEAR {
		purchasePrice = purchasePrice.Mul(decimal.NewFromInt(int64(cii.GetIndex(db, utils.FY(sellDate)))).Div(decimal.NewFromInt(int64(cii.GetIndex(db, utils.FY(purchaseDate))))))
	}

	if commodity.TaxCategory == config.UnlistedEquity && purchaseDate.After(CII_START_DATE) && dateDiff > TWO_YEAR {
		purchasePrice = purchasePrice.Mul(decimal.NewFromInt(int64(cii.GetIndex(db, utils.FY(sellDate)))).Div(decimal.NewFromInt(int64(cii.GetIndex(db, utils.FY(purchaseDate))))))
	}

	taxable := sellPrice.Mul(quantity).Sub(purchasePrice.Mul(quantity))
	shortTerm := decimal.Zero
	longTerm := decimal.Zero
	slab := decimal.Zero

	if commodity.TaxCategory == config.Equity || commodity.TaxCategory == config.Equity65 {
		if dateDiff > ONE_YEAR {
			longTerm = taxable.Mul(decimal.NewFromFloat(0.10))
		} else {
			shortTerm = taxable.Mul(decimal.NewFromFloat(0.15))
		}

	}

	if commodity.TaxCategory == config.Debt {
		if dateDiff > THREE_YEAR && purchaseDate.Before(DEBT_INDEXATION_REVOCATION_DATE) {
			longTerm = taxable.Mul(decimal.NewFromFloat(0.20))
		} else {
			slab = taxable
		}
	}

	if commodity.TaxCategory == config.Equity35 {
		if dateDiff > THREE_YEAR {
			longTerm = taxable.Mul(decimal.NewFromFloat(0.20))
		} else {
			slab = taxable
		}
	}

	if commodity.TaxCategory == config.UnlistedEquity {
		if dateDiff > TWO_YEAR {
			longTerm = taxable.Mul(decimal.NewFromFloat(0.20))
		} else {
			slab = taxable
		}
	}

	return Tax{Gain: gain, Taxable: taxable, ShortTerm: shortTerm, LongTerm: longTerm, Slab: slab}
}
