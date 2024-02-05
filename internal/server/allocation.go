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
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Aggregate struct {
	Date         time.Time       `json:"date"`
	Account      string          `json:"account"`
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

	now := utils.EndOfToday()
	postings = lo.Map(postings, func(p posting.Posting, _ int) posting.Posting {
		p.MarketAmount = service.GetMarketPrice(db, p, now)
		return p
	})
	aggregates := computeAggregate(db, postings, now)
	aggregates_timeline := computeAggregateTimeline(db, postings)
	allocation_targets := computeAllocationTargets(db, postings)
	return gin.H{"aggregates": aggregates, "aggregates_timeline": aggregates_timeline, "allocation_targets": allocation_targets}
}

func computeAggregateTimeline(db *gorm.DB, postings []posting.Posting) []map[string]Aggregate {
	var timeline []map[string]Aggregate

	var p posting.Posting

	type RunningSum struct {
		units decimal.Decimal
		cost  decimal.Decimal
	}

	accumulator := make(map[string]map[string]RunningSum)

	if len(postings) == 0 {
		return timeline
	}

	end := utils.EndOfToday()
	for start := postings[0].Date; start.Before(end); start = start.AddDate(0, 0, 1) {
		for len(postings) > 0 && (postings[0].Date.Before(start) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			rsByAccount := accumulator[p.Account]
			if rsByAccount == nil {
				rsByAccount = make(map[string]RunningSum)
			}

			rs := rsByAccount[p.Commodity]
			rs.units = rs.units.Add(p.Quantity)
			rs.cost = rs.cost.Add(p.Amount)

			rsByAccount[p.Commodity] = rs
			accumulator[p.Account] = rsByAccount

		}

		result := make(map[string]Aggregate)

		for account, rsByAccount := range accumulator {
			marketAmount := decimal.Zero

			for commodity, rs := range rsByAccount {
				if utils.IsCurrency(commodity) {
					marketAmount = marketAmount.Add(rs.cost)
				} else {
					price := service.GetUnitPrice(db, commodity, start)
					if !price.Value.Equal(decimal.Zero) {
						marketAmount = marketAmount.Add(rs.units.Mul(price.Value))
					} else {
						marketAmount = marketAmount.Add(rs.cost)
					}
				}
			}

			result[account] = Aggregate{Date: start, Account: account, MarketAmount: marketAmount}

		}

		timeline = append(timeline, result)

	}
	return timeline
}

func computeAllocationTargets(db *gorm.DB, postings []posting.Posting) []AllocationTarget {
	var targetAllocations []AllocationTarget
	allocationTargetConfigs := config.GetConfig().AllocationTargets

	if len(postings) == 0 {
		return targetAllocations
	}

	totalMarketAmount := accounting.CurrentBalance(postings)

	for _, allocationTargetConfig := range allocationTargetConfigs {
		targetAllocations = append(targetAllocations, computeAllocationTarget(db, postings, allocationTargetConfig, totalMarketAmount))
	}

	return targetAllocations
}

func computeAllocationTarget(db *gorm.DB, postings []posting.Posting, allocationTargetConfig config.AllocationTarget, total decimal.Decimal) AllocationTarget {
	date := utils.EndOfToday()
	postings = accounting.FilterByGlob(postings, allocationTargetConfig.Accounts)
	aggregates := computeAggregate(db, postings, date)
	currentTotal := accounting.CurrentBalance(postings)
	return AllocationTarget{Name: allocationTargetConfig.Name, Target: decimal.NewFromFloat(allocationTargetConfig.Target), Current: (currentTotal.Div(total)).Mul(decimal.NewFromInt(100)), Aggregates: aggregates}
}

func computeAggregate(db *gorm.DB, postings []posting.Posting, date time.Time) map[string]Aggregate {
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	result := make(map[string]Aggregate)
	for account, ps := range byAccount {
		var parts []string
		for _, part := range strings.Split(account, ":") {
			parts = append(parts, part)
			parent := strings.Join(parts, ":")
			result[parent] = Aggregate{Account: parent}
		}

		marketAmount := accounting.CurrentBalanceOn(db, ps, date)
		result[account] = Aggregate{Date: date, Account: account, MarketAmount: marketAmount}

	}
	return result
}
