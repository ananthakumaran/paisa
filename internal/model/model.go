package model

import (
	"fmt"
	"strings"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/ledger"
	"github.com/ananthakumaran/paisa/internal/model/cache"
	"github.com/ananthakumaran/paisa/internal/model/cii"
	"github.com/ananthakumaran/paisa/internal/model/commodity"
	mutualfundModel "github.com/ananthakumaran/paisa/internal/model/mutualfund/scheme"
	npsModel "github.com/ananthakumaran/paisa/internal/model/nps/scheme"
	"github.com/ananthakumaran/paisa/internal/model/portfolio"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/scraper"
	"github.com/ananthakumaran/paisa/internal/scraper/india"
	"github.com/ananthakumaran/paisa/internal/scraper/mutualfund"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&npsModel.Scheme{})
	db.AutoMigrate(&mutualfundModel.Scheme{})
	db.AutoMigrate(&posting.Posting{})
	db.AutoMigrate(&price.Price{})
	db.AutoMigrate(&portfolio.Portfolio{})
	db.AutoMigrate(&price.Price{})
	db.AutoMigrate(&cii.CII{})
	db.AutoMigrate(&cache.Cache{})
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
			message += error.String() + "\n\n"
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

func SyncCommodities(db *gorm.DB) ([]string, error) {
	conf := config.GetConfig()
	commodities, err := conf.OrderedPriceCommodities()
	if err != nil {
		return nil, err
	}

	AutoMigrate(db)
	log.Info("Fetching commodities price history")

	dependencies := make(map[string]bool)
	for _, commodity := range commodities {
		quote := commodity.Price.QuoteCommodity(conf.DefaultCurrency)
		if quote != conf.DefaultCurrency {
			dependencies[quote] = true
		}
	}
	histories := make(map[string]priceHistory)

	var errors []error
	var warnings []string
	for _, commodity := range commodities {
		name := commodity.Name
		quote := commodity.Price.QuoteCommodity(conf.DefaultCurrency)
		exchange, available := histories[quote]
		if quote != conf.DefaultCurrency && !available {
			err := fmt.Errorf("Skipped price for %s: quote commodity %s did not update successfully", name, quote)
			log.Error(err)
			errors = append(errors, err)
			continue
		}

		log.Info("Fetching commodity ", name)
		provider := scraper.GetProviderByCode(commodity.Price.Provider)
		raw, err := provider.GetPrices(commodity.Price.Code, name)
		if err != nil {
			log.Error(err)
			errors = append(errors, fmt.Errorf("Failed to fetch price for %s: %w", name, err))
			continue
		}

		prices, gap, err := normalizeCommodityPrices(commodity, raw, exchange, conf.DefaultCurrency)
		if err != nil {
			log.Error(err)
			errors = append(errors, fmt.Errorf("Failed to convert price for %s: %w", name, err))
			continue
		}
		if gap.Count > 0 {
			warning := fmt.Sprintf("%s: %s", name, gap)
			warnings = append(warnings, warning)
			log.Warn(warning)
		}

		if err := price.UpsertAllByTypeAndName(db, commodity.Type, name, prices); err != nil {
			log.Error(err)
			errors = append(errors, fmt.Errorf("Failed to store price for %s: %w", name, err))
			continue
		}
		if dependencies[name] {
			histories[name] = newPriceHistory(prices)
		}
	}

	if len(errors) > 0 {
		var message string
		for _, error := range errors {
			message += error.Error() + "\n"
		}
		return warnings, fmt.Errorf("%s", strings.Trim(message, "\n"))
	}
	return warnings, nil
}

func SyncCII(db *gorm.DB) error {
	AutoMigrate(db)
	log.Info("Fetching taxation related info")
	ciis, err := india.GetCostInflationIndex()
	if err != nil {
		log.Error(err)
		return fmt.Errorf("Failed to fetch CII: %w", err)
	}
	cii.UpsertAll(db, ciis)
	return nil
}

func SyncPortfolios(db *gorm.DB) error {
	db.AutoMigrate(&portfolio.Portfolio{})
	log.Info("Fetching commodities portfolio")
	commodities := commodity.FindByType(config.MutualFund)
	for _, commodity := range commodities {
		if commodity.Price.Provider != "in-mfapi" {
			continue
		}

		name := commodity.Name
		log.Info("Fetching portfolio for ", name)
		portfolios, err := mutualfund.GetPortfolio(commodity.Price.Code, commodity.Name)

		if err != nil {
			log.Error(err)
			return fmt.Errorf("Failed to fetch portfolio for %s: %w", name, err)
		}

		portfolio.UpsertAll(db, commodity.Type, commodity.Price.Code, portfolios)
	}
	return nil
}
