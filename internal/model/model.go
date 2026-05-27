package model

import (
	"fmt"
	"strings"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/ledger"
	"github.com/ananthakumaran/paisa/internal/model/cache"
	"github.com/ananthakumaran/paisa/internal/model/commodity"
	"github.com/ananthakumaran/paisa/internal/model/portfolio"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/prediction"
	"github.com/ananthakumaran/paisa/internal/scraper"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&posting.Posting{})
	db.AutoMigrate(&price.Price{})
	db.AutoMigrate(&portfolio.Portfolio{})
	db.AutoMigrate(&price.Price{})
	db.AutoMigrate(&cache.Cache{})
	// account_learning persists user-confirmed payee → account mappings
	// captured by the importer preview UI (issue #24). The prediction
	// package owns the schema so its table sits alongside the model
	// tables rather than under model/. Failure here is logged but not
	// fatal — the suggestion layer falls back to the seed dictionary
	// when the table is missing.
	if err := prediction.AutoMigrateLearning(db); err != nil {
		log.Warnf("failed to migrate account_learning table: %v", err)
	}
}

func SyncJournal(db *gorm.DB) (string, error) {
	AutoMigrate(db)
	log.Info("Syncing transactions from journal")

	errors, _, err := ledger.Cli().ValidateFile(config.GetJournalPath())
	if err != nil {

		if len(errors) == 0 {
			return err.Error(), err
		}

		var message string
		for _, error := range errors {
			message += error.Message + "\n\n"
		}
		return strings.TrimRight(message, "\n"), err
	}

	prices, err := ledger.Cli().Prices(config.GetJournalPath())
	if err != nil {
		return err.Error(), err
	}

	price.UpsertAllByType(db, config.Unknown, prices)

	postings, err := ledger.Cli().Parse(config.GetJournalPath(), prices)
	if err != nil {
		return err.Error(), err
	}
	posting.UpsertAll(db, postings)

	return "", nil
}

func SyncCommodities(db *gorm.DB) error {
	AutoMigrate(db)
	log.Info("Fetching commodities price history")
	commodities := lo.Shuffle(commodity.All())

	var errors []error
	for _, commodity := range commodities {
		name := commodity.Name
		log.Info("Fetching commodity ", name)
		code := commodity.Price.Code
		var prices []*price.Price
		var err error

		provider := scraper.GetProviderByCode(commodity.Price.Provider)
		prices, err = provider.GetPrices(code, name)

		if err != nil {
			log.Error(err)
			errors = append(errors, fmt.Errorf("Failed to fetch price for %s: %w", name, err))
			continue
		}

		price.UpsertAllByTypeNameAndID(db, commodity.Type, name, code, prices)
	}

	if len(errors) > 0 {
		var message string
		for _, error := range errors {
			message += error.Error() + "\n"
		}
		return fmt.Errorf("%s", strings.Trim(message, "\n"))
	}
	return nil
}

// SyncPortfolios is a placeholder that runs the portfolio auto-migration
// and returns. The previous implementation scraped Indian mutual-fund
// holdings via the AMFI portfolio endpoint; that scraper was removed
// when India-specific business logic was dropped. Until a generic
// portfolio provider is added, this is a no-op.
func SyncPortfolios(db *gorm.DB) error {
	db.AutoMigrate(&portfolio.Portfolio{})
	return nil
}
