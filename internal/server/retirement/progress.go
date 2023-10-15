package retirement

import (
	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func GetRetirementProgress(db *gorm.DB) gin.H {
	retirementConfig := config.GetConfig().Retirement

	savings := accounting.FilterByGlob(query.Init(db).Like("Assets:%").All(), retirementConfig.Savings)
	savings = service.PopulateMarketPrice(db, savings)
	savingsWithCapitalGains := accounting.FilterByGlob(query.Init(db).Like("Assets:%", "Income:CapitalGains:%").All(), retirementConfig.Savings)
	savingsWithCapitalGains = service.PopulateMarketPrice(db, savingsWithCapitalGains)
	savingsTotal := accounting.CurrentBalance(savings)

	yearlyExpenses := decimal.NewFromFloat(retirementConfig.YearlyExpenses)
	if !(yearlyExpenses.GreaterThan(decimal.Zero)) {
		yearlyExpenses = calculateAverageExpense(db, retirementConfig)
	}

	return gin.H{"savings_timeline": accounting.RunningBalance(db, savings), "savings_total": savingsTotal, "swr": retirementConfig.SWR, "yearly_expense": yearlyExpenses, "xirr": service.XIRR(db, savingsWithCapitalGains)}
}

func calculateAverageExpense(db *gorm.DB, retirementConfig config.Retirement) decimal.Decimal {
	now := utils.Now()
	end := utils.BeginningOfMonth(now)
	start := end.AddDate(-2, 0, 0)
	expenses := accounting.FilterByGlob(query.Init(db).Like("Expenses:%").Where("date between ? AND ?", start, end).All(), retirementConfig.Expenses)
	return utils.SumBy(expenses, func(p posting.Posting) decimal.Decimal { return p.Amount }).Div(decimal.NewFromInt(2))
}
