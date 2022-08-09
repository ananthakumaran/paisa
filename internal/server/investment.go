package server

import (
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type YearlyCard struct {
	StartDate time.Time         `json:"start_date"`
	EndDate   time.Time         `json:"end_date"`
	Postings  []posting.Posting `json:"postings"`
}

func GetInvestment(db *gorm.DB) gin.H {
	var postings []posting.Posting
	result := db.Where("account like ?", "Asset:%").Find(&postings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	postings = lo.Filter(postings, func(p posting.Posting, _ int) bool { return !service.IsInterest(db, p) })
	return gin.H{"postings": postings, "yearly_cards": computeYearlyCard(postings)}
}

func computeYearlyCard(postings []posting.Posting) []YearlyCard {
	var yearlyCards []YearlyCard = make([]YearlyCard, 0)

	if len(postings) == 0 {
		return yearlyCards
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

		yearlyCards = append(yearlyCards, YearlyCard{StartDate: start, EndDate: yearEnd, Postings: currentMonthPostings})

	}
	return yearlyCards
}
