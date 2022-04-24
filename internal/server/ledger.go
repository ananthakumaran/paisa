package server

import (
	"strings"

	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"

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

	postings = service.PopulateMarketPrice(db, postings)
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

		xirr := service.XIRR(db, ps)
		breakdown := Breakdown{InvestmentAmount: investmentAmount, WithdrawalAmount: withdrawalAmount, MarketAmount: marketAmount, XIRR: xirr, Group: group}
		result[group] = breakdown
	}

	return result
}
