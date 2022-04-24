package service

import (
	"github.com/ChizhovVadim/xirr"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"time"
)

func XIRR(db *gorm.DB, ps []posting.Posting) float64 {
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
