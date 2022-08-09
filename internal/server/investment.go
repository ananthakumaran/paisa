package server

import (
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type YearlyCard struct {
	StartDate         time.Time         `json:"start_date"`
	EndDate           time.Time         `json:"end_date"`
	Postings          []posting.Posting `json:"postings"`
	GrossSalaryIncome float64           `json:"gross_salary_income"`
	GrossOtherIncome  float64           `json:"gross_other_income"`
	NetTax            float64           `json:"net_tax"`
	NetIncome         float64           `json:"net_income"`
	NetInvestment     float64           `json:"net_investment"`
}

func GetInvestment(db *gorm.DB) gin.H {
	var postings []posting.Posting
	var incomes []posting.Posting
	var taxes []posting.Posting
	result := db.Where("account like ? order by date asc", "Asset:%").Find(&postings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	postings = lo.Filter(postings, func(p posting.Posting, _ int) bool { return !service.IsInterest(db, p) })

	result = db.Where("account like ? order by date asc", "Income:%").Find(&incomes)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	result = db.Where("account = ? order by date asc", "Tax").Find(&taxes)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	var p posting.Posting
	result = db.Order("date ASC").First(&p)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	return gin.H{"postings": postings, "yearly_cards": computeYearlyCard(p.Date, postings, taxes, incomes)}
}

func computeYearlyCard(start time.Time, assets []posting.Posting, taxes []posting.Posting, incomes []posting.Posting) []YearlyCard {
	var yearlyCards []YearlyCard = make([]YearlyCard, 0)

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
		for len(taxes) > 0 && utils.IsWithDate(taxes[0].Date, start, yearEnd) {
			p, taxes = taxes[0], taxes[1:]
			currentYearTaxes = append(currentYearTaxes, p)
		}

		netTax := lo.SumBy(currentYearTaxes, func(p posting.Posting) float64 { return p.Amount })

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

		yearlyCards = append(yearlyCards, YearlyCard{
			StartDate:         start,
			EndDate:           yearEnd,
			Postings:          currentYearPostings,
			NetTax:            netTax,
			GrossSalaryIncome: grossSalaryIncome,
			GrossOtherIncome:  grossOtherIncome,
			NetIncome:         grossSalaryIncome + grossOtherIncome - netTax,
			NetInvestment:     netInvestment,
		})

	}
	return yearlyCards
}
