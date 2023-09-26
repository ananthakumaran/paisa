package model

import (
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/ledger"
	"github.com/ananthakumaran/paisa/internal/model/cii"
	"github.com/ananthakumaran/paisa/internal/model/commodity"
	mutualfundModel "github.com/ananthakumaran/paisa/internal/model/mutualfund/scheme"
	npsModel "github.com/ananthakumaran/paisa/internal/model/nps/scheme"
	"github.com/ananthakumaran/paisa/internal/model/portfolio"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/model/template"
	"github.com/ananthakumaran/paisa/internal/scraper/india"
	"github.com/ananthakumaran/paisa/internal/scraper/mutualfund"
	"github.com/ananthakumaran/paisa/internal/scraper/nps"
	"github.com/ananthakumaran/paisa/internal/scraper/stock"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&npsModel.Scheme{})
	db.AutoMigrate(&mutualfundModel.Scheme{})
	db.AutoMigrate(&posting.Posting{})
	db.AutoMigrate(&price.Price{})
	db.AutoMigrate(&portfolio.Portfolio{})
	db.AutoMigrate(&template.Template{})
	db.AutoMigrate(&price.Price{})
	db.AutoMigrate(&cii.CII{})
}

func SyncJournal(db *gorm.DB) {
	AutoMigrate(db)
	log.Info("Syncing transactions from journal")
	prices, err := ledger.Cli().Prices(config.GetJournalPath())
	if err != nil {
		log.Fatal(err)
	}

	price.UpsertAllByType(db, config.Unknown, prices)

	postings, err := ledger.Cli().Parse(config.GetJournalPath(), prices)
	if err != nil {
		log.Fatal(err)
	}
	posting.UpsertAll(db, postings)
}

func SyncCommodities(db *gorm.DB) {
	AutoMigrate(db)
	log.Info("Fetching commodities price history")
	commodities := commodity.All()
	for _, commodity := range commodities {
		name := commodity.Name
		log.Info("Fetching commodity ", name)
		code := commodity.Price.Code
		var prices []*price.Price
		var err error

		switch commodity.Type {
		case config.MutualFund:
			prices, err = mutualfund.GetNav(code, name)
		case config.NPS:
			prices, err = nps.GetNav(code, name)
		case config.Stock:
			prices, err = stock.GetHistory(code, name)
		}

		if err != nil {
			log.Fatal(err)
		}

		price.UpsertAllByTypeAndID(db, commodity.Type, code, prices)
	}
}

func SyncCII(db *gorm.DB) {
	AutoMigrate(db)
	log.Info("Fetching taxation related info")
	ciis, err := india.GetCostInflationIndex()
	if err != nil {
		log.Fatal(err)
	}
	cii.UpsertAll(db, ciis)
}

func SyncPortfolios(db *gorm.DB) {
	db.AutoMigrate(&portfolio.Portfolio{})
	log.Info("Fetching commodities portfolio")
	commodities := commodity.FindByType(config.MutualFund)
	for _, commodity := range commodities {
		name := commodity.Name
		log.Info("Fetching portfolio for ", name)
		portfolios, err := mutualfund.GetPortfolio(commodity.Price.Code, commodity.Name)

		if err != nil {
			log.Fatal(err)
		}

		portfolio.UpsertAll(db, commodity.Type, commodity.Price.Code, portfolios)
	}
}
