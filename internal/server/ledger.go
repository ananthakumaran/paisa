package server

import (
	"strings"
	"time"

	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"

	"github.com/ChizhovVadim/xirr"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Breakdown struct {
	Group            string  `json:"group"`
	InvestmentAmount float64 `json:"investment_amount"`
	WithdrawalAmount float64 `json:"withdrawal_amount"`
	MarketAmount     float64 `json:"market_amount"`
	XIRR             float64 `json:"xirr"`
}

func GetLedger(db *gorm.DB) gin.H {
	var postings []posting.Posting
	result := db.Order("date DESC").Find(&postings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	date := time.Now()
	postings = lo.Map(postings, func(p posting.Posting, _ int) posting.Posting {
		p.MarketAmount = service.GetMarketPrice(db, p, date)
		return p
	})
	breakdowns := computeBreakdown(db, lo.Filter(postings, func(p posting.Posting, _ int) bool { return strings.HasPrefix(p.Account, "Asset:") }))
	return gin.H{"postings": postings, "breakdowns": breakdowns}
}

func computeBreakdown(db *gorm.DB, postings []posting.Posting) map[string]Breakdown {
	accounts := make(map[string]bool)
	for _, p := range postings {
		var parts []string
		for _, part := range strings.Split(p.Account, ":") {
			parts = append(parts, part)
			accounts[strings.Join(parts, ":")] = true
		}

	}

	today := time.Now()
	result := make(map[string]Breakdown)

	for group := range accounts {
		ps := lo.Filter(postings, func(p posting.Posting, _ int) bool { return strings.HasPrefix(p.Account, group) })
		investmentAmount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 {
			if p.Amount < 0 || service.IsInterest(db, p) {
				return acc
			} else {
				return acc + p.Amount
			}
		}, 0.0)
		withdrawalAmount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 {
			if p.Amount > 0 || service.IsInterest(db, p) {
				return acc
			} else {
				return acc + -p.Amount
			}
		}, 0.0)
		marketAmount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 { return acc + p.MarketAmount }, 0.0)
		payments := lo.Reverse(lo.Map(ps, func(p posting.Posting, _ int) xirr.Payment {
			if service.IsInterest(db, p) {
				return xirr.Payment{Date: p.Date, Amount: 0}
			} else {
				return xirr.Payment{Date: p.Date, Amount: -p.Amount}
			}

		}))
		payments = append(payments, xirr.Payment{Date: today, Amount: marketAmount})
		returns, err := xirr.XIRR(payments)
		if err != nil {
			log.Fatal(err)
		}
		breakdown := Breakdown{InvestmentAmount: investmentAmount, WithdrawalAmount: withdrawalAmount, MarketAmount: marketAmount, XIRR: (returns - 1) * 100, Group: group}
		result[group] = breakdown
	}

	return result
}
