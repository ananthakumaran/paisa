package mutualfund

import (
	"encoding/csv"
	"net/http"

	"github.com/ananthakumaran/paisa/internal/model/mutualfund/scheme"
	log "github.com/sirupsen/logrus"
)

func GetSchemes() ([]*scheme.Scheme, error) {
	log.Info("Fetching Mutual Fund Scheme list from AMFI Website")
	resp, err := http.Get("https://portal.amfiindia.com/DownloadSchemeData_Po.aspx?mf=0")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var schemes []*scheme.Scheme
	for _, record := range records[1:] {
		scheme := scheme.Scheme{AMC: record[0], Code: record[1], Name: record[2], Type: record[3], Category: record[4], NAVName: record[5]}
		schemes = append(schemes, &scheme)

	}
	return schemes, nil
}
