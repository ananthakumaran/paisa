package server

import (
	"strings"

	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/scraper"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetPrices(db *gorm.DB) gin.H {
	var commodities []string
	result := db.Model(&posting.Posting{}).Where("commodity != ?", config.DefaultCurrency()).Distinct().Pluck("commodity", &commodities)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	var prices = make(map[string][]price.Price)
	for _, commodity := range commodities {
		prices[commodity] = service.GetAllPrices(db, commodity)
	}
	return gin.H{"prices": prices}
}

type AutoCompleteRequest struct {
	Provider string            `json:"provider"`
	Field    string            `json:"field"`
	Filters  map[string]string `json:"filters"`
}

func GetPriceProviders(db *gorm.DB) gin.H {
	providers := scraper.GetAllProviders()
	return gin.H{
		"providers": lo.Map(providers, func(provider price.PriceProvider, _ int) gin.H {
			return gin.H{
				"code":        provider.Code(),
				"label":       provider.Label(),
				"description": provider.Description(),
				"fields":      provider.AutoCompleteFields(),
			}
		}),
	}

}

func ClearPriceProviderCache(db *gorm.DB, code string) gin.H {
	provider := scraper.GetProviderByCode(code)
	provider.ClearCache(db)
	return gin.H{}
}

func GetPriceAutoCompletions(db *gorm.DB, request AutoCompleteRequest) gin.H {
	provider := scraper.GetProviderByCode(request.Provider)
	completions := provider.AutoComplete(db, request.Field, request.Filters)

	completions = lo.Filter(completions, func(completion price.AutoCompleteItem, _ int) bool {
		item := completion.Label
		item = strings.Replace(strings.ToLower(item), " ", "", -1)
		words := strings.Split(strings.ToLower(request.Filters[request.Field]), " ")
		for _, word := range words {
			if strings.TrimSpace(word) != "" && !strings.Contains(item, word) {
				return false
			}
		}
		return true
	})

	return gin.H{"completions": completions}
}
