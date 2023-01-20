package server

import (
	"github.com/ananthakumaran/paisa/internal/model"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Sync(db *gorm.DB) gin.H {
	model.SyncJournal(db)
	model.SyncCommodities(db)
	model.SyncCII(db)
	service.ClearInterestCache()
	service.ClearPriceCache()
	return gin.H{"success": true}
}
