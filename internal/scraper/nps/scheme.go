package nps

import (
	"io"
	"net/http"

	"encoding/json"

	"github.com/ananthakumaran/paisa/internal/model/nps/scheme"
	log "github.com/sirupsen/logrus"
)

func GetSchemes() ([]*scheme.Scheme, error) {
	log.Info("Fetching NPS scheme list from Purified Bytes")
	resp, err := http.Get("https://nps.purifiedbytes.com/api/schemes.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type Scheme struct {
		Id      string
		Name    string
		PFMName string `json:"pfm_name"`
	}
	type Result struct {
		Data []Scheme
	}

	var result Result
	err = json.Unmarshal(respBytes, &result)
	if err != nil {
		return nil, err
	}

	var schemes []*scheme.Scheme
	for _, s := range result.Data {
		scheme := scheme.Scheme{PFMName: s.PFMName, SchemeID: s.Id, SchemeName: s.Name}
		schemes = append(schemes, &scheme)

	}
	return schemes, nil
}
