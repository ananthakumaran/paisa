package mutualfund

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"net/http"

	"github.com/ananthakumaran/paisa/internal/model/price"
)

func GetNav(schemeCode string) ([]*price.Price, error) {
	url := fmt.Sprintf("https://api.mfapi.in/mf/%s", schemeCode)
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
		date, err := time.Parse("02-01-2006", data.Date)
		if err != nil {
			return nil, err
		}
		value, err := strconv.ParseFloat(data.Nav, 64)
		if err != nil {
			return nil, err
		}

		price := price.Price{Date: date, CommodityType: price.MutualFund, CommodityID: schemeCode, Value: value}
		prices = append(prices, &price)
	}
	return prices, nil
}
