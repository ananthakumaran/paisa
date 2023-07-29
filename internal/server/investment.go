package server

import (
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type InvestmentYearlyCard struct {
	StartDate         time.Time         `json:"start_date"`
	EndDate           time.Time         `json:"end_date"`
	Postings          []posting.Posting `json:"postings"`
	GrossSalaryIncome float64           `json:"gross_salary_income"`
	GrossOtherIncome  float64           `json:"gross_other_income"`
	NetTax            float64           `json:"net_tax"`
	NetIncome         float64           `json:"net_income"`
	NetInvestment     float64           `json:"net_investment"`
	NetExpense        float64           `json:"net_expense"`
	SavingsRate       float64           `json:"savings_rate"`
}

func GetInvestment(db *gorm.DB) gin.H {
	assets := query.Init(db).Like("Assets:%").NotLike("Assets:Checking").
		Where("transaction_id not in (select transaction_id from postings p where p.account like ? and p.transaction_id = transaction_id)", "Liabilities:%").
		All()
	incomes := query.Init(db).Like("Income:%").All()
	expenses := query.Init(db).Like("Expenses:%").All()
	p := query.Init(db).First()

	return gin.H{"assets": assets, "yearly_cards": computeInvestmentYearlyCard(p.Date, assets, expenses, incomes)}
}

func computeInvestmentYearlyCard(start time.Time, assets []posting.Posting, expenses []posting.Posting, incomes []posting.Posting) []InvestmentYearlyCard {
	var yearlyCards []InvestmentYearlyCard = make([]InvestmentYearlyCard, 0)

	if len(assets) == 0 {
		return yearlyCards
	}

	var p posting.Posting
	end := time.Now()
	for start = utils.BeginningOfFinancialYear(start); start.Before(end); start = start.AddDate(1, 0, 0) {
		yearEnd := utils.EndOfFinancialYear(start)
		var currentYearPostings []posting.Posting = make([]posting.Posting, 0)
		for len(assets) > 0 && utils.IsWithDate(assets[0].Date, start, yearEnd) {
			p, assets = assets[0], assets[1:]
			currentYearPostings = append(currentYearPostings, p)
		}

		var currentYearTaxes []posting.Posting = make([]posting.Posting, 0)
		var currentYearExpenses []posting.Posting = make([]posting.Posting, 0)

		for len(expenses) > 0 && utils.IsWithDate(expenses[0].Date, start, yearEnd) {
			p, expenses = expenses[0], expenses[1:]
			if p.Account == "Expenses:Tax" {
				currentYearTaxes = append(currentYearTaxes, p)
			} else {
				currentYearExpenses = append(currentYearExpenses, p)
			}
		}

		netTax := lo.SumBy(currentYearTaxes, func(p posting.Posting) float64 { return p.Amount })
		netExpense := lo.SumBy(currentYearExpenses, func(p posting.Posting) float64 { return p.Amount })

		var currentYearIncomes []posting.Posting = make([]posting.Posting, 0)
		for len(incomes) > 0 && utils.IsWithDate(incomes[0].Date, start, yearEnd) {
			p, incomes = incomes[0], incomes[1:]
			currentYearIncomes = append(currentYearIncomes, p)
		}

		grossSalaryIncome := lo.SumBy(currentYearIncomes, func(p posting.Posting) float64 {
			if strings.HasPrefix(p.Account, "Income:Salary") {
				return -p.Amount
			} else {
				return 0
			}
		})
		grossOtherIncome := lo.SumBy(currentYearIncomes, func(p posting.Posting) float64 {
			if !strings.HasPrefix(p.Account, "Income:Salary") {
				return -p.Amount
			} else {
				return 0
			}
		})

		netInvestment := lo.SumBy(currentYearPostings, func(p posting.Posting) float64 { return p.Amount })

		netIncome := grossSalaryIncome + grossOtherIncome - netTax
		var savingsRate float64 = 0
		if netIncome != 0 {
			savingsRate = (netInvestment / netIncome) * 100
		}

		yearlyCards = append(yearlyCards, InvestmentYearlyCard{
			StartDate:         start,
			EndDate:           yearEnd,
			Postings:          currentYearPostings,
			NetTax:            netTax,
			GrossSalaryIncome: grossSalaryIncome,
			GrossOtherIncome:  grossOtherIncome,
			NetIncome:         netIncome,
			NetInvestment:     netInvestment,
			NetExpense:        netExpense,
			SavingsRate:       savingsRate,
		})

	}
	return yearlyCards
}
