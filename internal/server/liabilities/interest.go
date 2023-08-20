package liabilities

import (
	"sort"
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

type Overview struct {
	Date           time.Time       `json:"date"`
	DrawnAmount    decimal.Decimal `json:"drawn_amount"`
	RepaidAmount   decimal.Decimal `json:"repaid_amount"`
	InterestAmount decimal.Decimal `json:"interest_amount"`
}

type Interest struct {
	Account          string          `json:"account"`
	OverviewTimeline []Overview      `json:"overview_timeline"`
	APR              decimal.Decimal `json:"apr"`
}

func GetInterest(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Liabilities:%").All()
	expenses := query.Init(db).Like("Expenses:Interest:%").All()
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

		drawn := lo.Reduce(pastPostings, func(agg decimal.Decimal, p posting.Posting, _ int) decimal.Decimal {
			if p.Amount.GreaterThan(decimal.Zero) || service.IsInterest(db, p) {
				return agg
			} else {
				return p.Amount.Neg().Add(agg)
			}
		}, decimal.Zero)

		repaid := lo.Reduce(pastPostings, func(agg decimal.Decimal, p posting.Posting, _ int) decimal.Decimal {
			if p.Amount.LessThan(decimal.Zero) {
				return agg
			} else {
				return p.Amount.Add(agg)
			}
		}, decimal.Zero)

		balance := lo.Reduce(pastPostings, func(agg decimal.Decimal, p posting.Posting, _ int) decimal.Decimal {
			if service.IsInterest(db, p) {
				return agg
			} else {
				return service.GetMarketPrice(db, p, start).Neg().Add(agg)
			}
		}, decimal.Zero)

		interest := balance.Add(repaid).Sub(drawn)
		netliabilities = append(netliabilities, Overview{Date: start, DrawnAmount: drawn, RepaidAmount: repaid, InterestAmount: interest})

		if len(postings) == 0 && balance.Abs().LessThan(decimal.NewFromFloat(0.01)) {
			break
		}
	}
	return netliabilities
}
