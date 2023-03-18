package mutualfund

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/model/portfolio"
	"github.com/ananthakumaran/paisa/internal/model/price"
)

func GetPortfolio(schemeCode string, commodityName string) ([]*portfolio.Portfolio, error) {
	log.Info("Fetching Mutual Fund portfolio from Purified Bytes")
	url := "https://mutualfund.purifiedbytes.com?default_format=JSON"
	query := fmt.Sprintf("select p.name, p.isin, p.percentage_to_nav from latest_portfolio p join scheme s on p.fund_id = s.fund_id where s.code = %s and s.category != 'Hybrid Scheme - Arbitrage Fund' and p.percentage_to_nav > 0", schemeCode)

	req, err := http.NewRequest("POST", url, strings.NewReader(query))
	req.Header.Add("Content-Type", "text/plain")
	req.Header.Add("Authorization", "Basic cGxheTo=")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type Data struct {
		Name            string  `json:"name"`
		PercentageToNav float64 `json:"percentage_to_nav"`
		ISIN            string  `json:"isin"`
	}
	type Result struct {
		Data []Data
	}

	var result Result
	err = json.Unmarshal(respBytes, &result)
	if err != nil {
		return nil, err
	}

	var portfolios []*portfolio.Portfolio
	for _, data := range result.Data {

		portfolio := portfolio.Portfolio{CommodityName: data.Name, CommodityType: price.MutualFund, CommodityID: data.ISIN, Percentage: data.PercentageToNav, ParentCommodityID: schemeCode}
		portfolios = append(portfolios, &portfolio)
	}
	return portfolios, nil
}
