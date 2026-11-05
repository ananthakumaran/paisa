package server

import (
	"github.com/ananthakumaran/paisa/internal/cache"
	"github.com/ananthakumaran/paisa/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SyncRequest struct {
	Journal    bool `json:"journal"`
	Prices     bool `json:"prices"`
	Portfolios bool `json:"portfolios"`
}

func Sync(db *gorm.DB, request SyncRequest) gin.H {
	cache.Clear()
	defer cache.Clear()
	warnings := make([]string, 0)

	if request.Journal {
		message, err := model.SyncJournal(db)
		if err != nil {
			return gin.H{"success": false, "message": message, "warnings": warnings}
		}
	}

	if request.Prices {
		priceWarnings, err := model.SyncCommodities(db)
		warnings = append(warnings, priceWarnings...)
		if err != nil {
			return gin.H{"success": false, "message": err.Error(), "warnings": warnings}
		}
		err = model.SyncCII(db)
		if err != nil {
			return gin.H{"success": false, "message": err.Error(), "warnings": warnings}
		}
	}

	if request.Portfolios {
		err := model.SyncPortfolios(db)
		if err != nil {
			return gin.H{"success": false, "message": err.Error(), "warnings": warnings}
		}
	}

	return gin.H{"success": true, "warnings": warnings}
}
