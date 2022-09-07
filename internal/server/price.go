package server

import (
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

func GetPrices(db *gorm.DB) gin.H {
	var commodities []string
	result := db.Model(&posting.Posting{}).Where("commodity != ?", "INR").Distinct().Pluck("commodity", &commodities)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	today := time.Now()
	var prices []price.Price
	for _, commodity := range commodities {
		prices = append(prices, service.GetUnitPrice(db, commodity, today))
	}
	return gin.H{"prices": prices}
}
