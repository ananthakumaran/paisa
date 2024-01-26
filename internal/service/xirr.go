package service

import (
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/ananthakumaran/paisa/internal/xirr"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func XIRR(db *gorm.DB, ps []posting.Posting) decimal.Decimal {
	today := utils.EndOfToday()
	marketAmount := utils.SumBy(ps, func(p posting.Posting) decimal.Decimal {
		if IsCapitalGains(p) {
			return decimal.Zero
		}
		return p.MarketAmount
	})
	cashflows := lo.Reverse(lo.Map(ps, func(p posting.Posting, _ int) xirr.Cashflow {
		if IsInterest(db, p) || IsInterestRepayment(db, p) {
			return xirr.Cashflow{Date: p.Date, Amount: 0}
		} else {
			return xirr.Cashflow{Date: p.Date, Amount: p.Amount.Neg().Round(4).InexactFloat64()}
		}
	}))

	cashflows = append(cashflows, xirr.Cashflow{Date: today, Amount: marketAmount.Round(4).InexactFloat64()})
	return xirr.XIRR(cashflows)
}

func APR(db *gorm.DB, ps []posting.Posting) decimal.Decimal {
	today := utils.EndOfToday()
	marketAmount := utils.SumBy(ps, func(p posting.Posting) decimal.Decimal {
		return p.MarketAmount
	})
	cashflows := lo.Map(ps, func(p posting.Posting, _ int) xirr.Cashflow {
		return xirr.Cashflow{Date: p.Date, Amount: p.Amount.Round(4).InexactFloat64()}
	})
	cashflows = append(cashflows, xirr.Cashflow{Date: today, Amount: marketAmount.Neg().Round(4).InexactFloat64()})
	return xirr.XIRR(cashflows)
}
