package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func setupPriceConfig(t *testing.T) {
	t.Helper()
	oldConfig, oldPath, oldLocation := config, configPath, location
	t.Cleanup(func() { config, configPath, location = oldConfig, oldPath, oldLocation })
	configPath = ""
	require.NoError(t, LoadConfig([]byte("journal_path: main.ledger\ndb_path: paisa.db\ntime_zone: UTC\n"), filepath.Join(t.TempDir(), "paisa.yaml")))
}

func configuredPrice(name, currency string) Commodity {
	return Commodity{Name: name, Type: Currency, Price: Price{Provider: "com-yahoo", Code: name, Commodity: currency}}
}

func TestOrderedPriceCommodities(t *testing.T) {
	usd := configuredPrice("USD", "")
	eur := configuredPrice("EUR", "USD")
	apple := configuredPrice("APPLE", "USD")
	for name, tc := range map[string]struct {
		sources []Commodity
		order   []string
		error   string
	}{
		"default currency": {sources: []Commodity{configuredPrice("LOCAL", "")}, order: []string{"LOCAL"}},
		"shared and nested sources": {
			sources: []Commodity{apple, configuredPrice("EURO_STOCK", "EUR"), configuredPrice("OTHER", "USD"), eur, usd},
			order:   []string{"USD", "APPLE", "EUR", "EURO_STOCK", "OTHER"},
		},
		"default source is terminal": {
			sources: []Commodity{configuredPrice("APPLE", "INR"), configuredPrice("INR", "MISSING")},
			order:   []string{"APPLE"},
		},
		"self cycle":     {sources: []Commodity{configuredPrice("USD", "USD")}, error: "USD -> USD"},
		"duplicate":      {sources: []Commodity{usd, usd}, error: "Duplicate commodity price source"},
		"missing source": {sources: []Commodity{apple}, error: "USD has no configured price source"},
		"cycle":          {sources: []Commodity{apple, configuredPrice("USD", "EUR"), eur}, error: "USD -> EUR -> USD"},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := Config{DefaultCurrency: "INR", Commodities: tc.sources}
			ordered, err := candidate.OrderedPriceCommodities()
			if tc.error != "" {
				require.ErrorContains(t, err, tc.error)
				return
			}
			require.NoError(t, err)
			var names []string
			for _, source := range ordered {
				names = append(names, source.Name)
			}
			require.Equal(t, tc.order, names)
		})
	}
}

func TestPriceCommoditySerialization(t *testing.T) {
	setupPriceConfig(t)
	candidate := GetConfig()
	candidate.Commodities = []Commodity{configuredPrice("APPLE", "USD"), configuredPrice("USD", "")}
	require.NoError(t, SaveConfigObject(candidate))
	content, err := os.ReadFile(GetConfigPath())
	require.NoError(t, err)
	require.Contains(t, string(content), "commodity: USD")
	require.NotContains(t, string(content), `commodity: ""`)
	require.NoError(t, LoadConfig(content, ""))
	require.Equal(t, candidate.Commodities, GetConfig().Commodities)

	encoded, err := json.Marshal(candidate.Commodities[1].Price)
	require.NoError(t, err)
	require.JSONEq(t, `{"provider":"com-yahoo","code":"USD"}`, string(encoded))
	require.Equal(t, "INR", (Price{}).QuoteCommodity("INR"))
	require.Equal(t, "USD", (Price{}).QuoteCommodity("USD"))
	require.Equal(t, "USD", (Price{Commodity: "USD"}).QuoteCommodity("INR"))
}

func TestInvalidPriceConfigurationPreservesActiveState(t *testing.T) {
	setupPriceConfig(t)
	original := GetConfig()
	require.NoError(t, SaveConfigObject(original))
	before, err := os.ReadFile(GetConfigPath())
	require.NoError(t, err)
	candidate := original
	candidate.Commodities = []Commodity{configuredPrice("APPLE", "USD")}
	candidate.TimeZone = "Asia/Kolkata"
	content, err := yaml.Marshal(candidate)
	require.NoError(t, err)

	require.Error(t, SaveConfig(content))
	require.Error(t, LoadConfig(content, "another.yaml"))
	require.Equal(t, original, GetConfig(), "invalid settings must not replace the active configuration")
	require.Equal(t, "UTC", TimeZone().String())
	after, err := os.ReadFile(GetConfigPath())
	require.NoError(t, err)
	require.Equal(t, before, after, "invalid settings must not overwrite the file")
}

func TestPriceValidationUsesCandidateDefaultCurrency(t *testing.T) {
	setupPriceConfig(t)
	candidate := GetConfig()
	candidate.DefaultCurrency = "USD"
	candidate.Commodities = []Commodity{configuredPrice("APPLE", "USD")}
	require.NoError(t, SaveConfigObject(candidate))
	require.Equal(t, "USD", DefaultCurrency())

	candidate.DefaultCurrency = "INR"
	require.ErrorContains(t, SaveConfigObject(candidate), "USD has no configured price source")
	require.Equal(t, "USD", DefaultCurrency())
}

func TestFailedConfigWritePreservesActiveState(t *testing.T) {
	setupPriceConfig(t)
	original := GetConfig()
	candidate := original
	candidate.DefaultCurrency = "USD"
	configPath = t.TempDir()
	require.Error(t, SaveConfigObject(candidate))
	require.Equal(t, original, GetConfig())
}
