package nps

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
)

func GetNav(schemeCode string, commodityName string) ([]*price.Price, error) {
	log.Info("Fetching NPS Fund nav from Purified Bytes")
	url := fmt.Sprintf("https://nps.purifiedbytes.com/api/schemes/%s/nav.json", schemeCode)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type Data struct {
		Date string
		Nav  float64
	}
	type Result struct {
		Data []Data
	}

	var result Result
	err = json.Unmarshal(respBytes, &result)
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("UTC")

	var prices []*price.Price
	for _, data := range result.Data {
		date, err := time.ParseInLocation("2006-01-02", data.Date, loc)
		if err != nil {
			return nil, err
		}

		price := price.Price{Date: date, CommodityType: config.NPS, CommodityID: schemeCode, CommodityName: commodityName, Value: data.Nav}
		prices = append(prices, &price)
	}
	return prices, nil
}
