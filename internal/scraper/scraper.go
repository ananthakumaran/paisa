package scraper

import (
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/scraper/cn/eastmoney"
	"github.com/ananthakumaran/paisa/internal/scraper/cn/okx"
	"github.com/ananthakumaran/paisa/internal/scraper/cn/ttjj"
	"github.com/ananthakumaran/paisa/internal/scraper/stock"
	"github.com/ananthakumaran/paisa/internal/scraper/yahoo"
	log "github.com/sirupsen/logrus"
)

func GetAllProviders() []price.PriceProvider {
	return []price.PriceProvider{
		&stock.YahooPriceProvider{},
		&yahoo.PriceProvider{},
		&stock.AlphaVantagePriceProvider{},
		&eastmoney.PriceProvider{},
		&okx.PriceProvider{},
		&ttjj.PriceProvider{},
	}

}

func GetProviderByCode(code string) price.PriceProvider {
	switch code {
	case "com-yahoo":
		return &stock.YahooPriceProvider{}
	case "yahoo":
		return &yahoo.PriceProvider{}
	case "co-alphavantage":
		return &stock.AlphaVantagePriceProvider{}
	case "cn-eastmoney":
		return &eastmoney.PriceProvider{}
	case "cn-okx":
		return &okx.PriceProvider{}
	case "cn-ttjj":
		return &ttjj.PriceProvider{}
	}
	log.Fatal("Unknown price provider: ", code)
	return nil
}
