package cmd

import (
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/stretchr/testify/require"
)

type failedPriceTransport struct{}

func (failedPriceTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("fixture provider failure")
}

func TestUpdateReturnsPriceFailure(t *testing.T) {
	oldConfig, err := json.Marshal(config.GetConfig())
	require.NoError(t, err)
	if config.GetConfig().JournalPath == "" {
		oldConfig = []byte(`{"journal_path":"main.ledger","db_path":"paisa.db"}`)
	}
	t.Cleanup(func() { require.NoError(t, config.LoadConfig(oldConfig, "")) })
	oldJournal, oldCommodities, oldPortfolios := updateJournal, updateCommodities, updatePortfolios
	t.Cleanup(func() { updateJournal, updateCommodities, updatePortfolios = oldJournal, oldCommodities, oldPortfolios })
	updateJournal, updateCommodities, updatePortfolios = false, true, false
	body, err := json.Marshal(map[string]interface{}{
		"journal_path": filepath.Join(t.TempDir(), "main.ledger"),
		"db_path":      filepath.Join(t.TempDir(), "paisa.db"),
		"commodities": []config.Commodity{{
			Name: "APPLE", Type: config.Stock, Price: config.Price{Provider: "com-yahoo", Code: "AAPL"},
		}},
	})
	require.NoError(t, err)
	require.NoError(t, config.LoadConfig(body, ""))
	oldClient := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: failedPriceTransport{}}
	t.Cleanup(func() { http.DefaultClient = oldClient })
	require.ErrorContains(t, updateCmd.RunE(updateCmd, nil), "fixture provider failure")
}
