package metal

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io"
	"net/http"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type PriceProvider struct {
}

func (p *PriceProvider) Code() string {
	return "com-purifiedbytes-metal"
}

func (p *PriceProvider) Label() string {
	return "Purified Bytes Metals India"
}

func (p *PriceProvider) Description() string {
	return "Supports IBJA (India) gold and silver prices at various level of purity."
}

func (p *PriceProvider) AutoCompleteFields() []price.AutoCompleteField {
	return []price.AutoCompleteField{
		{Label: "Metal", ID: "metal", Help: "Metal name with purity."},
	}
}

func (p *PriceProvider) AutoComplete(db *gorm.DB, field string, filter map[string]string) []price.AutoCompleteItem {
	return []price.AutoCompleteItem{
		{Label: "Gold 999", ID: "gold-999"},
		{Label: "Gold 995", ID: "gold-995"},
		{Label: "Gold 916", ID: "gold-916"},
		{Label: "Gold 750", ID: "gold-750"},
		{Label: "Gold 585", ID: "gold-585"},
		{Label: "Silver 999", ID: "silver-999"},
	}
}

func (p *PriceProvider) ClearCache(db *gorm.DB) {
}

func (p *PriceProvider) GetPrices(code string, commodityName string) ([]*price.Price, error) {
	log.Info("Fetching Metal price history from Purified Bytes")
	url := fmt.Sprintf("https://india.purifiedbytes.com/api/metal/%s/price.json", code)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type Data struct {
		Date  string
		Open  decimal.Decimal
		Close decimal.Decimal
	}
	type Result struct {
		Data []Data
	}

	var result Result
	err = json.Unmarshal(respBytes, &result)
	if err != nil {
		return nil, err
	}

	var prices []*price.Price
	for _, data := range result.Data {
		date, err := time.ParseInLocation("2006-01-02", data.Date, config.TimeZone())
		if err != nil {
			return nil, err
		}

		price := price.Price{Date: date, CommodityType: config.Metal, CommodityID: code, CommodityName: commodityName, Value: data.Close.Div(decimal.NewFromInt(10))}
		prices = append(prices, &price)
	}
	return prices, nil
}
