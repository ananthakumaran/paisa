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
	Date   time.Time `json:"date"`
	Actual float64   `json:"actual"`
	Gain   float64   `json:"gain"`
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

		actual := lo.Reduce(pastPostings, func(agg float64, p posting.Posting, _ int) float64 {
			return p.Amount + agg
		}, 0)

		gain := lo.Reduce(pastPostings, func(agg float64, p posting.Posting, _ int) float64 {
			return service.GetMarketPrice(db, p, start) - p.Amount + agg
		}, 0)
		networths = append(networths, Networth{Date: start, Actual: actual, Gain: gain})
	}
	return networths
}
