package taxation

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGrandfatherPriceIsAdjustedForSplit(t *testing.T) {
	date := func(value string) time.Time {
		d, err := time.ParseInLocation("2006-01-02", value, config.TimeZone())
		require.NoError(t, err)
		return d
	}

	for _, tc := range []struct {
		name     string
		provider string
		source   config.CommodityType
		taxable  float64
	}{
		{"yahoo is already adjusted", "com-yahoo", config.Stock, 10 * (500 - 400)},
		{"alphavantage", "co-alphavantage", config.Stock, 10 * (500 - 200)},
		{"journal price", "com-yahoo", config.Unknown, 10 * (500 - 200)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
			require.NoError(t, err)
			require.NoError(t, db.AutoMigrate(&posting.Posting{}, &price.Price{}))
			service.ClearPriceCache()
			t.Cleanup(func() {
				service.ClearPriceCache()
				conn, err := db.DB()
				require.NoError(t, err)
				require.NoError(t, conn.Close())
			})
			require.NoError(t, db.Create(&posting.Posting{Date: date("2017-06-01"), Commodity: "ACME"}).Error)
			require.NoError(t, db.Create(&price.Price{Date: date("2018-01-31"), CommodityType: tc.source, CommodityName: "ACME", Value: decimal.NewFromInt(400)}).Error)

			commodity := config.Commodity{Name: "ACME", TaxCategory: config.Equity, Price: config.Price{Provider: tc.provider}}
			tax := Calculate(db, decimal.NewFromInt(10), commodity, decimal.NewFromInt(150), date("2017-06-01"), decimal.NewFromInt(500), date("2020-01-01"), decimal.NewFromInt(2))
			assert.Equal(t, 3500.0, tax.Gain.InexactFloat64())
			assert.Equal(t, tc.taxable, tax.Taxable.InexactFloat64())
		})
	}
}
