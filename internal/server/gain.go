package server

import (
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Gain struct {
	Account          string     `json:"account"`
	OverviewTimeline []Overview `json:"overview_timeline"`
}

func GetGain(db *gorm.DB) gin.H {
	var postings []posting.Posting
	result := db.Where("account like ?", "Asset:%").Order("date ASC").Find(&postings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	var gains []Gain
	for account, ps := range byAccount {
		gains = append(gains, Gain{Account: account, OverviewTimeline: computeOverviewTimeline(db, ps)})
	}

	return gin.H{"gain_timeline_breakdown": gains}
}
