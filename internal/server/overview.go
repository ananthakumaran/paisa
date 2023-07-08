package server

import (
	"math"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Overview struct {
	Date                time.Time `json:"date"`
	InvestmentAmount    float64   `json:"investment_amount"`
	WithdrawalAmount    float64   `json:"withdrawal_amount"`
	GainAmount          float64   `json:"gain_amount"`
	BalanceAmount       float64   `json:"balance_amount"`
	NetInvestmentAmount float64   `json:"net_investment_amount"`
}

func GetOverview(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%").OrLike("Liabilities:%").All()

	postings = service.PopulateMarketPrice(db, postings)
	overviewTimeline := computeOverviewTimeline(db, postings)
	xirr := service.XIRR(db, postings)
	return gin.H{"overview_timeline": overviewTimeline, "xirr": xirr}
}

func computeOverview(db *gorm.DB, postings []posting.Posting) Overview {
	var networth Overview

	if len(postings) == 0 {
		return networth
	}

	var investment float64 = 0
	var withdrawal float64 = 0
	var balance float64 = 0

	now := time.Now()
	for _, p := range postings {
		isInterest := service.IsInterest(db, p)

		if p.Amount > 0 && !isInterest {
			investment += p.Amount
		}

		if p.Amount < 0 && !isInterest {
			withdrawal += -p.Amount
		}

		if isInterest {
			balance += p.Amount
		} else {
			balance += service.GetMarketPrice(db, p, now)
		}
	}

	gain := balance + withdrawal - investment
	net_investment := investment - withdrawal
	networth = Overview{Date: now, InvestmentAmount: investment, WithdrawalAmount: withdrawal, GainAmount: gain, BalanceAmount: balance, NetInvestmentAmount: net_investment}

	return networth
}

func computeOverviewTimeline(db *gorm.DB, postings []posting.Posting) []Overview {
	var networths []Overview

	var p posting.Posting
	var pastPostings []posting.Posting

	if len(postings) == 0 {
		return networths
	}

	end := time.Now()
	for start := postings[0].Date; start.Before(end); start = start.AddDate(0, 0, 1) {
		for len(postings) > 0 && (postings[0].Date.Before(start) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			pastPostings = append(pastPostings, p)
		}

		var investment float64 = 0
		var withdrawal float64 = 0
		var balance float64 = 0

		for _, p := range pastPostings {
			isInterest := service.IsInterest(db, p)

			if p.Amount > 0 && !isInterest {
				investment += p.Amount
			}

			if p.Amount < 0 && !isInterest {
				withdrawal += -p.Amount
			}

			if isInterest {
				balance += p.Amount
			} else {
				balance += service.GetMarketPrice(db, p, start)
			}
		}

		gain := balance + withdrawal - investment
		net_investment := investment - withdrawal
		networths = append(networths, Overview{Date: start, InvestmentAmount: investment, WithdrawalAmount: withdrawal, GainAmount: gain, BalanceAmount: balance, NetInvestmentAmount: net_investment})

		if len(postings) == 0 && math.Abs(balance) < 0.01 {
			break
		}
	}
	return networths
}
