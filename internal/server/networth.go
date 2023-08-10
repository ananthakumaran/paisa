package server

import (
	"math"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Networth struct {
	Date                time.Time `json:"date"`
	InvestmentAmount    float64   `json:"investmentAmount"`
	WithdrawalAmount    float64   `json:"withdrawalAmount"`
	GainAmount          float64   `json:"gainAmount"`
	BalanceAmount       float64   `json:"balanceAmount"`
	NetInvestmentAmount float64   `json:"netInvestmentAmount"`
}

func GetNetworth(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%").OrLike("Liabilities:%").UntilToday().All()

	postings = service.PopulateMarketPrice(db, postings)
	networthTimeline := computeNetworthTimeline(db, postings)
	xirr := service.XIRR(db, postings)
	return gin.H{"networthTimeline": networthTimeline, "xirr": xirr}
}

func GetCurrentNetworth(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%").OrLike("Liabilities:%").UntilToday().All()

	postings = service.PopulateMarketPrice(db, postings)
	networth := computeNetworth(db, postings)
	xirr := service.XIRR(db, postings)
	return gin.H{"networth": networth, "xirr": xirr}
}

func computeNetworth(db *gorm.DB, postings []posting.Posting) Networth {
	var networth Networth

	if len(postings) == 0 {
		return networth
	}

	var investment float64 = 0
	var withdrawal float64 = 0
	var balance float64 = 0

	now := utils.BeginingOfDay(time.Now())
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
	networth = Networth{Date: now, InvestmentAmount: investment, WithdrawalAmount: withdrawal, GainAmount: gain, BalanceAmount: balance, NetInvestmentAmount: net_investment}

	return networth
}

func computeNetworthTimeline(db *gorm.DB, postings []posting.Posting) []Networth {
	var networths []Networth

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
		networths = append(networths, Networth{Date: start, InvestmentAmount: investment, WithdrawalAmount: withdrawal, GainAmount: gain, BalanceAmount: balance, NetInvestmentAmount: net_investment})

		if len(postings) == 0 && math.Abs(balance) < 0.01 {
			break
		}
	}
	return networths
}
