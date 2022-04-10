package server

import (
	"strings"
	"time"

	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Breakdown struct {
	Group        string  `json:"group"`
	Amount       float64 `json:"amount"`
	MarketAmount float64 `json:"market_amount"`
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
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	result := make(map[string]Breakdown)

	setOrUpdate := func(account string, breakdown Breakdown) {
		existingBreakdown, found := result[account]
		if !found {
			existingBreakdown = Breakdown{Group: account}
		}

		existingBreakdown.Amount += breakdown.Amount
		existingBreakdown.MarketAmount += breakdown.MarketAmount
		result[account] = existingBreakdown
	}

	for account, ps := range byAccount {
		amount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 { return acc + p.Amount }, 0.0)
		marketAmount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 { return acc + p.MarketAmount }, 0.0)
		breakdown := Breakdown{Amount: amount, MarketAmount: marketAmount}
		var parts []string
		for _, p := range strings.Split(account, ":") {
			parts = append(parts, p)
			setOrUpdate(strings.Join(parts, ":"), breakdown)
		}
	}

	return result
}
