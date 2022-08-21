package server

import (
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Gain struct {
	Account          string     `json:"account"`
	OverviewTimeline []Overview `json:"overview_timeline"`
	XIRR             float64    `json:"xirr"`
}

func GetGain(db *gorm.DB) gin.H {
	var postings []posting.Posting
	result := db.Where("account like ? and account != ?", "Assets:%", "Assets:Checking").Order("date ASC").Find(&postings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	postings = service.PopulateMarketPrice(db, postings)
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	var gains []Gain
	for account, ps := range byAccount {
		gains = append(gains, Gain{Account: account, XIRR: service.XIRR(db, ps), OverviewTimeline: computeOverviewTimeline(db, ps)})
	}

	return gin.H{"gain_timeline_breakdown": gains}
}
