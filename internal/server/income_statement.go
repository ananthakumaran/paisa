package server

import (
	"sort"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type IncomeStatement struct {
	StartingBalance decimal.Decimal            `json:"startingBalance"`
	EndingBalance   decimal.Decimal            `json:"endingBalance"`
	Date            time.Time                  `json:"date"`
	Income          map[string]decimal.Decimal `json:"income"`
	Interest        map[string]decimal.Decimal `json:"interest"`
	Equity          map[string]decimal.Decimal `json:"equity"`
	Pnl             map[string]decimal.Decimal `json:"pnl"`
	Liabilities     map[string]decimal.Decimal `json:"liabilities"`
	Tax             map[string]decimal.Decimal `json:"tax"`
	Expenses        map[string]decimal.Decimal `json:"expenses"`
}

type RunningBalance struct {
	amount   decimal.Decimal
	quantity map[string]decimal.Decimal
}

func GetIncomeStatement(db *gorm.DB) gin.H {
	postings := query.Init(db).All()
	statements := computeStatement(db, postings)
	monthlyStatements := computeMonthlyStatement(db, postings)
	return gin.H{"yearly": statements, "monthly": monthlyStatements}
}

func computeStatement(db *gorm.DB, postings []posting.Posting) map[string]IncomeStatement {
	statements := make(map[string]IncomeStatement)

	grouped := utils.GroupByFY(postings)
	fys := lo.Keys(grouped)
	sort.Strings(fys)

	runnings := make(map[string]RunningBalance)
	startingBalance := decimal.Zero

	for _, fy := range fys {
		incomeStatement := IncomeStatement{}
		start, end := utils.ParseFY(fy)
		incomeStatement.Date = start
		incomeStatement.StartingBalance = startingBalance
		incomeStatement.Income = make(map[string]decimal.Decimal)
		incomeStatement.Interest = make(map[string]decimal.Decimal)
		incomeStatement.Equity = make(map[string]decimal.Decimal)
		incomeStatement.Pnl = make(map[string]decimal.Decimal)
		incomeStatement.Liabilities = make(map[string]decimal.Decimal)
		incomeStatement.Tax = make(map[string]decimal.Decimal)
		incomeStatement.Expenses = make(map[string]decimal.Decimal)

		for _, p := range grouped[fy] {

			category := utils.FirstName(p.Account)

			switch category {
			case "Income":
				if service.IsCapitalGains(p) {
					sourceAccount := service.CapitalGainsSourceAccount(p.Account)
					r := runnings[sourceAccount]
					if r.quantity == nil {
						r.quantity = make(map[string]decimal.Decimal)
					}
					r.amount = r.amount.Add(p.Amount)
					runnings[sourceAccount] = r
				} else if strings.HasPrefix(p.Account, "Income:Interest") {
					incomeStatement.Interest[p.Account] = incomeStatement.Interest[p.Account].Add(p.Amount)
				} else {
					incomeStatement.Income[p.Account] = incomeStatement.Income[p.Account].Add(p.Amount)
				}
			case "Equity":
				incomeStatement.Equity[p.Account] = incomeStatement.Equity[p.Account].Add(p.Amount)
			case "Expenses":
				if strings.HasPrefix(p.Account, "Expenses:Tax") {
					incomeStatement.Tax[p.Account] = incomeStatement.Tax[p.Account].Add(p.Amount)
				} else {
					incomeStatement.Expenses[p.Account] = incomeStatement.Expenses[p.Account].Add(p.Amount)
				}
			case "Liabilities":
				incomeStatement.Liabilities[p.Account] = incomeStatement.Liabilities[p.Account].Add(p.Amount)
			case "Assets":
				r := runnings[p.Account]
				if r.quantity == nil {
					r.quantity = make(map[string]decimal.Decimal)
				}
				r.amount = r.amount.Add(p.Amount)
				r.quantity[p.Commodity] = r.quantity[p.Commodity].Add(p.Quantity)
				runnings[p.Account] = r
			default:
				// ignore
			}
		}

		for account, r := range runnings {
			diff := r.amount.Neg()
			for commodity, quantity := range r.quantity {
				diff = diff.Add(service.GetPrice(db, commodity, quantity, end))
			}
			incomeStatement.Pnl[account] = diff

			r.amount = r.amount.Add(diff)
			runnings[account] = r
		}

		startingBalance = startingBalance.
			Add(sumBalance(incomeStatement.Income).Neg()).
			Add(sumBalance(incomeStatement.Interest).Neg()).
			Add(sumBalance(incomeStatement.Equity).Neg()).
			Add(sumBalance(incomeStatement.Tax).Neg()).
			Add(sumBalance(incomeStatement.Expenses).Neg()).
			Add(sumBalance(incomeStatement.Pnl)).
			Add(sumBalance(incomeStatement.Liabilities).Neg())

		incomeStatement.EndingBalance = startingBalance

		statements[fy] = incomeStatement
	}

	return statements
}

func computeMonthlyStatement(db *gorm.DB, postings []posting.Posting) map[string]map[string]IncomeStatement {
	monthlyStatements := make(map[string]map[string]IncomeStatement)

	grouped := utils.GroupByFY(postings)
	fys := lo.Keys(grouped)
	sort.Strings(fys)

	for _, fy := range fys {
		fyPostings := grouped[fy]
		monthlyStatements[fy] = make(map[string]IncomeStatement)

		// Group postings by month within the financial year
		monthlyGrouped := utils.GroupByMonth(fyPostings)

		// Get all months in the financial year
		start, end := utils.ParseFY(fy)
		current := start

		runnings := make(map[string]RunningBalance)
		startingBalance := decimal.Zero

		for current.Before(end) || current.Equal(end) {
			monthKey := current.Format("2006-01")
			incomeStatement := IncomeStatement{}
			incomeStatement.Date = current
			incomeStatement.StartingBalance = startingBalance
			incomeStatement.Income = make(map[string]decimal.Decimal)
			incomeStatement.Interest = make(map[string]decimal.Decimal)
			incomeStatement.Equity = make(map[string]decimal.Decimal)
			incomeStatement.Pnl = make(map[string]decimal.Decimal)
			incomeStatement.Liabilities = make(map[string]decimal.Decimal)
			incomeStatement.Tax = make(map[string]decimal.Decimal)
			incomeStatement.Expenses = make(map[string]decimal.Decimal)

			monthPostings := monthlyGrouped[monthKey]

			for _, p := range monthPostings {
				category := utils.FirstName(p.Account)

				switch category {
				case "Income":
					if service.IsCapitalGains(p) {
						sourceAccount := service.CapitalGainsSourceAccount(p.Account)
						r := runnings[sourceAccount]
						if r.quantity == nil {
							r.quantity = make(map[string]decimal.Decimal)
						}
						r.amount = r.amount.Add(p.Amount)
						runnings[sourceAccount] = r
					} else if strings.HasPrefix(p.Account, "Income:Interest") {
						incomeStatement.Interest[p.Account] = incomeStatement.Interest[p.Account].Add(p.Amount)
					} else {
						incomeStatement.Income[p.Account] = incomeStatement.Income[p.Account].Add(p.Amount)
					}
				case "Equity":
					incomeStatement.Equity[p.Account] = incomeStatement.Equity[p.Account].Add(p.Amount)
				case "Expenses":
					if strings.HasPrefix(p.Account, "Expenses:Tax") {
						incomeStatement.Tax[p.Account] = incomeStatement.Tax[p.Account].Add(p.Amount)
					} else {
						incomeStatement.Expenses[p.Account] = incomeStatement.Expenses[p.Account].Add(p.Amount)
					}
				case "Liabilities":
					incomeStatement.Liabilities[p.Account] = incomeStatement.Liabilities[p.Account].Add(p.Amount)
				case "Assets":
					r := runnings[p.Account]
					if r.quantity == nil {
						r.quantity = make(map[string]decimal.Decimal)
					}
					r.amount = r.amount.Add(p.Amount)
					r.quantity[p.Commodity] = r.quantity[p.Commodity].Add(p.Quantity)
					runnings[p.Account] = r
				default:
					// ignore
				}
			}

			// Calculate PnL for this month
			monthEndDate := current.AddDate(0, 1, -1) // Last day of current month
			if monthEndDate.After(end) {
				monthEndDate = end
			}

			for account, r := range runnings {
				diff := r.amount.Neg()
				for commodity, quantity := range r.quantity {
					diff = diff.Add(service.GetPrice(db, commodity, quantity, monthEndDate))
				}
				if !diff.IsZero() {
					incomeStatement.Pnl[account] = diff
				}

				r.amount = r.amount.Add(diff)
				runnings[account] = r
			}

			startingBalance = startingBalance.
				Add(sumBalance(incomeStatement.Income).Neg()).
				Add(sumBalance(incomeStatement.Interest).Neg()).
				Add(sumBalance(incomeStatement.Equity).Neg()).
				Add(sumBalance(incomeStatement.Tax).Neg()).
				Add(sumBalance(incomeStatement.Expenses).Neg()).
				Add(sumBalance(incomeStatement.Pnl)).
				Add(sumBalance(incomeStatement.Liabilities).Neg())

			incomeStatement.EndingBalance = startingBalance

			monthlyStatements[fy][monthKey] = incomeStatement

			// Move to next month
			current = current.AddDate(0, 1, 0)
		}
	}

	return monthlyStatements
}

func sumBalance(breakdown map[string]decimal.Decimal) decimal.Decimal {
	total := decimal.Zero
	for k, v := range breakdown {
		total = total.Add(v)

		if v.Equal(decimal.Zero) {
			delete(breakdown, k)
		}
	}
	return total
}
