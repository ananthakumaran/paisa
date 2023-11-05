package goal

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

func getRetirementSummary(db *gorm.DB, conf config.RetirementGoal) GoalSummary {
	savings := accounting.FilterByGlob(query.Init(db).Like("Assets:%").All(), conf.Savings)
	savings = service.PopulateMarketPrice(db, savings)
	savingsTotal := accounting.CurrentBalance(savings)

	yearlyExpenses := decimal.NewFromFloat(conf.YearlyExpenses)
	if !(yearlyExpenses.GreaterThan(decimal.Zero)) {
		yearlyExpenses = calculateAverageExpense(db, conf)
	}

	target := yearlyExpenses.Div(decimal.NewFromFloat(conf.SWR)).Mul(decimal.NewFromFloat(100))

	return GoalSummary{
		Type:    "retirement",
		Name:    conf.Name,
		Current: savingsTotal,
		Target:  target,
		Icon:    conf.Icon,
	}
}

func calculateAverageExpense(db *gorm.DB, conf config.RetirementGoal) decimal.Decimal {
	now := utils.Now()
	end := utils.BeginningOfMonth(now)
	start := end.AddDate(-2, 0, 0)
	expenses := accounting.FilterByGlob(query.Init(db).Like("Expenses:%").Where("date between ? AND ?", start, end).All(), conf.Expenses)
	return utils.SumBy(expenses, func(p posting.Posting) decimal.Decimal { return p.Amount }).Div(decimal.NewFromInt(2))
}

func getRetirementDetail(db *gorm.DB, conf config.RetirementGoal) gin.H {
	savings := accounting.FilterByGlob(query.Init(db).Like("Assets:%").All(), conf.Savings)
	savings = service.PopulateMarketPrice(db, savings)
	savingsWithCapitalGains := accounting.FilterByGlob(query.Init(db).Like("Assets:%", "Income:CapitalGains:%").All(), conf.Savings)
	savingsWithCapitalGains = service.PopulateMarketPrice(db, savingsWithCapitalGains)
	savingsTotal := accounting.CurrentBalance(savings)

	yearlyExpenses := decimal.NewFromFloat(conf.YearlyExpenses)
	if !(yearlyExpenses.GreaterThan(decimal.Zero)) {
		yearlyExpenses = calculateAverageExpense(db, conf)
	}

	return gin.H{
		"type":            "retirement",
		"name":            conf.Name,
		"icon":            conf.Icon,
		"savingsTimeline": accounting.RunningBalance(db, savings),
		"savingsTotal":    savingsTotal,
		"swr":             conf.SWR,
		"yearlyExpense":   yearlyExpenses,
		"xirr":            service.XIRR(db, savingsWithCapitalGains),
	}
}
