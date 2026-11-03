package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const providerConfig = `journal_path: main.ledger
db_path: paisa.db
commodities:
  - name: com-purifiedbytes-nps
    type: unknown
    price:
      provider: %s
      code: %s
`

func TestNormalizePriceProviders(t *testing.T) {
	for _, tc := range []struct {
		provider, canonical, code string
	}{
		{"com-purifiedbytes-nps", "com-finbodhi-nps", "SM008002"},
		{"com-purifiedbytes-metal", "com-finbodhi-metal", "gold-999"},
		{"com-finbodhi-nps", "com-finbodhi-nps", "SM008002"},
		{"com-finbodhi-metal", "com-finbodhi-metal", "gold-999"},
		{"in-mfapi", "in-mfapi", "120716"},
		{"in-mfapi", "in-mfapi", "012345"},
	} {
		t.Run(tc.provider+"/"+tc.code, func(t *testing.T) {
			oldConfig, oldPath, oldLocation := config, configPath, location
			t.Cleanup(func() { config, configPath, location = oldConfig, oldPath, oldLocation })
			configPath = ""

			path := filepath.Join(t.TempDir(), "paisa.yaml")
			content := fmt.Appendf(nil, providerConfig, tc.provider, tc.code)
			require.NoError(t, os.WriteFile(path, content, 0600))
			require.NoError(t, LoadConfig(content, path))
			commodity := GetConfig().Commodities[0]
			require.Equal(t, tc.canonical, commodity.Price.Provider)
			require.Equal(t, tc.code, commodity.Price.Code)
			require.Equal(t, "com-purifiedbytes-nps", commodity.Name)

			unchanged, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, content, unchanged, "loading must not rewrite the file")

			require.NoError(t, SaveConfig(content))
			content, err = os.ReadFile(path)
			require.NoError(t, err)
			var saved Config
			require.NoError(t, yaml.Unmarshal(content, &saved))
			require.Equal(t, tc.canonical, saved.Commodities[0].Price.Provider)
			require.Equal(t, tc.code, saved.Commodities[0].Price.Code)
			require.Equal(t, commodity.Name, saved.Commodities[0].Name)
		})
	}
}

func TestNormalizePriceProvidersPreservesValidation(t *testing.T) {
	for _, tc := range []struct {
		name, provider, extra string
	}{
		{"unknown provider", "com-purifiedbytes-unknown", ""},
		{"unknown field", "com-purifiedbytes-nps", "unexpected_field: true\n"},
		{"invalid provider type", "[com-purifiedbytes-nps]", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content := fmt.Sprintf(providerConfig, tc.provider, "SM008002") + tc.extra
			require.Error(t, LoadConfig([]byte(content), ""))
		})
	}
}
