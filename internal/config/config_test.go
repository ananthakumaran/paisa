package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// M2-A removed the India-specific harvest / tax_category / schedule_al
// / CII config surface, plus the nps + metal CommodityType values and
// the in-mfapi / com-purifiedbytes-* price providers. Existing user
// yaml files still carry these keys, and the schema is locked down
// with `additionalProperties: false` at the root and inside
// `commodities[*]` — so a literal upgrade without graceful migration
// would crash `paisa update` / `paisa serve` with a schema error.
//
// LoadConfig must therefore silently strip the deprecated surface
// (emitting an INFO log per stripped field), then validate. These
// tests pin that behavior — if they fail, every existing user with
// a pre-M2-A config will hit a hard log.Fatal on first run.

const baseConfig = `
journal_path: main.ledger
db_path: paisa.db
`

func TestLoadConfig_StripsRootScheduleAL(t *testing.T) {
	yaml := baseConfig + `
schedule_al:
  - code: bank
    accounts:
      - Assets:Checking
  - code: share
    accounts:
      - Assets:Equity:*
`
	err := LoadConfig([]byte(yaml), "")
	assert.NoError(t, err, "legacy schedule_al must not crash schema validation")
}

func TestLoadConfig_StripsRootCII(t *testing.T) {
	yaml := baseConfig + `
cii:
  - year: 2023-24
    index: 348
`
	err := LoadConfig([]byte(yaml), "")
	assert.NoError(t, err, "legacy cii block must not crash schema validation")
}

func TestLoadConfig_StripsHarvestAndTaxCategory(t *testing.T) {
	yaml := baseConfig + `
commodities:
  - name: AAPL
    type: stock
    price:
      provider: com-yahoo
      code: AAPL
    harvest: 365
    tax_category: equity
`
	err := LoadConfig([]byte(yaml), "")
	assert.NoError(t, err, "legacy harvest/tax_category on a commodity must not crash schema validation")

	// The commodity itself should survive (only the deprecated keys are
	// stripped), so the typed Config should still expose it.
	c := GetConfig()
	assert.Len(t, c.Commodities, 1, "non-deprecated commodity must be preserved")
	assert.Equal(t, "AAPL", c.Commodities[0].Name)
}

func TestLoadConfig_DropsNPSCommodityEntirely(t *testing.T) {
	yaml := baseConfig + `
commodities:
  - name: AAPL
    type: stock
    price:
      provider: com-yahoo
      code: AAPL
  - name: NPS_HDFC_E
    type: nps
    price:
      provider: com-purifiedbytes-nps
      code: SM008001
`
	err := LoadConfig([]byte(yaml), "")
	assert.NoError(t, err, "legacy NPS commodity must not crash schema validation")

	c := GetConfig()
	assert.Len(t, c.Commodities, 1, "NPS commodity must be dropped, AAPL kept")
	assert.Equal(t, "AAPL", c.Commodities[0].Name)
}

func TestLoadConfig_DropsMetalCommodityEntirely(t *testing.T) {
	yaml := baseConfig + `
commodities:
  - name: Gold
    type: metal
    price:
      provider: com-purifiedbytes-metal
      code: gold-999
`
	err := LoadConfig([]byte(yaml), "")
	assert.NoError(t, err, "legacy metal commodity must not crash schema validation")

	c := GetConfig()
	assert.Empty(t, c.Commodities, "metal commodity must be dropped")
}

func TestLoadConfig_DropsInMfapiByProviderOnly(t *testing.T) {
	// Some legacy configs used in-mfapi with type: mutualfund (still a
	// supported type), so the type-only check wouldn't catch it. The
	// provider-code check must.
	yaml := baseConfig + `
commodities:
  - name: PPFAS
    type: mutualfund
    price:
      provider: in-mfapi
      code: 122639
`
	err := LoadConfig([]byte(yaml), "")
	assert.NoError(t, err, "legacy in-mfapi commodity must not crash schema validation")

	c := GetConfig()
	assert.Empty(t, c.Commodities, "in-mfapi commodity must be dropped")
}

func TestLoadConfig_StripCombined(t *testing.T) {
	// End-to-end: a realistic legacy yaml carrying every deprecated
	// surface at once. Must load cleanly.
	yaml := baseConfig + `
schedule_al:
  - code: bank
    accounts: [Assets:Checking]
commodities:
  - name: AAPL
    type: stock
    price:
      provider: com-yahoo
      code: AAPL
    tax_category: unlisted_equity
  - name: NIFTY
    type: mutualfund
    price:
      provider: in-mfapi
      code: 120716
    harvest: 365
    tax_category: equity65
  - name: NPS_HDFC_E
    type: nps
    price:
      provider: com-purifiedbytes-nps
      code: SM008001
`
	err := LoadConfig([]byte(yaml), "")
	assert.NoError(t, err, "combined legacy fields must not crash schema validation")

	c := GetConfig()
	assert.Len(t, c.Commodities, 1, "only AAPL should survive scrubbing")
	assert.Equal(t, "AAPL", c.Commodities[0].Name)
}

// A clean post-M2-A yaml must of course still load.
func TestLoadConfig_ModernConfigStillLoads(t *testing.T) {
	yaml := baseConfig + `
commodities:
  - name: "600519"
    type: stock
    price:
      provider: cn-eastmoney
      code: "0.600519"
`
	err := LoadConfig([]byte(yaml), "")
	assert.NoError(t, err)
	c := GetConfig()
	assert.Len(t, c.Commodities, 1)
	assert.Equal(t, "600519", c.Commodities[0].Name)
	assert.Equal(t, "cn-eastmoney", c.Commodities[0].Price.Provider)
}
