package server

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Networth struct {
	Date             time.Time `json:"date"`
	InvestmentAmount float64   `json:"investment_amount"`
	WithdrawalAmount float64   `json:"withdrawal_amount"`
	GainAmount       float64   `json:"gain_amount"`
}

func GetOverview(db *gorm.DB) gin.H {
	var postings []posting.Posting
	result := db.Where("account like ?", "Asset:%").Order("date ASC").Find(&postings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	networthTimeline := ComputeTimeline(db, postings)
	return gin.H{"networth_timeline": networthTimeline}
}

func ComputeTimeline(db *gorm.DB, postings []posting.Posting) []Networth {
	var networths []Networth

	var p posting.Posting
	var pastPostings []posting.Posting

	end := time.Now()
	for start := postings[0].Date; start.Before(end); start = start.AddDate(0, 0, 1) {
		for len(postings) > 0 && (postings[0].Date.Before(start) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			pastPostings = append(pastPostings, p)
		}

		investment := lo.Reduce(pastPostings, func(agg float64, p posting.Posting, _ int) float64 {
			if p.Amount < 0 || service.IsInterest(db, p) {
				return agg
			} else {
				return p.Amount + agg
			}
		}, 0)

		withdrawal := lo.Reduce(pastPostings, func(agg float64, p posting.Posting, _ int) float64 {
			if p.Amount > 0 || service.IsInterest(db, p) {
				return agg
			} else {
				return -p.Amount + agg
			}
		}, 0)

		gain := lo.Reduce(pastPostings, func(agg float64, p posting.Posting, _ int) float64 {
			if service.IsInterest(db, p) {
				return p.Amount + agg
			} else {
				return service.GetMarketPrice(db, p, start) - p.Amount + agg
			}
		}, 0)
		networths = append(networths, Networth{Date: start, InvestmentAmount: investment, WithdrawalAmount: withdrawal, GainAmount: gain})
	}
	return networths
}
