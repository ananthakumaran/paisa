package server

import (
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Gain struct {
	Account  string            `json:"account"`
	Overview Overview          `json:"overview"`
	XIRR     float64           `json:"xirr"`
	Postings []posting.Posting `json:"postings"`
}

type AccountGain struct {
	Account          string            `json:"account"`
	OverviewTimeline []Overview        `json:"overview_timeline"`
	XIRR             float64           `json:"xirr"`
	Postings         []posting.Posting `json:"postings"`
}

func GetGain(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%").NotLike("Assets:Checking").All()
	postings = service.PopulateMarketPrice(db, postings)
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	var gains []Gain
	for account, ps := range byAccount {
		gains = append(gains, Gain{Account: account, XIRR: service.XIRR(db, ps), Overview: computeOverview(db, ps), Postings: ps})
	}

	return gin.H{"gain_breakdown": gains}
}

func GetAccountGain(db *gorm.DB, account string) gin.H {
	postings := query.Init(db).AccountPrefix(account).All()
	postings = service.PopulateMarketPrice(db, postings)
	gain := AccountGain{Account: account, XIRR: service.XIRR(db, postings), OverviewTimeline: computeOverviewTimeline(db, postings), Postings: postings}

	commodities := lo.Uniq(lo.Map(postings, func(p posting.Posting, _ int) string { return p.Commodity }))
	var portfolio_groups PortfolioAllocationGroups
	portfolio_groups = GetAccountPortfolioAllocation(db, account)
	if !(len(commodities) > 0 && len(portfolio_groups.Commomdities) == len(commodities)) {
		portfolio_groups = PortfolioAllocationGroups{Commomdities: []string{}, NameAndSecurityType: []PortfolioAggregate{}, SecurityType: []PortfolioAggregate{}, Rating: []PortfolioAggregate{}, Industry: []PortfolioAggregate{}}
	}

	return gin.H{"gain_timeline_breakdown": gain, "portfolio_allocation": portfolio_groups}
}
