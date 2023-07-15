package retirement

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func GetRetirementProgress(db *gorm.DB) gin.H {
	retirementConfig := config.GetConfig().Retirement

	savings := accounting.FilterByGlob(query.Init(db).Like("Assets:%").All(), retirementConfig.Savings)
	savings = service.PopulateMarketPrice(db, savings)
	savingsTotal := accounting.CurrentBalance(savings)

	yearlyExpenses := retirementConfig.YearlyExpenses
	if !(yearlyExpenses > 0) {
		yearlyExpenses = calculateAverageExpense(db, retirementConfig)
	}

	return gin.H{"savings_timeline": accounting.RunningBalance(db, savings), "savings_total": savingsTotal, "swr": retirementConfig.SWR, "yearly_expense": yearlyExpenses}
}

func calculateAverageExpense(db *gorm.DB, retirementConfig config.Retirement) float64 {
	now := time.Now()
	end := utils.BeginningOfMonth(now)
	start := end.AddDate(-2, 0, 0)
	expenses := accounting.FilterByGlob(query.Init(db).Like("Expenses:%").Where("date between ? AND ?", start, end).All(), retirementConfig.Expenses)
	return lo.SumBy(expenses, func(p posting.Posting) float64 { return p.Amount }) / 2
}
