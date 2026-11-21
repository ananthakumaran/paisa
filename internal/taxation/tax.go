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

// Sections 55(2)(ac) and 112A use separate valuation, acquisition, and transfer dates.
const (
	equityGrandfatherFMVDay            = "2018-01-31"
	equityGrandfatherPurchaseCutoffDay = "2018-02-01"
	equityLTCGStartDay                 = "2018-04-01"
)

const (
	specifiedDebtFundPurchaseCutoffDay = "2023-04-01" // Section 50AA
	capitalGainsReformDay              = "2024-07-23" // Sections 48, 111A, 112 and 112A
)

var (
	equityGrandfatherFMVDate            time.Time
	equityGrandfatherPurchaseCutoffDate time.Time
	equityLTCGStartDate                 time.Time
)

var specifiedDebtFundPurchaseCutoffDate, capitalGainsReformDate, CII_START_DATE time.Time
var ONE_YEAR = time.Hour * 24 * 365

func init() {
	equityGrandfatherFMVDate, _ = time.ParseInLocation(time.DateOnly, equityGrandfatherFMVDay, config.TimeZone())
	equityGrandfatherPurchaseCutoffDate, _ = time.ParseInLocation(time.DateOnly, equityGrandfatherPurchaseCutoffDay, config.TimeZone())
	equityLTCGStartDate, _ = time.ParseInLocation(time.DateOnly, equityLTCGStartDay, config.TimeZone())
	specifiedDebtFundPurchaseCutoffDate, _ = time.Parse(time.DateOnly, specifiedDebtFundPurchaseCutoffDay)
	capitalGainsReformDate, _ = time.Parse(time.DateOnly, capitalGainsReformDay)
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

// Keep the journal's calendar date while comparing it to date-only cutoffs.
func calendarDay(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func heldMoreThanYears(purchaseDay, sellDay time.Time, years int) bool {
	anniversary := purchaseDay.AddDate(years, 0, 0)
	// Go rolls Feb 29 into March 1 in a non-leap year.
	if anniversary.Month() != purchaseDay.Month() {
		anniversary = anniversary.AddDate(0, 0, -1)
	}
	return sellDay.After(anniversary)
}

func calculateSection112(db *gorm.DB, category config.TaxCategoryType, quantity, purchasePrice, sellPrice, gain decimal.Decimal, purchaseDate, sellDate time.Time) Tax {
	purchaseDay, sellDay := calendarDay(purchaseDate), calendarDay(sellDate)
	shortTerm := Tax{Gain: gain, Taxable: gain, Slab: gain}

	// Only specified debt funds are deemed short-term under Section 50AA.
	if (category == config.Debt || category == config.ListedDebt) && !purchaseDay.Before(specifiedDebtFundPurchaseCutoffDate) {
		return shortTerm
	}

	if !sellDay.Before(capitalGainsReformDate) {
		holdingYears := 2
		if category == config.ListedDebt {
			holdingYears = 1
		}
		if heldMoreThanYears(purchaseDay, sellDay, holdingYears) {
			return Tax{Gain: gain, Taxable: gain, LongTerm: gain.Mul(decimal.NewFromFloat(0.125))}
		}
		return shortTerm
	}

	// Before July 23, unlisted shares already used a 24-month threshold.
	oldHoldingYears := 3
	if category == config.UnlistedEquity {
		oldHoldingYears = 2
	}
	if !heldMoreThanYears(purchaseDay, sellDay, oldHoldingYears) {
		return shortTerm
	}

	if purchaseDate.After(CII_START_DATE) {
		sellIndex := decimal.NewFromInt(int64(cii.GetIndex(db, utils.FY(sellDate))))
		purchaseIndex := decimal.NewFromInt(int64(cii.GetIndex(db, utils.FY(purchaseDate))))
		purchasePrice = purchasePrice.Mul(sellIndex.Div(purchaseIndex))
	}
	taxable := sellPrice.Sub(purchasePrice).Mul(quantity)
	return Tax{Gain: gain, Taxable: taxable, LongTerm: taxable.Mul(decimal.NewFromFloat(0.20))}
}

func Calculate(db *gorm.DB, quantity decimal.Decimal, commodity config.Commodity, purchasePrice decimal.Decimal, purchaseDate time.Time, sellPrice decimal.Decimal, sellDate time.Time) Tax {

	dateDiff := sellDate.Sub(purchaseDate)
	gain := sellPrice.Mul(quantity).Sub(purchasePrice.Mul(quantity))

	switch commodity.TaxCategory {
	case config.Debt, config.ListedDebt, config.Equity35, config.UnlistedEquity:
		return calculateSection112(db, commodity.TaxCategory, quantity, purchasePrice, sellPrice, gain, purchaseDate, sellDate)
	}

	isEquity := commodity.TaxCategory == config.Equity || commodity.TaxCategory == config.Equity65
	isLongTermEquity := isEquity && dateDiff > ONE_YEAR

	if isLongTermEquity {
		// Qualifying long-term gains before Section 112A were exempt under Section 10(38).
		if sellDate.Before(equityLTCGStartDate) {
			return Tax{Gain: gain}
		}

		if purchaseDate.Before(equityGrandfatherPurchaseCutoffDate) {
			fmv := service.GetUnitPrice(db, commodity.Name, equityGrandfatherFMVDate).Value
			purchasePrice = decimal.Max(purchasePrice, decimal.Min(fmv, sellPrice))
		}
	}

	taxable := sellPrice.Mul(quantity).Sub(purchasePrice.Mul(quantity))
	shortTerm := decimal.Zero
	longTerm := decimal.Zero

	if isEquity {
		longTermRate, shortTermRate := 0.10, 0.15
		if !calendarDay(sellDate).Before(capitalGainsReformDate) {
			longTermRate, shortTermRate = 0.125, 0.20
		}
		if isLongTermEquity {
			longTerm = taxable.Mul(decimal.NewFromFloat(longTermRate))
		} else {
			shortTerm = taxable.Mul(decimal.NewFromFloat(shortTermRate))
		}
	}

	return Tax{Gain: gain, Taxable: taxable, ShortTerm: shortTerm, LongTerm: longTerm}
}
