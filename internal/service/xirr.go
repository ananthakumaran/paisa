package service

import (
	"time"

	"github.com/ChizhovVadim/xirr"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func XIRR(db *gorm.DB, ps []posting.Posting) float64 {
	ps = lo.Filter(ps, func(p posting.Posting, _ int) bool { return p.Account != "Assets:Checking" })
	today := time.Now()
	marketAmount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 { return acc + p.MarketAmount }, 0.0)
	payments := lo.Reverse(lo.Map(ps, func(p posting.Posting, _ int) xirr.Payment {
		if IsInterest(db, p) {
			return xirr.Payment{Date: p.Date, Amount: 0}
		} else {
			return xirr.Payment{Date: p.Date, Amount: -p.Amount}
		}
	}))
	payments = append(payments, xirr.Payment{Date: today, Amount: marketAmount})
	returns, err := xirr.XIRR(payments)
	if err != nil {
		return 0
	}

	return (returns - 1) * 100
}

func APR(db *gorm.DB, ps []posting.Posting) float64 {
	today := time.Now()
	marketAmount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 { return acc + p.MarketAmount }, 0.0)
	payments := lo.Map(ps, func(p posting.Posting, _ int) xirr.Payment {
		return xirr.Payment{Date: p.Date, Amount: p.Amount}
	})
	payments = append(payments, xirr.Payment{Date: today, Amount: -marketAmount})
	returns, err := xirr.XIRR(payments)
	if err != nil {
		return 0
	}

	return (returns - 1) * -100
}
