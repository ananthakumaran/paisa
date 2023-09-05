package server

import (
	"sort"
	"time"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AccountBudget struct {
	Account   string            `json:"account"`
	Forecast  decimal.Decimal   `json:"forecast"`
	Actual    decimal.Decimal   `json:"actual"`
	Budgeted  decimal.Decimal   `json:"budgeted"`
	Rollover  decimal.Decimal   `json:"rollover"`
	Available decimal.Decimal   `json:"available"`
	Date      time.Time         `json:"date"`
	Expenses  []posting.Posting `json:"expenses"`
}

type Budget struct {
	Date               time.Time       `json:"date"`
	Accounts           []AccountBudget `json:"accounts"`
	AvailableThisMonth decimal.Decimal `json:"availableThisMonth"`
	EndOfMonthBalance  decimal.Decimal `json:"endOfMonthBalance"`
}

func GetBudget(db *gorm.DB) gin.H {
	forecastPostings := query.Init(db).Like("Expenses:%").Forecast().All()
	expenses := query.Init(db).Like("Expenses:%").All()
	return computeBudet(db, forecastPostings, expenses)
}

func GetCurrentBudget(db *gorm.DB) gin.H {
	forecastPostings := query.Init(db).Like("Expenses:%").Forecast().UntilThisMonthEnd().All()
	expenses := query.Init(db).Like("Expenses:%").UntilThisMonthEnd().All()
	return computeBudet(db, forecastPostings, expenses)
}

func computeBudet(db *gorm.DB, forecastPostings, expensesPostings []posting.Posting) gin.H {
	checkingBalance := accounting.CostSum(query.Init(db).AccountPrefix("Assets:Checking").All())
	availableForBudgeting := checkingBalance

	forecasts := utils.GroupByMonth(forecastPostings)
	expenses := utils.GroupByMonth(expensesPostings)

	accounts := lo.Uniq(lo.Map(forecastPostings, func(p posting.Posting, _ int) string {
		return p.Account
	}))
	sort.Strings(accounts)

	budgetsByMonth := make(map[string]Budget)
	balance := make(map[string]decimal.Decimal)

	if len(forecastPostings) > 0 {
		start := utils.BeginningOfMonth(forecastPostings[0].Date)
		end := utils.EndOfMonth(forecastPostings[len(forecastPostings)-1].Date)

		for start := start; start.Before(end) || start.Equal(end); start = start.AddDate(0, 1, 0) {
			month := start.Format("2006-01")
			var accountBudgets []AccountBudget

			forecastsByMonth := forecasts[month]
			date := lo.Must(time.Parse("2006-01", month))
			expensesByMonth, ok := expenses[month]
			if !ok {
				expensesByMonth = []posting.Posting{}
			}

			forecastsByAccount := utils.GroupByAccount(forecastsByMonth)
			expensesByAccount := utils.GroupByAccount(expensesByMonth)

			for _, account := range accounts {
				fs := forecastsByAccount[account]
				es := popExpenses(account, expensesByAccount)
				if !ok {
					es = []posting.Posting{}
				}

				budget := buildBudget(date, account, balance[account], fs, es)
				if budget.Available.IsPositive() {
					balance[account] = budget.Available
				} else {
					balance[account] = decimal.Zero
				}

				accountBudgets = append(accountBudgets, budget)
			}

			availableThisMonth := utils.SumBy(
				accountBudgets, func(budget AccountBudget) decimal.Decimal {
					if budget.Available.IsPositive() {
						return budget.Available
					}
					return decimal.Zero
				})

			endOfMonthBalance := checkingBalance.Sub(availableThisMonth)
			availableForBudgeting = endOfMonthBalance

			budgetsByMonth[month] = Budget{
				Date:               date,
				Accounts:           accountBudgets,
				EndOfMonthBalance:  endOfMonthBalance,
				AvailableThisMonth: availableThisMonth,
			}
		}
	}

	return gin.H{
		"budgetsByMonth":        budgetsByMonth,
		"checkingBalance":       checkingBalance,
		"availableForBudgeting": availableForBudgeting,
	}
}

func buildBudget(date time.Time, account string, balance decimal.Decimal, forecasts []posting.Posting, expenses []posting.Posting) AccountBudget {
	forecast := accounting.CostSum(forecasts)
	actual := accounting.CostSum(expenses)
	return AccountBudget{
		Account:   account,
		Forecast:  forecast,
		Actual:    actual,
		Rollover:  balance,
		Available: balance.Add(forecast.Sub(actual)),
		Date:      date,
		Expenses:  expenses,
	}
}

func popExpenses(forecastAccount string, expensesByAccount map[string][]posting.Posting) []posting.Posting {
	expenses := []posting.Posting{}
	for account, es := range expensesByAccount {
		if utils.IsSameOrParent(account, forecastAccount) {
			expenses = append(expenses, es...)
			delete(expensesByAccount, account)
		}
	}
	return expenses

}
