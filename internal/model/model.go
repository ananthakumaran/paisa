package model

import (
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/ledger"
	"github.com/ananthakumaran/paisa/internal/model/cii"
	"github.com/ananthakumaran/paisa/internal/model/commodity"
	"github.com/ananthakumaran/paisa/internal/model/portfolio"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/scraper/india"
	"github.com/ananthakumaran/paisa/internal/scraper/mutualfund"
	"github.com/ananthakumaran/paisa/internal/scraper/nps"
	"github.com/ananthakumaran/paisa/internal/scraper/stock"
	"github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func SyncJournal(db *gorm.DB) {
	db.AutoMigrate(&posting.Posting{})
	log.Info("Syncing transactions from journal")
	postings, err := ledger.Cli().Parse(config.JournalPath())
	if err != nil {
		log.Fatal(err)
	}
	posting.UpsertAll(db, postings)
}

func SyncCommodities(db *gorm.DB) {
	db.AutoMigrate(&price.Price{})
	log.Info("Fetching commodities price history")
	commodities := commodity.All()
	for _, commodity := range commodities {
		name := commodity.Name
		log.Info("Fetching commodity ", aurora.Bold(name))
		code := commodity.Code
		var prices []*price.Price
		var err error

		switch commodity.Type {
		case price.MutualFund:
			prices, err = mutualfund.GetNav(code, name)
		case price.NPS:
			prices, err = nps.GetNav(code, name)
		case price.Stock:
			prices, err = stock.GetHistory(code, name)
		}

		if err != nil {
			log.Fatal(err)
		}

		price.UpsertAll(db, commodity.Type, code, prices)
	}
}

func SyncCII(db *gorm.DB) {
	db.AutoMigrate(&cii.CII{})
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
	commodities := commodity.FindByType(price.MutualFund)
	for _, commodity := range commodities {
		name := commodity.Name
		log.Info("Fetching portfolio for ", aurora.Bold(name))
		portfolios, err := mutualfund.GetPortfolio(commodity.Code, commodity.Name)

		if err != nil {
			log.Fatal(err)
		}

		portfolio.UpsertAll(db, commodity.Type, commodity.Code, portfolios)
	}
}
