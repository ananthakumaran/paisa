package ttjj

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/ananthakumaran/paisa/internal/config"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return body
}

func TestParseNetWorthTrend(t *testing.T) {
	body := readFixture(t, "000311.js")

	prices, err := parseNetWorthTrend(body, "000311", "HS300")
	assert.NoError(t, err)
	assert.Len(t, prices, 3)

	assert.Equal(t, "000311", prices[0].CommodityID)
	assert.Equal(t, "HS300", prices[0].CommodityName)
	assert.Equal(t, config.MutualFund, prices[0].CommodityType)

	// Timestamps in the fixture (Asia/Shanghai midnight):
	//   1706198400000 -> 2024-01-26
	//   1706284800000 -> 2024-01-27
	//   1706457600000 -> 2024-01-29
	assert.Equal(t, 2024, prices[0].Date.Year())
	assert.Equal(t, "January", prices[0].Date.Month().String())
	assert.Equal(t, 26, prices[0].Date.Day())
	assert.Equal(t, 27, prices[1].Date.Day())
	assert.Equal(t, 29, prices[2].Date.Day())

	assert.True(t, decimal.NewFromFloat(1.2345).Equal(prices[0].Value), "got %s", prices[0].Value)
	assert.True(t, decimal.NewFromFloat(1.2400).Equal(prices[1].Value), "got %s", prices[1].Value)
	assert.True(t, decimal.NewFromFloat(1.2380).Equal(prices[2].Value), "got %s", prices[2].Value)
}

func TestParseNetWorthTrend_EmptyArray(t *testing.T) {
	// 清盘 (delisted) funds expose an empty array; we want no error.
	body := readFixture(t, "empty.js")

	prices, err := parseNetWorthTrend(body, "999999", "DEAD")
	assert.NoError(t, err)
	assert.Empty(t, prices)
}

func TestParseNetWorthTrend_Missing(t *testing.T) {
	body := []byte(`var fS_name = "foo"; var something = 1;`)

	_, err := parseNetWorthTrend(body, "000000", "X")
	assert.Error(t, err)
}

func TestProviderMetadata(t *testing.T) {
	p := &PriceProvider{}
	assert.Equal(t, "cn-ttjj", p.Code())
	assert.NotEmpty(t, p.Label())
	assert.NotEmpty(t, p.Description())
	assert.NotEmpty(t, p.AutoCompleteFields())
}
