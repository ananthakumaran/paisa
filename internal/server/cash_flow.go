package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CashFlow struct {
	Date        time.Time `json:"date"`
	Income      float64   `json:"income"`
	Expenses    float64   `json:"expenses"`
	Liabilities float64   `json:"liabilities"`
	Investment  float64   `json:"investment"`
	Tax         float64   `json:"tax"`
	Checking    float64   `json:"checking"`
	Balance     float64   `json:"balance"`
}

func GetCashFlow(db *gorm.DB) gin.H {
	return gin.H{"cash_flows": computeCashFlow(db)}
}

func computeCashFlow(db *gorm.DB) []CashFlow {
	var cashFlows []CashFlow

	expenses := posting.GroupByMonth(query.Init(db).Like("Expenses:%").NotLike("Expenses:Tax").All())
	incomes := posting.GroupByMonth(query.Init(db).Like("Income:%").All())
	liabilities := posting.GroupByMonth(query.Init(db).Like("Liabilities:%").All())
	investments := posting.GroupByMonth(query.Init(db).Like("Assets:%").NotLike("Assets:Checking").All())
	taxes := posting.GroupByMonth(query.Init(db).Like("Expenses:Tax").All())
	checkings := posting.GroupByMonth(query.Init(db).Like("Assets:Checking").All())
	postings := query.Init(db).All()

	end := utils.MaxTime(time.Now(), postings[len(postings)-1].Date)
	var balance float64 = 0
	for start := utils.BeginningOfMonth(postings[0].Date); start.Before(end); start = start.AddDate(0, 1, 0) {
		cashFlow := CashFlow{Date: start}

		key := start.Format("2006-01")
		ps, ok := expenses[key]
		if ok {
			cashFlow.Expenses = accounting.CostSum(ps)
		}

		ps, ok = incomes[key]
		if ok {
			cashFlow.Income = -accounting.CostSum(ps)
		}

		ps, ok = liabilities[key]
		if ok {
			cashFlow.Liabilities = -accounting.CostSum(ps)
		}

		ps, ok = investments[key]
		if ok {
			cashFlow.Investment = accounting.CostSum(ps)
		}

		ps, ok = taxes[key]
		if ok {
			cashFlow.Tax = accounting.CostSum(ps)
		}

		ps, ok = checkings[key]
		if ok {
			cashFlow.Checking = accounting.CostSum(ps)
		}

		balance += cashFlow.Checking
		cashFlow.Balance = balance

		cashFlows = append(cashFlows, cashFlow)
	}

	return cashFlows
}
