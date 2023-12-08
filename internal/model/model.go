package model

import (
	"fmt"
	"strings"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/ledger"
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
	commodities := commodity.All()
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
			return fmt.Errorf("Failed to fetch price for %s: %w", name, err)
		}

		price.UpsertAllByTypeNameAndID(db, commodity.Type, name, code, prices)
	}
	return nil
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
