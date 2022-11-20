package server

import (
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type Aggregate struct {
	Date         time.Time `json:"date"`
	Account      string    `json:"account"`
	Amount       float64   `json:"amount"`
	MarketAmount float64   `json:"market_amount"`
}

type AllocationTargetConfig struct {
	Name     string
	Target   float64
	Accounts []string
}

type AllocationTarget struct {
	Name       string               `json:"name"`
	Target     float64              `json:"target"`
	Current    float64              `json:"current"`
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
	var configs []AllocationTargetConfig
	viper.UnmarshalKey("allocation_targets", &configs)

	totalMarketAmount := lo.Reduce(postings, func(acc float64, p posting.Posting, _ int) float64 { return acc + p.MarketAmount }, 0.0)

	for _, config := range configs {
		targetAllocations = append(targetAllocations, computeAllocationTarget(postings, config, totalMarketAmount))
	}

	return targetAllocations
}

func computeAllocationTarget(postings []posting.Posting, config AllocationTargetConfig, total float64) AllocationTarget {
	date := time.Now()
	postings = accounting.FilterByGlob(postings, config.Accounts)
	aggregates := computeAggregate(postings, date)
	currentTotal := lo.Reduce(postings, func(acc float64, p posting.Posting, _ int) float64 { return acc + p.MarketAmount }, 0.0)
	return AllocationTarget{Name: config.Name, Target: config.Target, Current: (currentTotal / total) * 100, Aggregates: aggregates}
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

		amount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 { return acc + p.Amount }, 0.0)
		marketAmount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 { return acc + p.MarketAmount }, 0.0)
		result[account] = Aggregate{Date: date, Account: account, Amount: amount, MarketAmount: marketAmount}

	}
	return result
}
