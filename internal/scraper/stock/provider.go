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

func (p *PriceProvider) Description() string {
	return "Supports a large set of stocks, ETFs, mutual funds, currencies, bonds, commodities, and cryptocurrencies. The stock price will be automatically converted to your default currency using the yahoo exchange rate."
}

func (p *PriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "Ticker", ID: "ticker", Help: "Stock ticker symbol, can be located on Yahoo's website. For example, AAPL is the ticker symbol for Apple Inc. (AAPL)", InputType: "text"},
	}
}

func (p *PriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	return []price.AutoCompleteItem{}
}

func (p *PriceProvider) ClearCache(db *gorm.DB) {
}

func (p *PriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	return GetHistory(code, commodityName)
}
