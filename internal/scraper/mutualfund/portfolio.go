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
	q := `
SELECT coalesce(nullIf(i.issuer, ''), nullIf(i.name, ''), p.name) as name,
       p.isin as isin,
       p.percentage_to_nav as percentage_to_nav,
       nullIf(i.type, '') as type,
       nullIf(i.rating, '') as rating
FROM latest_portfolio p
JOIN scheme s ON p.fund_id = s.fund_id
LEFT JOIN security i ON p.isin = i.isin
WHERE s.code = %s
      AND s.category not in ('Hybrid Scheme - Arbitrage Fund', 'Other Scheme - FoF Overseas', 'Other Scheme - Other  ETFs', 'Other Scheme - FoF Domestic')
      AND p.percentage_to_nav > 0
`
	query := fmt.Sprintf(q, schemeCode)

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
		Type            string  `json:"type"`
		Rating          string  `json:"rating"`
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

		portfolio := portfolio.Portfolio{
			SecurityName:      data.Name,
			CommodityType:     price.MutualFund,
			SecurityID:        data.ISIN,
			Percentage:        data.PercentageToNav,
			ParentCommodityID: schemeCode,
			SecurityRating:    data.Rating,
			SecurityType:      data.Type}
		portfolios = append(portfolios, &portfolio)
	}
	return portfolios, nil
}
