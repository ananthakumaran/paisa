package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/cache"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func serverPriceSource(name, code, currency string, kind config.CommodityType) config.Commodity {
	return config.Commodity{Name: name, Type: kind, Price: config.Price{Provider: "com-yahoo", Code: code, Commodity: currency}}
}

func priceRecord(name string, kind config.CommodityType, value string) *price.Price {
	return &price.Price{Date: time.Unix(200, 0).UTC(), CommodityName: name, CommodityType: kind, Value: decimal.RequireFromString(value)}
}

func setupPriceServer(t *testing.T, sources ...config.Commodity) *gorm.DB {
	t.Helper()
	original, err := json.Marshal(config.GetConfig())
	require.NoError(t, err)
	if config.GetConfig().JournalPath == "" {
		original = []byte(`{"journal_path":"main.ledger","db_path":"paisa.db"}`)
	}
	t.Cleanup(func() { require.NoError(t, config.LoadConfig(original, "")) })
	content, err := json.Marshal(map[string]interface{}{
		"journal_path": "main.ledger", "db_path": "paisa.db", "default_currency": "INR",
		"time_zone": "UTC", "commodities": sources,
	})
	require.NoError(t, err)
	require.NoError(t, config.LoadConfig(content, ""))
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	model.AutoMigrate(db)
	cache.Clear()
	t.Cleanup(cache.Clear)
	return db
}

func TestPricePageIncludesCurrenciesWithoutPostings(t *testing.T) {
	db := setupPriceServer(t,
		serverPriceSource("IDR", "IDRINR=X", "", config.Currency),
		serverPriceSource("EMPTY", "EMPTY", "", config.Stock))
	require.NoError(t, db.Create([]*price.Price{
		priceRecord("IDR", config.Currency, "0.005"),
		priceRecord("JOURNAL", config.Unknown, "5"),
	}).Error)
	require.NoError(t, db.Create(&posting.Posting{Commodity: "JOURNAL"}).Error)

	prices := GetPrices(db)["prices"].(map[string][]price.Price)
	require.Len(t, prices["IDR"], 1)
	require.Equal(t, config.Currency, prices["IDR"][0].CommodityType)
	require.Equal(t, "0.005", prices["IDR"][0].Value.String())
	require.Len(t, prices["JOURNAL"], 1)
	require.Empty(t, prices["EMPTY"])
	require.Empty(t, service.GetAllPrices(db, "ABSENT"))

	require.NoError(t, db.Create(&posting.Posting{Commodity: "IDR"}).Error)
	require.NoError(t, db.Create(priceRecord("IDR", config.Unknown, "1")).Error)
	cache.Clear()
	prices = GetPrices(db)["prices"].(map[string][]price.Price)
	require.Len(t, prices["IDR"], 1, "merge journal and fetched prices without duplicates")
	require.Equal(t, "0.005", prices["IDR"][0].Value.String(), "prefer the fetched price")
}

type syncRoundTripFunc func(*http.Request) (*http.Response, error)

func (f syncRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSyncWarningsAndCacheInvalidation(t *testing.T) {
	for name, unavailable := range map[string]string{
		"success":               "",
		"partial price failure": "/v8/finance/chart/LOCAL",
		"CII failure":           "/api/cii/v2.json",
	} {
		t.Run(name, func(t *testing.T) {
			db := setupPriceServer(t,
				serverPriceSource("APPLE", "AAPL", "USD", config.Stock),
				serverPriceSource("USD", "USDINR=X", "", config.Currency),
				serverPriceSource("LOCAL", "LOCAL", "", config.Stock))
			old := priceRecord("APPLE", config.Stock, "10")
			require.NoError(t, db.Create(old).Error)
			responses := map[string]string{
				"/v8/finance/chart/USDINR=X": `{"chart":{"result":[{"timestamp":[100],"indicators":{"quote":[{"close":[80]}]}}]}}`,
				"/v8/finance/chart/AAPL":     `{"chart":{"result":[{"timestamp":[50,200],"indicators":{"quote":[{"close":[1,2]}]}}]}}`,
				"/v8/finance/chart/LOCAL":    `{"chart":{"result":[{"timestamp":[100],"indicators":{"quote":[{"close":[10]}]}}]}}`,
				"/api/cii/v2.json":           `{"data":[]}`,
			}
			delete(responses, unavailable)
			originalClient := http.DefaultClient
			http.DefaultClient = &http.Client{Transport: syncRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == "/v8/finance/chart/USDINR=X" {
					require.Equal(t, "10", service.GetUnitPrice(db, "APPLE", old.Date).Value.String())
				}
				body, exists := responses[req.URL.Path]
				if !exists {
					return nil, errors.New("fixture unavailable")
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
			})}
			t.Cleanup(func() { http.DefaultClient = originalClient })

			response := Sync(db, SyncRequest{Prices: true})
			require.Equal(t, unavailable == "", response["success"])
			warnings := response["warnings"].([]string)
			require.Len(t, warnings, 1)
			require.Contains(t, warnings[0], "APPLE: skipped 1")
			if unavailable != "" {
				require.Contains(t, response["message"], "unavailable")
			}
			require.Equal(t, "160", service.GetUnitPrice(db, "APPLE", old.Date).Value.String(), "refresh cached prices even after partial failure")
		})
	}
}

func TestCurrencyIsExcludedFromCapitalGains(t *testing.T) {
	currency := serverPriceSource("USD", "USDINR=X", "", config.Currency)
	currency.TaxCategory = config.Equity
	db := setupPriceServer(t, currency)
	require.NoError(t, db.Create(&posting.Posting{Commodity: "USD", Account: "Assets:ForeignCurrency"}).Error)
	require.Empty(t, GetCapitalGains(db)["capital_gains"])
}
