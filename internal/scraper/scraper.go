package scraper

import (
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/scraper/metal"
	"github.com/ananthakumaran/paisa/internal/scraper/mutualfund"
	"github.com/ananthakumaran/paisa/internal/scraper/nps"
	"github.com/ananthakumaran/paisa/internal/scraper/stock"
	log "github.com/sirupsen/logrus"
)

func GetAllProviders() []price.PriceProvider {
	return []price.PriceProvider{
		&stock.PriceProvider{},
		&mutualfund.PriceProvider{},
		&nps.PriceProvider{},
		&metal.PriceProvider{},
	}

}

func GetProviderByCode(code string) price.PriceProvider {
	switch code {
	case "in-mfapi":
		return &mutualfund.PriceProvider{}
	case "com-purifiedbytes-nps":
		return &nps.PriceProvider{}
	case "com-purifiedbytes-metal":
		return &metal.PriceProvider{}
	case "com-yahoo":
		return &stock.PriceProvider{}
	}
	log.Fatal("Unknown price provider: ", code)
	return nil
}
