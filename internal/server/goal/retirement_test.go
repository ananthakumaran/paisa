package goal

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/cache"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRetirementDetailExcludesInterestFromInvestmentTotal(t *testing.T) {
	if err := config.LoadConfig([]byte(`{"journal_path":"test.ledger","db_path":"test.db"}`), ""); err != nil {
		t.Fatal(err)
	}
	service.ClearInterestCache()
	service.ClearPriceCache()
	db, err := gorm.Open(sqlite.Open("file:retirement-interest?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&posting.Posting{}, &cache.Cache{}, &price.Price{}); err != nil {
		t.Fatal(err)
	}
	date := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	rows := []posting.Posting{
		{Date: date, Payee: "Opening balance", Account: "Assets:TestAccount", Commodity: "INR", Amount: decimal.NewFromInt(100)},
		{Date: date.AddDate(0, 1, 0), Payee: "Interest earned", Account: "Assets:TestAccount", Commodity: "INR", Amount: decimal.NewFromInt(2)},
		{Date: date.AddDate(0, 1, 0), Payee: "Interest earned", Account: "Income:Interest:TestAccount", Commodity: "INR", Amount: decimal.NewFromInt(-2)},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	detail := getRetirementDetail(db, config.RetirementGoal{Name: "Test", Savings: []string{"Assets:TestAccount"}, YearlyExpenses: 45000, SWR: 4})
	investmentTotal := detail["investmentTotal"].(decimal.Decimal)
	gainTotal := detail["gainTotal"].(decimal.Decimal)
	if !investmentTotal.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("investmentTotal = %s, want 100", investmentTotal)
	}
	if !gainTotal.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("gainTotal = %s, want 2", gainTotal)
	}
}
