package server

import (
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetPrices(db *gorm.DB) gin.H {
	var commodities []string
	result := db.Model(&posting.Posting{}).Where("commodity != ?", config.DefaultCurrency()).Distinct().Pluck("commodity", &commodities)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	var prices = make(map[string][]price.Price)
	for _, commodity := range commodities {
		prices[commodity] = service.GetAllPrices(db, commodity)
	}
	return gin.H{"prices": prices}
}
