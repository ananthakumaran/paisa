package accounting

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const acme = "Assets:Equity:ACME"

type entry struct {
	transaction string
	date        string
	account     string
	commodity   string
	quantity    float64
	amount      float64
}

type expectedLot struct {
	date     string
	quantity float64
	amount   float64
}

func openTestDB(t *testing.T, entries []entry) *gorm.DB {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, config.LoadConfig([]byte("journal_path: main.ledger\ndb_path: paisa.db"), filepath.Join(dir, "paisa.yaml")))
	db, err := gorm.Open(sqlite.Open(filepath.Join(dir, "test.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&posting.Posting{}))
	t.Cleanup(func() {
		transaction.ClearCache()
		conn, err := db.DB()
		require.NoError(t, err)
		require.NoError(t, conn.Close())
	})

	for _, e := range entries {
		date, err := time.ParseInLocation("2006-01-02", e.date, config.TimeZone())
		require.NoError(t, err)
		require.NoError(t, db.Create(&posting.Posting{
			TransactionID: e.transaction,
			Date:          date,
			Account:       e.account,
			Commodity:     e.commodity,
			Quantity:      decimal.NewFromFloat(e.quantity),
			Amount:        decimal.NewFromFloat(e.amount),
		}).Error)
	}
	transaction.ClearCache()
	return db
}

func matchLots(t *testing.T, entries []entry) ([]Sale, []Lot) {
	t.Helper()
	db := openTestDB(t, entries)
	return MatchLots(db, query.Init(db).AccountPrefix(acme).All())
}

func assertLots(t *testing.T, expected []expectedLot, lots []Lot) {
	t.Helper()
	require.Len(t, lots, len(expected))
	for i, e := range expected {
		assert.Equal(t, e.date, lots[i].Date.Format("2006-01-02"))
		assert.InDelta(t, e.quantity, lots[i].Quantity.InexactFloat64(), 1e-9)
		assert.InDelta(t, e.amount, lots[i].Amount.InexactFloat64(), 1e-9)
	}
}

func buy(id string, date string, quantity float64, amount float64) []entry {
	return []entry{
		{id, date, acme, "ACME", quantity, amount},
		{id, date, "Assets:Checking", "INR", -amount, -amount},
	}
}

func sell(id string, date string, quantity float64, amount float64) []entry {
	return buy(id, date, -quantity, -amount)
}

func splitUnits(id string, date string, removed float64, added float64) []entry {
	// valued at market price
	return []entry{
		{id, date, acme, "ACME", -removed, -removed * 1000},
		{id, date, acme, "ACME", added, removed * 1000},
	}
}

func bonus(id string, date string, quantity float64, income string) []entry {
	return []entry{
		{id, date, acme, "ACME", quantity, quantity * 1000},
		{id, date, income, "ACME", -quantity, -quantity * 1000},
	}
}

func concat(entries ...[]entry) []entry {
	var all []entry
	for _, e := range entries {
		all = append(all, e...)
	}
	return all
}

func TestMatchLotsFIFO(t *testing.T) {
	sales, available := matchLots(t, concat(
		buy("1", "2019-01-01", 10, 1000),
		buy("2", "2019-02-01", 10, 2000),
		sell("3", "2020-01-01", 15, 4500)))

	require.Len(t, sales, 1)
	assertLots(t, []expectedLot{{"2019-01-01", 10, 1000}, {"2019-02-01", 5, 1000}}, sales[0].Lots)
	assertLots(t, []expectedLot{{"2019-02-01", 5, 1000}}, available)
}

func TestMatchLotsSplit(t *testing.T) {
	for _, tc := range []struct {
		name  string
		split []entry
	}{
		{"whole holding", splitUnits("3", "2019-06-01", 20, 40)},
		{"per lot", concat(splitUnits("3", "2019-06-01", 10, 20), splitUnits("4", "2019-06-01", 10, 20))},
		{"partial units", splitUnits("3", "2019-06-01", 1, 21)},
		{"more than holding", splitUnits("3", "2019-06-01", 30, 60)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sales, available := matchLots(t, concat(
				buy("1", "2019-01-01", 10, 1000),
				buy("2", "2019-02-01", 10, 2000),
				tc.split,
				sell("5", "2020-01-01", 25, 4500)))

			require.Len(t, sales, 1, "split should not be considered as a sale")
			assertLots(t, []expectedLot{{"2019-01-01", 20, 1000}, {"2019-02-01", 5, 500}}, sales[0].Lots)
			assertLots(t, []expectedLot{{"2019-02-01", 15, 1500}}, available)
		})
	}
}

func TestMatchLotsReverseSplitDoesNotLoseUnits(t *testing.T) {
	sales, available := matchLots(t, concat(
		buy("1", "2019-01-01", 40, 1200),
		buy("2", "2019-02-01", 50, 3000),
		splitUnits("3", "2019-06-01", 90, 30),
		sell("4", "2020-01-01", 30, 4500)))

	require.Len(t, sales, 1)
	assertLots(t, []expectedLot{{"2019-01-01", 40.0 / 3, 1200}, {"2019-02-01", 50.0 / 3, 3000}}, sales[0].Lots)
	assert.Empty(t, available)
}

func TestMatchLotsSplitOnSameDay(t *testing.T) {
	// trades happen at the adjusted price
	_, available := matchLots(t, concat(
		buy("1", "2019-01-01", 10, 1000),
		buy("2", "2019-06-01", 10, 500),
		splitUnits("3", "2019-06-01", 10, 20)))
	assertLots(t, []expectedLot{{"2019-01-01", 20, 1000}, {"2019-06-01", 10, 500}}, available)

	// not enough units to split before the trade
	_, available = matchLots(t, concat(
		buy("1", "2019-06-01", 10, 1000),
		splitUnits("2", "2019-06-01", 10, 20)))
	assertLots(t, []expectedLot{{"2019-06-01", 20, 1000}}, available)

	// order of the journal is retained
	_, available = matchLots(t, concat(
		buy("1", "2019-01-01", 10, 1000),
		bonus("2", "2019-06-01", 10, "Income:Bonus:ACME"),
		splitUnits("3", "2019-06-01", 20, 40)))
	assertLots(t, []expectedLot{{"2019-01-01", 20, 1000}, {"2019-06-01", 20, 0}}, available)
}

func TestSplitRatioAfter(t *testing.T) {
	_, available := matchLots(t, concat(
		buy("1", "2017-01-01", 10, 1000),
		splitUnits("2", "2017-06-01", 10, 50),
		splitUnits("3", "2019-06-01", 50, 100),
		splitUnits("4", "2020-06-01", 100, 300)))

	require.Len(t, available, 1)
	date := time.Date(2018, 2, 1, 0, 0, 0, 0, config.TimeZone())
	assert.Equal(t, 6.0, available[0].SplitRatioAfter(date).InexactFloat64())
}

func TestMatchLotsBonus(t *testing.T) {
	for _, tc := range []struct {
		name  string
		bonus []entry
		lot   expectedLot
	}{
		{"against income", bonus("2", "2019-06-01", 10, "Income:Bonus:ACME"), expectedLot{"2019-06-01", 10, 0}},
		{"zero price without income posting", []entry{{"2", "2019-06-01", acme, "ACME", 10, 5000}}, expectedLot{"2019-06-01", 10, 0}},
		{"zero price with income posting", []entry{
			{"2", "2019-06-01", acme, "ACME", 10, 10},
			{"2", "2019-06-01", "Income:Bonus:ACME", "INR", 0, 0}}, expectedLot{"2019-06-01", 10, 0}},
		{"priced income", []entry{
			{"2", "2019-06-01", acme, "ACME", 10, 5000},
			{"2", "2019-06-01", "Income:Bonus:ACME", "INR", -5000, -5000}}, expectedLot{"2019-06-01", 10, 5000}},
		{"salary", bonus("2", "2019-06-01", 10, "Income:Salary:Acme"), expectedLot{"2019-06-01", 10, 10000}},
		{"dividend", bonus("2", "2019-06-01", 10, "Income:Dividend:ACME"), expectedLot{"2019-06-01", 10, 10000}},
		{"opening balance", bonus("2", "2019-06-01", 10, "Equity:OpeningBalance"), expectedLot{"2019-06-01", 10, 10000}},
		{"transfer", bonus("2", "2019-06-01", 10, "Assets:Broker:ACME"), expectedLot{"2019-06-01", 10, 10000}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, available := matchLots(t, concat(buy("1", "2019-01-01", 10, 1000), tc.bonus))
			assertLots(t, []expectedLot{{"2019-01-01", 10, 1000}, tc.lot}, available)
		})
	}
}

func TestMatchLotsSellAmount(t *testing.T) {
	for _, tc := range []struct {
		name     string
		sale     []entry
		expected []float64
	}{
		{"sell price", sell("2", "2020-01-01", 10, 1500), []float64{-1500}},
		{"cost price and gain", []entry{
			{"2", "2020-01-01", acme, "ACME", -10, -1000},
			{"2", "2020-01-01", "Income:CapitalGains:Equity:ACME", "INR", -500, -500},
			{"2", "2020-01-01", "Assets:Checking", "INR", 1500, 1500}}, []float64{-1500}},
		{"cost price and loss", []entry{
			{"2", "2020-01-01", acme, "ACME", -10, -1000},
			{"2", "2020-01-01", "Income:CapitalGains:Equity:ACME", "INR", 200, 200},
			{"2", "2020-01-01", "Expenses:Charges", "INR", 10, 10},
			{"2", "2020-01-01", "Assets:Checking", "INR", 790, 790}}, []float64{-800}},
		{"multiple lots", []entry{
			{"2", "2020-01-01", acme, "ACME", -6, -600},
			{"2", "2020-01-01", acme, "ACME", -4, -100},
			{"2", "2020-01-01", "Income:CapitalGains:Equity:ACME", "INR", -800, -800},
			{"2", "2020-01-01", "Assets:Checking", "INR", 1500, 1500}}, []float64{-600, -900}},
		{"sell price and gain", []entry{
			{"2", "2020-01-01", acme, "ACME", -10, -1500},
			{"2", "2020-01-01", "Income:CapitalGains:Equity:ACME", "INR", -500, -500},
			{"2", "2020-01-01", "Assets:Checking", "INR", 1500, 1500}}, []float64{-1500}},
		{"gain of another account", []entry{
			{"2", "2020-01-01", acme, "ACME", -10, -1000},
			{"2", "2020-01-01", "Income:CapitalGains:Equity:BETA", "INR", -500, -500},
			{"2", "2020-01-01", "Assets:Checking", "INR", 1500, 1500}}, []float64{-1000}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sales, available := matchLots(t, concat(buy("1", "2019-01-01", 10, 1000), tc.sale))
			require.Len(t, sales, len(tc.expected))
			actual := []float64{}
			for _, s := range sales {
				actual = append(actual, s.Posting.Amount.InexactFloat64())
			}
			assert.ElementsMatch(t, tc.expected, actual)
			assert.Empty(t, available)
		})
	}
}
