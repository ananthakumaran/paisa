package sif

import (
	"github.com/ananthakumaran/paisa/internal/model/price"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PriceProvider struct {
}

func (p *PriceProvider) Code() string {
	return "com-finbodhi-sif"
}

func (p *PriceProvider) Label() string {
	return "FinBodhi SIF India"
}

func (p *PriceProvider) Description() string {
	return "NAV history for Specialized Investment Funds in India, provided by FinBodhi. Prices are quoted in INR."
}

func (p *PriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "SIF scheme", ID: "scheme", Help: "Search by scheme name or SIF number. The selected SIF number is saved as the price code."},
	}
}

func (p *PriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	completions := []price.AutoCompleteItem{}
	if field != "scheme" {
		return completions
	}

	schemes, err := GetSchemes(filter["scheme"])
	if err != nil {
		log.Error("Failed to fetch SIF catalogue: ", err)
		return completions
	}

	for _, item := range schemes {
		completions = append(completions, price.AutoCompleteItem{
			ID:    item.ID,
			Label: item.NAVName + " — " + item.ID,
		})
	}
	return completions
}

func (p *PriceProvider) ClearCache(db *gorm.DB) {
}

func (p *PriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	return GetNav(code, commodityName)
}
