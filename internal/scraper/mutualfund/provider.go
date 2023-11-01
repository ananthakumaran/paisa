package mutualfund

import (
	"github.com/ananthakumaran/paisa/internal/model/mutualfund/scheme"
	"github.com/ananthakumaran/paisa/internal/model/price"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PriceProvider struct {
}

func (p *PriceProvider) Code() string {
	return "in-mfapi"
}

func (p *PriceProvider) Label() string {
	return "MF API India"
}

func (p *PriceProvider) Description() string {
	return "Supports all mutual funds in India."
}

func (p *PriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "AMC", ID: "amc", Help: "Asset Management Company"},
		{Label: "Fund Name", ID: "scheme"},
	}
}

func (p *PriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	count := scheme.Count(db)
	if count == 0 {
		schemes, err := GetSchemes()
		if err != nil {
			log.Fatal(err)
		}
		scheme.UpsertAll(db, schemes)
	} else {
		log.Info("Using cached results")
	}

	switch field {
	case "amc":
		return scheme.GetAMCCompletions(db)
	case "scheme":
		return scheme.GetNAVNameCompletions(db, filter["amc"])
	}
	return []price.AutoCompleteItem{}
}

func (p *PriceProvider) ClearCache(db *gorm.DB) {
	db.Exec("DELETE FROM schemes")
}

func (p *PriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	return GetNav(code, commodityName)
}
