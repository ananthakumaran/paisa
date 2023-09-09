package mutualfund

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
)

func GetNav(schemeCode string, commodityName string) ([]*price.Price, error) {
	log.Info("Fetching Mutual Fund nav from mfapi.in")
	url := fmt.Sprintf("https://api.mfapi.in/mf/%s", schemeCode)
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
		Date string
		Nav  string
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
		date, err := time.ParseInLocation("02-01-2006", data.Date, time.Local)
		if err != nil {
			return nil, err
		}
		value, err := strconv.ParseFloat(data.Nav, 64)
		if err != nil {
			return nil, err
		}

		price := price.Price{Date: date, CommodityType: config.MutualFund, CommodityID: schemeCode, CommodityName: commodityName, Value: decimal.NewFromFloat(value)}
		prices = append(prices, &price)
	}
	return prices, nil
}
