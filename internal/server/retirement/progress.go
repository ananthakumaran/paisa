package retirement

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type RetirementConfig struct {
	Swr            float64
	YearlyExpenses float64 `mapstructure:"yearly_expenses"`
	Expenses       []string
	Savings        []string
}

func GetRetirementProgress(db *gorm.DB) gin.H {
	var config RetirementConfig = RetirementConfig{Swr: 4, Savings: []string{"Assets:*"}, Expenses: []string{"Expenses:*"}, YearlyExpenses: 0}
	viper.UnmarshalKey("retirement", &config)

	savings := accounting.FilterByGlob(query.Init(db).Like("Assets:%").All(), config.Savings)
	savings = service.PopulateMarketPrice(db, savings)
	savingsTotal := accounting.CurrentBalance(savings)

	yearlyExpenses := config.YearlyExpenses
	if !(yearlyExpenses > 0) {
		yearlyExpenses = calculateAverageExpense(db, config)
	}

	return gin.H{"savings_timeline": accounting.RunningBalance(db, savings), "savings_total": savingsTotal, "swr": config.Swr, "yearly_expense": yearlyExpenses}
}

func calculateAverageExpense(db *gorm.DB, config RetirementConfig) float64 {
	now := time.Now()
	end := utils.BeginningOfMonth(now)
	start := end.AddDate(-2, 0, 0)
	expenses := accounting.FilterByGlob(query.Init(db).Like("Expenses:%").Where("date between ? AND ?", start, end).All(), config.Expenses)
	return lo.SumBy(expenses, func(p posting.Posting) float64 { return p.Amount }) / 2
}
