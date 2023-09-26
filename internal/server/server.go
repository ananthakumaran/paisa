package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/generator"
	"github.com/ananthakumaran/paisa/internal/ledger"
	"github.com/ananthakumaran/paisa/internal/model/template"
	"github.com/ananthakumaran/paisa/internal/prediction"
	"github.com/ananthakumaran/paisa/internal/server/assets"
	"github.com/ananthakumaran/paisa/internal/server/liabilities"
	"github.com/ananthakumaran/paisa/internal/server/retirement"
	"github.com/ananthakumaran/paisa/web"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func Build(db *gorm.DB, enableCompression bool) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	if enableCompression {
		router.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	router.Use(Logger(log.StandardLogger()), gin.Recovery())

	router.GET("/_app/*filepath", func(c *gin.Context) {
		c.FileFromFS("/static"+c.Request.URL.Path, http.FS(web.Static))
	})

	router.GET("/api/config", func(c *gin.Context) {
		c.JSON(200, gin.H{"config": config.GetConfig(), "schema": config.GetSchema()})
	})

	router.POST("/api/config", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(200, gin.H{"success": true})
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		err = config.SaveConfig(body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"success": true})
	})

	router.POST("/api/init", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(200, gin.H{"success": true})
			return
		}

		generator.Demo(config.GetConfigDir())
		config.LoadConfigFile(config.GetConfigPath())
		Sync(db, SyncRequest{Journal: true, Prices: true, Portfolios: true})
		c.JSON(200, gin.H{"success": true})
	})

	router.POST("/api/sync", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(200, gin.H{"success": true})
			return
		}

		var syncRequest SyncRequest
		if err := c.ShouldBindJSON(&syncRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, Sync(db, syncRequest))
	})

	router.GET("/api/dashboard", func(c *gin.Context) {
		c.JSON(200, GetDashboard(db))
	})

	router.GET("/api/networth", func(c *gin.Context) {
		c.JSON(200, GetNetworth(db))
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
	router.GET("/api/gain/:account", func(c *gin.Context) {
		account := c.Param("account")
		c.JSON(200, GetAccountGain(db, account))
	})
	router.GET("/api/income", func(c *gin.Context) {
		c.JSON(200, GetIncome(db))
	})
	router.GET("/api/expense", func(c *gin.Context) {
		c.JSON(200, GetExpense(db))
	})

	router.GET("/api/budget", func(c *gin.Context) {
		c.JSON(200, GetBudget(db))
	})

	router.GET("/api/cash_flow", func(c *gin.Context) {
		c.JSON(200, GetCashFlow(db))
	})
	router.GET("/api/recurring", func(c *gin.Context) {
		c.JSON(200, GetRecurringTransactions(db))
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
	router.GET("/api/price/providers", func(c *gin.Context) {
		c.JSON(200, GetPriceProviders(db))
	})

	router.POST("/api/price/providers/delete/:provider", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(200, gin.H{"success": true})
			return
		}

		provider := c.Param("provider")
		c.JSON(200, ClearPriceProviderCache(db, provider))
	})

	router.POST("/api/price/autocomplete", func(c *gin.Context) {
		var autoCompleteRequest AutoCompleteRequest
		if err := c.ShouldBindJSON(&autoCompleteRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, GetPriceAutoCompletions(db, autoCompleteRequest))
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

	router.GET("/api/liabilities/repayment", func(c *gin.Context) {
		c.JSON(200, liabilities.GetRepayment(db))
	})

	router.GET("/api/logs", func(c *gin.Context) {
		c.JSON(200, GetLogs())
	})

	router.GET("/api/editor/files", func(c *gin.Context) {
		c.JSON(200, GetFiles(db))
	})

	router.POST("/api/editor/file", func(c *gin.Context) {
		var ledgerFile LedgerFile
		if err := c.ShouldBindJSON(&ledgerFile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, GetFile(ledgerFile))
	})

	router.POST("/api/editor/file/delete_backups", func(c *gin.Context) {
		var ledgerFile LedgerFile
		if err := c.ShouldBindJSON(&ledgerFile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, DeleteBackups(ledgerFile))
	})

	router.POST("/api/editor/validate", func(c *gin.Context) {
		var ledgerFile LedgerFile
		if err := c.ShouldBindJSON(&ledgerFile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, ValidateFile(ledgerFile))
	})

	router.POST("/api/editor/save", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(200, gin.H{"errors": []ledger.LedgerFileError{}, "saved": false, "message": "Readonly mode"})
			return
		}

		var ledgerFile LedgerFile
		if err := c.ShouldBindJSON(&ledgerFile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, SaveFile(db, ledgerFile))
	})

	router.GET("/api/account/tf_idf", func(c *gin.Context) {
		c.JSON(200, prediction.GetTfIdf(db))
	})

	router.GET("/api/templates", func(c *gin.Context) {
		c.JSON(200, gin.H{"templates": template.All(db)})
	})

	router.POST("/api/templates/upsert", func(c *gin.Context) {
		var t template.Template
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, template.Upsert(db, t.Name, t.Content))
	})

	router.POST("/api/templates/delete", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(200, gin.H{})
			return
		}

		var t template.Template
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		template.Delete(db, t.ID)
		c.JSON(200, gin.H{})
	})

	router.NoRoute(func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(web.Index))
	})

	return router
}

func Listen(db *gorm.DB, port int) {
	router := Build(db, true)

	log.Infof("Listening on http://localhost:%d", port)
	err := router.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
}
