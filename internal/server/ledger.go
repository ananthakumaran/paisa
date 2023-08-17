package server

import (
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetLedger(db *gorm.DB) gin.H {
	postings := query.Init(db).Unbudgeted().Desc().All()
	postings = service.PopulateMarketPrice(db, postings)
	return gin.H{"postings": postings}
}
