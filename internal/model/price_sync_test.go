package model

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func priceAt(timestamp int64, value string) *price.Price {
	return &price.Price{Date: time.Unix(timestamp, 0).UTC(), Value: decimal.RequireFromString(value)}
}

func yahooSource(name, code, currency string, kind config.CommodityType) config.Commodity {
	return config.Commodity{Name: name, Type: kind, Price: config.Price{Provider: "com-yahoo", Code: code, Commodity: currency}}
}

func syncTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	AutoMigrate(db)
	return db
}

func configurePriceSync(t *testing.T, currency string, sources ...config.Commodity) {
	t.Helper()
	original, err := json.Marshal(config.GetConfig())
	require.NoError(t, err)
	if config.GetConfig().JournalPath == "" {
		original = []byte(`{"journal_path":"main.ledger","db_path":"paisa.db"}`)
	}
	t.Cleanup(func() { require.NoError(t, config.LoadConfig(original, "")) })
	body, err := json.Marshal(map[string]interface{}{
		"journal_path": "main.ledger", "db_path": "paisa.db", "time_zone": "UTC",
		"default_currency": currency, "commodities": sources,
	})
	require.NoError(t, err)
	require.NoError(t, config.LoadConfig(body, ""))
}

type priceTransport struct {
	responses map[string]string
	requests  []string
}

func (p *priceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	code := strings.TrimPrefix(req.URL.Path, "/v8/finance/chart/")
	p.requests = append(p.requests, code)
	body, exists := p.responses[code]
	if !exists {
		return nil, fmt.Errorf("price unavailable: %s", code)
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
}

func mockPriceRequests(t *testing.T, responses map[string]string) *priceTransport {
	t.Helper()
	transport := &priceTransport{responses: responses}
	original := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: transport}
	t.Cleanup(func() { http.DefaultClient = original })
	return transport
}

func chart(timestamps, closes string) string {
	return fmt.Sprintf(`{"chart":{"result":[{"timestamp":%s,"meta":{"currency":"IGNORED"},"indicators":{"quote":[{"close":%s}]}}]}}`, timestamps, closes)
}

func priceValues(prices []*price.Price) []string {
	values := make([]string, 0, len(prices))
	for _, p := range prices {
		values = append(values, p.Value.String())
	}
	return values
}

func storedValues(t *testing.T, db *gorm.DB, name string) []string {
	t.Helper()
	var prices []*price.Price
	require.NoError(t, db.Where("commodity_name = ?", name).Order("date").Find(&prices).Error)
	return priceValues(prices)
}

func TestNormalizeCommodityPrices(t *testing.T) {
	asset := yahooSource("APPLE", "AAPL", "USD", config.Stock)
	raw := []*price.Price{priceAt(400, "3"), priceAt(50, "1"), priceAt(200, "2"), priceAt(25, "4")}
	rates := newPriceHistory([]*price.Price{priceAt(100, "80"), priceAt(300, "90")})

	prices, gap, err := normalizeCommodityPrices(asset, raw, rates, "INR")
	require.NoError(t, err)
	require.Equal(t, []string{"270", "160"}, priceValues(prices))
	require.Equal(t, priceCoverageGap{
		Count: 2,
		From:  time.Unix(25, 0).UTC(),
		To:    time.Unix(50, 0).UTC(),
	}, gap)
	require.Equal(t, priceAt(400, "3"), raw[0], "keep provider output unchanged")

	t.Run("default currency needs no exchange rate", func(t *testing.T) {
		for _, currency := range []string{"", "INR"} {
			asset.Price.Commodity = currency
			prices, gap, err := normalizeCommodityPrices(asset, raw, priceHistory{}, "INR")
			require.NoError(t, err)
			require.Equal(t, priceValues(raw), priceValues(prices))
			require.Zero(t, gap.Count)
		}
	})

	t.Run("zero prices and rates are valid", func(t *testing.T) {
		asset.Price.Commodity = "USD"
		raw := []*price.Price{priceAt(100, "0"), priceAt(250, "5")}
		rates := newPriceHistory([]*price.Price{priceAt(100, "2"), priceAt(200, "0")})
		prices, gap, err := normalizeCommodityPrices(asset, raw, rates, "INR")
		require.NoError(t, err)
		require.Equal(t, []string{"0", "0"}, priceValues(prices))
		require.Zero(t, gap.Count)
	})

	t.Run("missing and future rates cannot convert prices", func(t *testing.T) {
		asset.Price.Commodity = "USD"
		for _, rates := range []priceHistory{{}, newPriceHistory([]*price.Price{priceAt(1000, "80")})} {
			_, gap, err := normalizeCommodityPrices(asset, raw, rates, "INR")
			require.ErrorContains(t, err, "no historical USD conversion rates")
			require.Equal(t, len(raw), gap.Count)
		}
	})
}

func TestSyncCommoditiesSharedDependenciesAndRefresh(t *testing.T) {
	sources := []config.Commodity{
		yahooSource("APPLE", "AAPL", "USD", config.Stock),
		yahooSource("OTHER", "MSFT", "USD", config.Stock),
		yahooSource("EURO_STOCK", "EURO", "EUR", config.Stock),
		yahooSource("USD", "USDINR=X", "", config.Currency),
		yahooSource("EUR", "EURUSD=X", "USD", config.Currency),
	}
	configurePriceSync(t, "INR", sources...)
	db := syncTestDB(t)
	responses := map[string]string{
		"USDINR=X": chart("[100,300]", "[80,90]"),
		"AAPL":     chart("[50,200,400]", "[1,2,3]"),
		"MSFT":     chart("[200]", "[4]"),
		"EURUSD=X": chart("[100]", "[2]"),
		"EURO":     chart("[200]", "[5]"),
	}
	fetches := mockPriceRequests(t, responses)

	warnings, err := SyncCommodities(db)
	require.NoError(t, err)
	require.Len(t, warnings, 1)
	require.Contains(t, warnings[0], "APPLE: skipped 1")
	require.Equal(t, []string{"USDINR=X", "AAPL", "MSFT", "EURUSD=X", "EURO"}, fetches.requests)
	require.Equal(t, []string{"160", "270"}, storedValues(t, db, "APPLE"))
	require.Equal(t, []string{"320"}, storedValues(t, db, "OTHER"))
	require.Equal(t, []string{"800"}, storedValues(t, db, "EURO_STOCK"))
	var currency price.Price
	require.NoError(t, db.Where("commodity_name = ?", "USD").First(&currency).Error)
	require.Equal(t, config.Currency, currency.CommodityType)

	responses["USDINR=X"] = chart("[100,300]", "[100,110]")
	_, err = SyncCommodities(db)
	require.NoError(t, err)
	require.Len(t, fetches.requests, 10, "fetch each source once per update")
	require.Equal(t, []string{"200", "330"}, storedValues(t, db, "APPLE"))
	require.Equal(t, []string{"1000"}, storedValues(t, db, "EURO_STOCK"))

	sources[0].Price.Commodity = "EUR"
	configurePriceSync(t, "INR", sources...)
	_, err = SyncCommodities(db)
	require.NoError(t, err)
	require.Equal(t, []string{"400", "600"}, storedValues(t, db, "APPLE"))

	configurePriceSync(t, "USD", sources...)
	fetches.requests = nil
	_, err = SyncCommodities(db)
	require.NoError(t, err)
	require.NotContains(t, fetches.requests, "USDINR=X")
	require.Equal(t, []string{"4", "6"}, storedValues(t, db, "APPLE"))
	require.Equal(t, []string{"10"}, storedValues(t, db, "EURO_STOCK"))
}

func TestSyncCommoditiesFailedDependency(t *testing.T) {
	for _, failure := range []string{"fetch", "empty history", "store"} {
		t.Run(failure, func(t *testing.T) {
			configurePriceSync(t, "INR",
				yahooSource("APPLE", "AAPL", "USD", config.Stock),
				yahooSource("LOCAL", "LOCAL", "", config.Stock),
				yahooSource("USD", "USDINR=X", "", config.Currency))
			db := syncTestDB(t)
			old := priceAt(1, "999")
			old.CommodityName, old.CommodityType = "USD", config.Currency
			require.NoError(t, db.Create(old).Error)
			responses := map[string]string{"LOCAL": chart("[200]", "[10]")}
			if failure == "empty history" {
				responses["USDINR=X"] = chart("[]", "[]")
			}
			if failure == "store" {
				responses["USDINR=X"] = chart("[100]", "[80]")
				require.NoError(t, db.Exec(`CREATE TRIGGER fail_usd BEFORE INSERT ON prices
					WHEN NEW.commodity_name = 'USD' BEGIN SELECT RAISE(FAIL, 'forced failure'); END`).Error)
			}
			fetches := mockPriceRequests(t, responses)

			_, err := SyncCommodities(db)
			require.ErrorContains(t, err, "Skipped price for APPLE")
			require.Equal(t, []string{"USDINR=X", "LOCAL"}, fetches.requests)
			require.Empty(t, storedValues(t, db, "APPLE"))
			require.Equal(t, []string{"10"}, storedValues(t, db, "LOCAL"))
			require.Equal(t, []string{"999"}, storedValues(t, db, "USD"), "never use stale rates after a failed fetch")
		})
	}
}

func TestSyncCommoditiesIndianProviders(t *testing.T) {
	fund := config.Commodity{Name: "FUND", Type: config.MutualFund, Price: config.Price{Provider: "in-mfapi", Code: "120716", Commodity: "INR"}}
	nps := config.Commodity{Name: "NPS", Type: config.NPS, Price: config.Price{Provider: "com-finbodhi-nps", Code: "SM008001", Commodity: "INR"}}
	metal := config.Commodity{Name: "GOLD", Type: config.Metal, Price: config.Price{Provider: "com-finbodhi-metal", Code: "gold-999", Commodity: "INR"}}
	configurePriceSync(t, "USD", fund, nps, metal, yahooSource("INR", "INRUSD=X", "", config.Currency))
	db := syncTestDB(t)
	fetches := mockPriceRequests(t, map[string]string{
		"INRUSD=X":                       chart("[1703980800]", "[0.012]"),
		"/mf/120716":                     `{"data":[{"date":"01-01-2024","nav":"100"}]}`,
		"/api/schemes/SM008001/nav.json": `{"data":[{"date":"2024-01-01","nav":200}]}`,
		"/api/metal/gold-999/price.json": `{"data":[{"date":"2024-01-01","close":60000}]}`,
	})

	warnings, err := SyncCommodities(db)
	require.NoError(t, err)
	require.Empty(t, warnings)
	require.Len(t, fetches.requests, 4)
	require.Equal(t, []string{"1.2"}, storedValues(t, db, "FUND"))
	require.Equal(t, []string{"2.4"}, storedValues(t, db, "NPS"))
	require.Equal(t, []string{"72"}, storedValues(t, db, "GOLD"), "preserve the per-gram metal price")
}

func TestSyncCommoditiesValidatesBeforeFetching(t *testing.T) {
	configurePriceSync(t, "INR", yahooSource("APPLE", "AAPL", "", config.Stock))
	config.GetConfig().Commodities[0].Price.Commodity = "MISSING"
	fetches := mockPriceRequests(t, nil)

	_, err := SyncCommodities(nil)
	require.ErrorContains(t, err, "MISSING has no configured price source")
	require.Empty(t, fetches.requests)
}
