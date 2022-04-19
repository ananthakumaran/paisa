package server

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func Listen(db *gorm.DB) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Static("/static", "web/static")
	router.StaticFile("/", "web/static/index.html")
	router.GET("/api/overview", func(c *gin.Context) {
		c.JSON(200, GetOverview(db))
	})
	router.GET("/api/investment", func(c *gin.Context) {
		c.JSON(200, GetInvestment(db))
	})
	router.GET("/api/allocation", func(c *gin.Context) {
		c.JSON(200, GetAllocation(db))
	})
	router.GET("/api/ledger", func(c *gin.Context) {
		c.JSON(200, GetLedger(db))
	})
	log.Info("Listening on 7500")
	err := router.Run(":7500")
	if err != nil {
		log.Fatal(err)
	}
}
