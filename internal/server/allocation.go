package server

import (
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Aggregate struct {
	Date         time.Time       `json:"date"`
	Account      string          `json:"account"`
	Amount       decimal.Decimal `json:"amount"`
	MarketAmount decimal.Decimal `json:"market_amount"`
}

type AllocationTargetConfig struct {
	Name     string
	Target   decimal.Decimal
	Accounts []string
}

type AllocationTarget struct {
	Name       string               `json:"name"`
	Target     decimal.Decimal      `json:"target"`
	Current    decimal.Decimal      `json:"current"`
	Aggregates map[string]Aggregate `json:"aggregates"`
}

func GetAllocation(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%").All()

	now := time.Now()
	postings = lo.Map(postings, func(p posting.Posting, _ int) posting.Posting {
		p.MarketAmount = service.GetMarketPrice(db, p, now)
		return p
	})
	aggregates := computeAggregate(postings, now)
	aggregates_timeline := computeAggregateTimeline(postings)
	allocation_targets := computeAllocationTargets(postings)
	return gin.H{"aggregates": aggregates, "aggregates_timeline": aggregates_timeline, "allocation_targets": allocation_targets}
}

func computeAggregateTimeline(postings []posting.Posting) []map[string]Aggregate {
	var timeline []map[string]Aggregate

	var p posting.Posting
	var pastPostings []posting.Posting

	end := time.Now()
	for start := postings[0].Date; start.Before(end); start = start.AddDate(0, 0, 1) {
		for len(postings) > 0 && (postings[0].Date.Before(start) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			pastPostings = append(pastPostings, p)
		}

		timeline = append(timeline, computeAggregate(pastPostings, start))
	}
	return timeline
}

func computeAllocationTargets(postings []posting.Posting) []AllocationTarget {
	var targetAllocations []AllocationTarget
	allocationTargetConfigs := config.GetConfig().AllocationTargets

	totalMarketAmount := accounting.CurrentBalance(postings)

	for _, allocationTargetConfig := range allocationTargetConfigs {
		targetAllocations = append(targetAllocations, computeAllocationTarget(postings, allocationTargetConfig, totalMarketAmount))
	}

	return targetAllocations
}

func computeAllocationTarget(postings []posting.Posting, allocationTargetConfig config.AllocationTarget, total decimal.Decimal) AllocationTarget {
	date := time.Now()
	postings = accounting.FilterByGlob(postings, allocationTargetConfig.Accounts)
	aggregates := computeAggregate(postings, date)
	currentTotal := accounting.CurrentBalance(postings)
	return AllocationTarget{Name: allocationTargetConfig.Name, Target: decimal.NewFromFloat(allocationTargetConfig.Target), Current: (currentTotal.Div(total)).Mul(decimal.NewFromInt(100)), Aggregates: aggregates}
}

func computeAggregate(postings []posting.Posting, date time.Time) map[string]Aggregate {
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	result := make(map[string]Aggregate)
	for account, ps := range byAccount {
		var parts []string
		for _, part := range strings.Split(account, ":") {
			parts = append(parts, part)
			parent := strings.Join(parts, ":")
			result[parent] = Aggregate{Account: parent}
		}

		amount := accounting.CostSum(ps)
		marketAmount := accounting.CurrentBalance(ps)
		result[account] = Aggregate{Date: date, Account: account, Amount: amount, MarketAmount: marketAmount}

	}
	return result
}
