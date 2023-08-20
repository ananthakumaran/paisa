package service

import (
	"time"

	"github.com/ChizhovVadim/xirr"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func XIRR(db *gorm.DB, ps []posting.Posting) decimal.Decimal {
	today := time.Now()
	marketAmount := utils.SumBy(ps, func(p posting.Posting) decimal.Decimal {
		return p.MarketAmount
	})
	payments := lo.Reverse(lo.Map(ps, func(p posting.Posting, _ int) xirr.Payment {
		if IsInterest(db, p) {
			return xirr.Payment{Date: p.Date, Amount: 0}
		} else {
			return xirr.Payment{Date: p.Date, Amount: p.Amount.Neg().Round(4).InexactFloat64()}
		}
	}))
	payments = append(payments, xirr.Payment{Date: today, Amount: marketAmount.Round(4).InexactFloat64()})
	returns, err := xirr.XIRR(payments)
	if err != nil {
		log.Warn(err)
		return decimal.Zero
	}

	return decimal.NewFromFloat(returns).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100))
}

func APR(db *gorm.DB, ps []posting.Posting) decimal.Decimal {
	today := time.Now()
	marketAmount := utils.SumBy(ps, func(p posting.Posting) decimal.Decimal {
		return p.MarketAmount
	})
	payments := lo.Map(ps, func(p posting.Posting, _ int) xirr.Payment {
		return xirr.Payment{Date: p.Date, Amount: p.Amount.Round(4).InexactFloat64()}
	})
	payments = append(payments, xirr.Payment{Date: today, Amount: marketAmount.Neg().Round(4).InexactFloat64()})
	returns, err := xirr.XIRR(payments)
	if err != nil {
		log.Warn(err)
		return decimal.Zero
	}

	return decimal.NewFromFloat(returns).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(-100))
}
