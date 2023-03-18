package server

import (
	"net/http"

	"github.com/ananthakumaran/paisa/internal/server/assets"
	"github.com/ananthakumaran/paisa/internal/server/liabilities"
	"github.com/ananthakumaran/paisa/internal/server/retirement"
	"github.com/ananthakumaran/paisa/web"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func Listen(db *gorm.DB) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.GET("/static/*filepath", func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, http.FS(web.Static))
	})
	router.POST("/api/sync", func(c *gin.Context) {
		var syncRequest SyncRequest
		if err := c.ShouldBindJSON(&syncRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, Sync(db, syncRequest))
	})
	router.GET("/api/overview", func(c *gin.Context) {
		c.JSON(200, GetOverview(db))
	})
	router.GET("/api/assets/balance", func(c *gin.Context) {
		c.JSON(200, assets.GetBalance(db))
	})

	router.GET("/api/investment", func(c *gin.Context) {
		c.JSON(200, GetInvestment(db))
	})
	router.GET("/api/gain", func(c *gin.Context) {
		c.JSON(200, GetGain(db))
	})
	router.GET("/api/income", func(c *gin.Context) {
		c.JSON(200, GetIncome(db))
	})
	router.GET("/api/expense", func(c *gin.Context) {
		c.JSON(200, GetExpense(db))
	})
	router.GET("/api/allocation", func(c *gin.Context) {
		c.JSON(200, GetAllocation(db))
	})
	router.GET("/api/portfolio_allocation", func(c *gin.Context) {
		c.JSON(200, GetPortfolioAllocation(db))
	})
	router.GET("/api/ledger", func(c *gin.Context) {
		c.JSON(200, GetLedger(db))
	})
	router.GET("/api/price", func(c *gin.Context) {
		c.JSON(200, GetPrices(db))
	})
	router.GET("/api/transaction", func(c *gin.Context) {
		c.JSON(200, GetTransactions(db))
	})
	router.GET("/api/harvest", func(c *gin.Context) {
		c.JSON(200, GetHarvest(db))
	})

	router.GET("/api/capital_gains", func(c *gin.Context) {
		c.JSON(200, GetCapitalGains(db))
	})

	router.GET("/api/schedule_al", func(c *gin.Context) {
		c.JSON(200, GetScheduleAL(db))
	})
	router.GET("/api/diagnosis", func(c *gin.Context) {
		c.JSON(200, GetDiagnosis(db))
	})

	router.GET("/api/retirement/progress", func(c *gin.Context) {
		c.JSON(200, retirement.GetRetirementProgress(db))
	})

	router.GET("/api/liabilities/interest", func(c *gin.Context) {
		c.JSON(200, liabilities.GetInterest(db))
	})

	router.GET("/api/liabilities/balance", func(c *gin.Context) {
		c.JSON(200, liabilities.GetBalance(db))
	})

	router.NoRoute(func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(web.Index))
	})

	log.Info("Listening on 7500")
	err := router.Run(":7500")
	if err != nil {
		log.Fatal(err)
	}
}
