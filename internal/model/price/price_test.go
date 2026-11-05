package price

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func priceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.AutoMigrate(&Price{}))
	return db
}

func priceRow(name string, kind config.CommodityType, value int64) *Price {
	return &Price{Date: time.Unix(100, 0), CommodityName: name, CommodityType: kind, CommodityID: "AAPL", Value: decimal.NewFromInt(value)}
}

func storedPrices(t *testing.T, db *gorm.DB) map[string][]string {
	t.Helper()
	var rows []Price
	require.NoError(t, db.Order("date").Find(&rows).Error)
	values := map[string][]string{}
	for _, row := range rows {
		source := row.CommodityName + "/" + string(row.CommodityType)
		values[source] = append(values[source], row.Value.String())
	}
	return values
}

func TestUpsertPrices(t *testing.T) {
	db := priceTestDB(t)
	alias := priceRow("ALIAS", config.Stock, 200)
	require.NoError(t, db.Create([]*Price{
		priceRow("APPLE", config.Stock, 100),
		priceRow("APPLE", config.Unknown, 50),
		alias,
	}).Error)

	updated := priceRow("APPLE", config.Stock, 300)
	require.NoError(t, UpsertAllByTypeAndName(db, config.Stock, "APPLE", []*Price{updated}))
	want := map[string][]string{
		"APPLE/stock":   {"300"},
		"APPLE/unknown": {"50"},
		"ALIAS/stock":   {"200"},
	}
	require.Equal(t, want, storedPrices(t, db), "replace only this source's fetched history")

	changedType := priceRow("APPLE", config.Currency, 300)
	require.NoError(t, UpsertAllByTypeAndName(db, config.Currency, "APPLE", []*Price{changedType}))
	delete(want, "APPLE/stock")
	want["APPLE/currency"] = []string{"300"}
	require.Equal(t, want, storedPrices(t, db), "remove the old type while keeping journal prices")

	bad := priceRow("APPLE", config.Currency, 999)
	bad.ID = alias.ID
	require.Error(t, UpsertAllByTypeAndName(db, config.Currency, "APPLE", []*Price{bad}))
	require.Equal(t, want, storedPrices(t, db), "roll back the entire replacement when an insert fails")
}
