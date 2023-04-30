package server

import (
	"github.com/ananthakumaran/paisa/internal/model"
	"github.com/ananthakumaran/paisa/internal/prediction"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SyncRequest struct {
	Journal    bool `json:"journal"`
	Prices     bool `json:"prices"`
	Portfolios bool `json:"portfolios"`
}

func Sync(db *gorm.DB, request SyncRequest) gin.H {
	if request.Journal {
		model.SyncJournal(db)
	}

	if request.Prices {
		model.SyncCommodities(db)
		model.SyncCII(db)
	}

	if request.Portfolios {
		model.SyncPortfolios(db)
	}

	service.ClearInterestCache()
	service.ClearPriceCache()
	prediction.ClearCache()
	return gin.H{"success": true}
}
