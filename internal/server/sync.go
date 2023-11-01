package server

import (
	"github.com/ananthakumaran/paisa/internal/accounting"
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
	service.ClearInterestCache()
	service.ClearPriceCache()
	accounting.ClearCache()
	prediction.ClearCache()

	if request.Journal {
		message, err := model.SyncJournal(db)
		if err != nil {
			return gin.H{"success": false, "message": message}
		}
	}

	if request.Prices {
		err := model.SyncCommodities(db)
		if err != nil {
			return gin.H{"success": false, "message": err.Error()}
		}
		err = model.SyncCII(db)
		if err != nil {
			return gin.H{"success": false, "message": err.Error()}
		}
	}

	if request.Portfolios {
		err := model.SyncPortfolios(db)
		if err != nil {
			return gin.H{"success": false, "message": err.Error()}
		}
	}

	return gin.H{"success": true}
}
