package server

import (
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func GetInvestment(db *gorm.DB) gin.H {
	var postings []posting.Posting
	result := db.Where("account like ?", "Asset:%").Find(&postings)
	postings = lo.Filter(postings, func(p posting.Posting, _ int) bool { return !service.IsInterest(db, p) })
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return gin.H{"postings": postings}
}
