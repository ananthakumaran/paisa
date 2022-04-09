package server

import (
	"time"

	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
	return gin.H{"postings": postings}
}
