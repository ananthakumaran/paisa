package liabilities

import (
	"math"
	"sort"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Overview struct {
	Date           time.Time `json:"date"`
	DrawnAmount    float64   `json:"drawn_amount"`
	RepaidAmount   float64   `json:"repaid_amount"`
	InterestAmount float64   `json:"interest_amount"`
}

type Interest struct {
	Account          string     `json:"account"`
	OverviewTimeline []Overview `json:"overview_timeline"`
	APR              float64    `json:"apr"`
}

func GetInterest(db *gorm.DB) gin.H {
	postings := query.Init(db).Unbudgeted().Like("Liabilities:%").All()
	expenses := query.Init(db).Unbudgeted().Like("Expenses:Interest:%").All()
	postings = service.PopulateMarketPrice(db, postings)
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.RestName(1) })
	var interests []Interest
	for account, ps := range byAccount {
		es := lo.Filter(expenses, func(e posting.Posting, _ int) bool { return e.RestName(1) == "Interest:"+account })
		ps = append(ps, es...)
		interests = append(interests, Interest{Account: "Liabilities:" + account, APR: service.APR(db, ps), OverviewTimeline: computeOverviewTimeline(db, ps)})
	}

	return gin.H{"interest_timeline_breakdown": interests}
}

func computeOverviewTimeline(db *gorm.DB, postings []posting.Posting) []Overview {
	sort.Slice(postings, func(i, j int) bool { return postings[i].Date.Before(postings[j].Date) })
	netliabilities := []Overview{}

	var p posting.Posting
	var pastPostings []posting.Posting

	if len(postings) == 0 {
		return netliabilities
	}

	end := utils.MaxTime(time.Now(), postings[len(postings)-1].Date)
	for start := postings[0].Date; start.Before(end) || start.Equal(end); start = start.AddDate(0, 0, 1) {
		for len(postings) > 0 && (postings[0].Date.Before(start) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			pastPostings = append(pastPostings, p)
		}

		drawn := lo.Reduce(pastPostings, func(agg float64, p posting.Posting, _ int) float64 {
			if p.Amount > 0 || service.IsInterest(db, p) {
				return agg
			} else {
				return -p.Amount + agg
			}
		}, 0)

		repaid := lo.Reduce(pastPostings, func(agg float64, p posting.Posting, _ int) float64 {
			if p.Amount < 0 {
				return agg
			} else {
				return p.Amount + agg
			}
		}, 0)

		balance := lo.Reduce(pastPostings, func(agg float64, p posting.Posting, _ int) float64 {
			if service.IsInterest(db, p) {
				return agg
			} else {
				return -service.GetMarketPrice(db, p, start) + agg
			}
		}, 0)

		interest := balance + repaid - drawn
		netliabilities = append(netliabilities, Overview{Date: start, DrawnAmount: drawn, RepaidAmount: repaid, InterestAmount: interest})

		if len(postings) == 0 && math.Abs(balance) < 0.01 {
			break
		}
	}
	return netliabilities
}
