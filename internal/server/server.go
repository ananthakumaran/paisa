package server

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/encryption"
	"github.com/ananthakumaran/paisa/internal/generator"
	"github.com/ananthakumaran/paisa/internal/ledger"
	"github.com/ananthakumaran/paisa/internal/model/template"
	"github.com/ananthakumaran/paisa/internal/prediction"
	"github.com/ananthakumaran/paisa/internal/server/assets"
	"github.com/ananthakumaran/paisa/internal/server/goal"
	"github.com/ananthakumaran/paisa/internal/server/liabilities"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/ananthakumaran/paisa/web"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"
	"gorm.io/gorm"
)

func Build(db *gorm.DB, enableCompression bool) *gin.Engine {
	return buildRouter(db, enableCompression, web.Static, web.Index)
}

// BuildWithAssets is like Build but serves the SPA from the given filesystem and index HTML (e.g. rupee's embed).
func BuildWithAssets(db *gorm.DB, enableCompression bool, static fs.FS, indexHTML string) *gin.Engine {
	return buildRouter(db, enableCompression, static, indexHTML)
}

func buildRouter(db *gorm.DB, enableCompression bool, static fs.FS, indexHTML string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	if enableCompression {
		router.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	router.Use(Logger(log.StandardLogger()), gin.Recovery())

	router.Use(TokenAuthMiddleware())

	router.GET("/robots.txt", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("User-agent: *\nDisallow: /"))
	})

	router.GET("/_app/*filepath", func(c *gin.Context) {
		c.FileFromFS("/static"+c.Request.URL.Path, http.FS(static))
	})

	router.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	router.GET("/api/config", func(c *gin.Context) {
		var now *time.Time
		if utils.IsNowDefined() {
			n := utils.Now()
			now = &n
		}
		c.JSON(200, gin.H{"config": config.GetConfig(), "accounts": accounting.AllAccounts(db), "now": now, "schema": config.GetSchema()})
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
	router.GET("/api/income_statement", func(c *gin.Context) {
		c.JSON(200, GetIncomeStatement(db))
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
	router.POST("/api/price/delete", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(200, gin.H{"success": true})
			return
		}

		c.JSON(200, ClearPriceCache(db))
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

	router.GET("/api/transaction/balanced", func(c *gin.Context) {
		c.JSON(200, GetBalancedPostings(db))
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
		h, code := GetFiles(db)
		c.JSON(code, h)
	})

	router.POST("/api/editor/file", func(c *gin.Context) {
		var ledgerFile LedgerFile
		if err := c.ShouldBindJSON(&ledgerFile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		h, code := GetFile(ledgerFile)
		c.JSON(code, h)
	})

	router.POST("/api/editor/file/delete_backups", func(c *gin.Context) {
		var ledgerFile LedgerFile
		if err := c.ShouldBindJSON(&ledgerFile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		h, code := DeleteBackups(ledgerFile)
		c.JSON(code, h)
	})

	router.POST("/api/editor/validate", func(c *gin.Context) {
		var ledgerFile LedgerFile
		if err := c.ShouldBindJSON(&ledgerFile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		h, code := ValidateFile(ledgerFile)
		c.JSON(code, h)
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

		h, code := SaveFile(db, ledgerFile)
		c.JSON(code, h)
	})

	router.GET("/api/sheets/files", func(c *gin.Context) {
		h, code := GetSheets(db)
		c.JSON(code, h)
	})

	router.POST("/api/sheets/file", func(c *gin.Context) {
		var sheetFile SheetFile
		if err := c.ShouldBindJSON(&sheetFile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		h, code := GetSheet(sheetFile)
		c.JSON(code, h)
	})

	router.POST("/api/sheets/file/delete_backups", func(c *gin.Context) {
		var sheetFile SheetFile
		if err := c.ShouldBindJSON(&sheetFile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		h, code := DeleteSheetBackups(sheetFile)
		c.JSON(code, h)
	})

	router.POST("/api/sheets/save", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(200, gin.H{"saved": false, "message": "Readonly mode"})
			return
		}

		var sheetFile SheetFile
		if err := c.ShouldBindJSON(&sheetFile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		h, code := SaveSheetFile(db, sheetFile)
		c.JSON(code, h)
	})

	router.GET("/api/account/tf_idf", func(c *gin.Context) {
		c.JSON(200, prediction.GetTfIdf(db))
	})

	router.GET("/api/templates", func(c *gin.Context) {
		c.JSON(200, gin.H{"templates": template.All()})
	})

	router.POST("/api/templates/upsert", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(200, gin.H{"saved": false, "message": "Readonly mode"})
			return
		}

		var t template.Template
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"template": template.Upsert(t.Name, t.Content), "saved": true})
	})

	router.POST("/api/templates/delete", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(200, gin.H{"success": false, "message": "Readonly mode"})
			return
		}

		var t template.Template
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		template.Delete(t.Name)
		c.JSON(200, gin.H{"success": true})
	})

	router.GET("/api/goals", func(c *gin.Context) {
		c.JSON(200, gin.H{"goals": goal.GetGoalSummaries(db)})
	})

	router.GET("/api/goals/:type/:name", func(c *gin.Context) {
		c.JSON(200, goal.GetGoalDetails(db, c.Param("type"), c.Param("name")))
	})

	router.GET("/api/credit_cards", func(c *gin.Context) {
		c.JSON(200, GetCreditCards(db))
	})

	router.GET("/api/credit_cards/:account", func(c *gin.Context) {
		c.JSON(200, GetCreditCard(db, c.Param("account")))
	})

	router.GET("/api/encryption/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"enabled":      config.IsEncryptionEnabled(),
			"password_set": encryption.IsPasswordSet(),
			"needs_unlock": encryptionNeedsUnlock(),
		})
	})

	router.POST("/api/encryption/password", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Readonly mode"})
			return
		}

		var body struct {
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}

		if err := trySetEncryptionPasswordFromInput(body.Password); err != nil {
			if errors.Is(err, encryption.ErrDecryptFailed) {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Wrong password or corrupted encrypted file",
					"code":    "encryption_bad_password",
				})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	router.POST("/api/encryption/encrypt", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(200, gin.H{"success": false, "message": "Readonly mode"})
			return
		}

		if !encryption.IsPasswordSet() {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Encryption password not set. Enter it in the encryption dialog."})
			return
		}

		journalPath := config.GetJournalPath()
		ext := filepath.Ext(journalPath)
		dir := filepath.Dir(journalPath)

		count, err := encryption.EncryptExistingFiles(dir, ext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
			return
		}

		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("Encrypted %d file(s)", count)})
	})

	router.POST("/api/encryption/decrypt", func(c *gin.Context) {
		if config.GetConfig().Readonly {
			c.JSON(200, gin.H{"success": false, "message": "Readonly mode"})
			return
		}

		if !encryption.IsPasswordSet() {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Encryption password not set. Enter it in the encryption dialog."})
			return
		}

		journalPath := config.GetJournalPath()
		ext := filepath.Ext(journalPath)
		dir := filepath.Dir(journalPath)

		count, err := encryption.DecryptExistingFiles(dir, ext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
			return
		}

		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("Decrypted %d file(s)", count)})
	})

	router.NoRoute(func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(indexHTML))
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

// ListenWithAssets serves the API with a custom embedded frontend (used by rupee).
func ListenWithAssets(db *gorm.DB, port int, static fs.FS, indexHTML string) {
	router := BuildWithAssets(db, true, static, indexHTML)

	log.Infof("Listening on http://localhost:%d", port)
	err := router.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
}

func TokenAuthMiddleware() gin.HandlerFunc {
	store, err := memstore.NewCtx(10)
	if err != nil {
		log.Fatal(err)
	}

	quota := throttled.RateQuota{
		MaxRate:  throttled.PerMin(6),
		MaxBurst: 3,
	}

	rateLimiter, err := throttled.NewGCRARateLimiterCtx(store, quota)
	if err != nil {
		log.Fatal(err)
	}

	return func(c *gin.Context) {
		userAccounts := config.GetConfig().UserAccounts
		if len(userAccounts) == 0 || !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}

		_, detail, _ := rateLimiter.RateLimitCtx(c.Request.Context(), "user", 0)
		if detail.Remaining <= 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		tokens := strings.SplitN(c.Request.Header.Get("X-Auth"), ":", 2)
		if len(tokens) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
			return
		}

		hashed := utils.Sha256(tokens[1])
		for _, userAccount := range userAccounts {
			if subtle.ConstantTimeCompare([]byte(userAccount.Username), []byte(tokens[0])) == 1 &&
				subtle.ConstantTimeCompare([]byte(userAccount.Password), []byte("sha256:"+hashed)) == 1 {
				c.Next()
				return
			}
		}

		rateLimiter.RateLimitCtx(c.Request.Context(), "user", 1)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return

	}
}
