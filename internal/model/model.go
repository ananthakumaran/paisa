package model

import (
	"strconv"

	"github.com/ananthakumaran/paisa/internal/ledger"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/scraper/mutualfund"
	"github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func Sync(db *gorm.DB) {
	db.AutoMigrate(&posting.Posting{})
	log.Info("Syncing postings from ledger")
	postings, _ := ledger.Parse(viper.GetString("journal_path"))
	posting.UpsertAll(db, postings)

	db.AutoMigrate(&price.Price{})
	log.Info("Fetching mutual fund history")
	type Commodity struct {
		Name string
		Type string
		Code int
	}

	var commodities []Commodity
	viper.UnmarshalKey("commodities", &commodities)
	for _, commodity := range commodities {
		name := commodity.Name
		log.Info("Fetching commodity ", aurora.Bold(name))
		schemeCode := strconv.Itoa(commodity.Code)
		prices, err := mutualfund.GetNav(schemeCode, name)
		if err != nil {
			log.Fatal(err)
		}

		price.UpsertAll(db, price.MutualFund, schemeCode, prices)
	}
}
