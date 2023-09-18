package nps

import (
	"github.com/ananthakumaran/paisa/internal/model/nps/scheme"
	"github.com/ananthakumaran/paisa/internal/model/price"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PriceProvider struct {
}

func (p *PriceProvider) Code() string {
	return "com-purifiedbytes-nps"
}

func (p *PriceProvider) Label() string {
	return "Purified Bytes NPS India"
}

func (p *PriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "PFM", ID: "pfm", Help: "Pension Fund Manager"},
		{Label: "Scheme Name", ID: "scheme"},
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
	case "pfm":
		return scheme.GetPFMCompletions(db)
	case "scheme":
		return scheme.GetSchemeNameCompletions(db, filter["pfm"])
	}
	return []price.AutoCompleteItem{}
}

func (p *PriceProvider) ClearCache(db *gorm.DB) {
	db.Exec("DELETE FROM nps_schemes")
}
