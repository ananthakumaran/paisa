package india

import (
	"io"
	"net/http"

	"encoding/json"

	"github.com/ananthakumaran/paisa/internal/model/cii"
	log "github.com/sirupsen/logrus"
)

func GetCostInflationIndex() ([]*cii.CII, error) {
	log.Info("Fetching Cost Inflation Index from Purified Bytes")
	resp, err := http.Get("https://india.finbodhi.com/api/cii/v2.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type CII struct {
		FinancialYear      string `json:"financial_year"`
		CostInflationIndex uint   `json:"cost_inflation_index"`
	}
	type Result struct {
		Data []CII
	}

	var result Result
	err = json.Unmarshal(respBytes, &result)
	if err != nil {
		return nil, err
	}

	var ciis []*cii.CII
	for _, s := range result.Data {
		c := cii.CII{FinancialYear: s.FinancialYear, CostInflationIndex: s.CostInflationIndex}
		ciis = append(ciis, &c)

	}
	return ciis, nil
}
