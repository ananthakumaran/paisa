package server

import (
	"log"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetInvestment(db *gorm.DB) gin.H {
	var postings []posting.Posting
	result := db.Where("account like ?", "Asset:%").Find(&postings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return gin.H{"postings": postings}
}
