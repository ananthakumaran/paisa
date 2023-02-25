package server

import (
	"github.com/ananthakumaran/paisa/internal/model"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SyncRequest struct {
	Journal bool `json:"journal"`
	Prices  bool `json:"prices"`
}

func Sync(db *gorm.DB, request SyncRequest) gin.H {
	if request.Journal {
		model.SyncJournal(db)
	}

	if request.Prices {
		model.SyncCommodities(db)
		model.SyncCII(db)
	}

	service.ClearInterestCache()
	service.ClearPriceCache()
	return gin.H{"success": true}
}
