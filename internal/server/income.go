package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type IncomeYearlyCard struct {
	StartDate   time.Time         `json:"start_date"`
	EndDate     time.Time         `json:"end_date"`
	Postings    []posting.Posting `json:"postings"`
	GrossIncome decimal.Decimal   `json:"gross_income"`
	NetTax      decimal.Decimal   `json:"net_tax"`
	NetIncome   decimal.Decimal   `json:"net_income"`
}

type Income struct {
	Date     time.Time         `json:"date"`
	Postings []posting.Posting `json:"postings"`
}

type Tax struct {
	StartDate time.Time         `json:"start_date"`
	EndDate   time.Time         `json:"end_date"`
	Postings  []posting.Posting `json:"postings"`
}

func GetIncome(db *gorm.DB) gin.H {
	incomePostings := query.Init(db).Like("Income:%").All()
	taxPostings := query.Init(db).Like("Expenses:Tax").All()
	p := query.Init(db).First()

	return gin.H{"income_timeline": computeIncomeTimeline(incomePostings), "tax_timeline": computeTaxTimeline(taxPostings), "yearly_cards": computeIncomeYearlyCard(p.Date, taxPostings, incomePostings)}
}

func computeIncomeTimeline(postings []posting.Posting) []Income {
	var incomes []Income = make([]Income, 0)

	if len(postings) == 0 {
		return incomes
	}

	var p posting.Posting
	end := time.Now()
	for start := utils.BeginningOfMonth(postings[0].Date); start.Before(end); start = start.AddDate(0, 1, 0) {
		var currentMonthPostings []posting.Posting = make([]posting.Posting, 0)
		for len(postings) > 0 && (postings[0].Date.Before(utils.EndOfMonth(start)) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			currentMonthPostings = append(currentMonthPostings, p)
		}

		incomes = append(incomes, Income{Date: start, Postings: currentMonthPostings})

	}
	return incomes
}

func computeTaxTimeline(postings []posting.Posting) []Tax {
	var taxes []Tax = make([]Tax, 0)

	if len(postings) == 0 {
		return taxes
	}

	var p posting.Posting
	end := time.Now()
	for start := utils.BeginningOfFinancialYear(postings[0].Date); start.Before(end); start = start.AddDate(1, 0, 0) {
		yearEnd := utils.EndOfFinancialYear(start)
		var currentMonthPostings []posting.Posting = make([]posting.Posting, 0)
		for len(postings) > 0 && (postings[0].Date.Before(yearEnd) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			currentMonthPostings = append(currentMonthPostings, p)
		}

		taxes = append(taxes, Tax{StartDate: start, EndDate: yearEnd, Postings: currentMonthPostings})

	}
	return taxes
}

func computeIncomeYearlyCard(start time.Time, taxes []posting.Posting, incomes []posting.Posting) []IncomeYearlyCard {
	var yearlyCards []IncomeYearlyCard = make([]IncomeYearlyCard, 0)

	var p posting.Posting
	end := time.Now()
	for start = utils.BeginningOfFinancialYear(start); start.Before(end); start = start.AddDate(1, 0, 0) {
		yearEnd := utils.EndOfFinancialYear(start)
		var netTax decimal.Decimal = decimal.Zero
		for len(taxes) > 0 && utils.IsWithDate(taxes[0].Date, start, yearEnd) {
			p, taxes = taxes[0], taxes[1:]
			netTax = netTax.Add(p.Amount)
		}

		var currentYearIncomes []posting.Posting = make([]posting.Posting, 0)
		for len(incomes) > 0 && utils.IsWithDate(incomes[0].Date, start, yearEnd) {
			p, incomes = incomes[0], incomes[1:]
			currentYearIncomes = append(currentYearIncomes, p)
		}

		grossIncome := utils.SumBy(currentYearIncomes, func(p posting.Posting) decimal.Decimal {
			return p.Amount.Neg()
		})

		yearlyCards = append(yearlyCards, IncomeYearlyCard{
			StartDate:   start,
			EndDate:     yearEnd,
			Postings:    currentYearIncomes,
			NetTax:      netTax,
			GrossIncome: grossIncome,
			NetIncome:   grossIncome.Sub(netTax),
		})

	}
	return yearlyCards
}
