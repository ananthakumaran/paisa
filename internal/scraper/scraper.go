package scraper

import (
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/scraper/metal"
	"github.com/ananthakumaran/paisa/internal/scraper/mutualfund"
	"github.com/ananthakumaran/paisa/internal/scraper/nps"
	"github.com/ananthakumaran/paisa/internal/scraper/sif"
	"github.com/ananthakumaran/paisa/internal/scraper/stock"
	log "github.com/sirupsen/logrus"
)

func GetAllProviders() []price.PriceProvider {
	return []price.PriceProvider{
		&stock.YahooPriceProvider{},
		&stock.FinBodhiPriceProvider{},
		&mutualfund.PriceProvider{},
		&mutualfund.FinBodhiPriceProvider{},
		&stock.AlphaVantagePriceProvider{},
		&nps.PriceProvider{},
		&sif.PriceProvider{},
		&metal.PriceProvider{},
	}

}

func GetProviderByCode(code string) price.PriceProvider {
	switch code {
	case "in-mfapi":
		return &mutualfund.PriceProvider{}
	case "com-finbodhi-mutualfund":
		return &mutualfund.FinBodhiPriceProvider{}
	case "com-finbodhi-nps":
		return &nps.PriceProvider{}
	case "com-finbodhi-sif":
		return &sif.PriceProvider{}
	case "com-finbodhi-metal":
		return &metal.PriceProvider{}
	case "com-yahoo":
		return &stock.YahooPriceProvider{}
	case "com-finbodhi-stock":
		return &stock.FinBodhiPriceProvider{}
	case "co-alphavantage":
		return &stock.AlphaVantagePriceProvider{}
	}
	log.Fatal("Unknown price provider: ", code)
	return nil
}
