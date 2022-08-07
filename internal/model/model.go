package model

import (
	"github.com/ananthakumaran/paisa/internal/ledger"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/scraper/mutualfund"
	"github.com/ananthakumaran/paisa/internal/scraper/nps"
	"github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func Sync(db *gorm.DB) {
	db.AutoMigrate(&posting.Posting{})
	log.Info("Syncing transactions from journal")
	postings, _ := ledger.Parse(viper.GetString("journal_path"))
	posting.UpsertAll(db, postings)

	db.AutoMigrate(&price.Price{})
	log.Info("Fetching commodities price history")
	type Commodity struct {
		Name string
		Type string
		Code string
	}

	var commodities []Commodity
	viper.UnmarshalKey("commodities", &commodities)
	for _, commodity := range commodities {
		name := commodity.Name
		log.Info("Fetching commodity ", aurora.Bold(name))
		schemeCode := commodity.Code
		var prices []*price.Price
		var err error

		switch commodity.Type {
		case string(price.MutualFund):
			prices, err = mutualfund.GetNav(schemeCode, name)
		case string(price.NPS):
			prices, err = nps.GetNav(schemeCode, name)
		}

		if err != nil {
			log.Fatal(err)
		}

		price.UpsertAll(db, price.MutualFund, schemeCode, prices)
	}
}
