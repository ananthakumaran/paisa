package stock

import (
	"github.com/ananthakumaran/paisa/internal/model/price"
	"gorm.io/gorm"
)

type PriceProvider struct {
}

func (p *PriceProvider) Code() string {
	return "com-yahoo"
}

func (p *PriceProvider) Label() string {
	return "Yahoo Finance"
}

func (p *PriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "Ticker", ID: "ticker", Help: "Stock ticker symbol", InputType: "text"},
	}
}

func (p *PriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	return []price.AutoCompleteItem{}
}

func (p *PriceProvider) ClearCache(db *gorm.DB) {
}
