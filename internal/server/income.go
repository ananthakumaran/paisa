package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

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
	var incomePostings []posting.Posting
	result := db.Where("account like ?", "Income:%").Order("date ASC").Find(&incomePostings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	var taxPostings []posting.Posting
	result = db.Where("account = ?", "Expenses:Tax").Order("date ASC").Find(&taxPostings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return gin.H{"income_timeline": computeIncomeTimeline(incomePostings), "tax_timeline": computeTaxTimeline(taxPostings)}
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
