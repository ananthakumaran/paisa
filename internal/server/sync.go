package server

import (
	"github.com/ananthakumaran/paisa/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Sync(db *gorm.DB) gin.H {
	model.SyncJournal(db)
	model.SyncCommodities(db)
	model.SyncCII(db)
	return gin.H{"success": true}
}
