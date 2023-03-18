package server

import (
	"sort"

	"github.com/samber/lo"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/model/commodity"
	"github.com/ananthakumaran/paisa/internal/model/portfolio"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommodityBreakdown struct {
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
	Amount     float64 `json:"amount"`
}

type PortfolioAggregate struct {
	Name       string               `json:"name"`
	ID         string               `json:"id"`
	Percentage float64              `json:"percentage"`
	Amount     float64              `json:"amount"`
	Breakdowns []CommodityBreakdown `json:"breakdowns"`
}

func GetPortfolioAllocation(db *gorm.DB) gin.H {
	commodities := commodity.FindByType(price.MutualFund)
	postings := query.Init(db).Like("Assets:%").Commodities(commodities).All()
	postings = service.PopulateMarketPrice(db, postings)
	byCommodity := lo.GroupBy(postings, func(p posting.Posting) string { return p.Commodity })
	pas := lo.FlatMap(lo.Keys(byCommodity), func(commodity string, _ int) []PortfolioAggregate {
		ps := byCommodity[commodity]
		balance := accounting.CurrentBalance(ps)
		return computePortfolioAggregate(db, commodity, balance)
	})

	pas = rollupPortfolioAggregate(pas, accounting.CurrentBalance(postings))
	sort.Slice(pas, func(i, j int) bool { return pas[i].Percentage > pas[j].Percentage })

	return gin.H{"portfolio_aggregates": pas}
}

func computePortfolioAggregate(db *gorm.DB, commodityName string, total float64) []PortfolioAggregate {
	commodity := commodity.FindByName(commodityName)
	portfolios := portfolio.GetPortfolios(db, commodity.Code)
	return lo.Map(portfolios, func(p portfolio.Portfolio, _ int) PortfolioAggregate {
		amount := (total * p.Percentage) / 100
		commodityBreakdown := CommodityBreakdown{Name: commodity.Name, Amount: amount}
		return PortfolioAggregate{Name: p.CommodityName, ID: p.CommodityID, Amount: amount, Breakdowns: []CommodityBreakdown{commodityBreakdown}}
	})
}

func rollupPortfolioAggregate(pas []PortfolioAggregate, total float64) []PortfolioAggregate {
	byPortfolioID := lo.GroupBy(pas, func(p PortfolioAggregate) string { return p.ID })
	return lo.Map(lo.Keys(byPortfolioID), func(id string, _ int) PortfolioAggregate {
		pas := byPortfolioID[id]
		breakdowns := lo.FlatMap(pas, func(p PortfolioAggregate, _ int) []CommodityBreakdown { return p.Breakdowns })
		portfolioTotal := lo.SumBy(breakdowns, func(b CommodityBreakdown) float64 { return b.Amount })
		breakdowns = lo.Map(breakdowns, func(breakdown CommodityBreakdown, _ int) CommodityBreakdown {
			breakdown.Percentage = (breakdown.Amount / portfolioTotal) * 100
			return breakdown
		})
		totalPercentage := (portfolioTotal / total) * 100
		return PortfolioAggregate{Name: pas[0].Name, ID: pas[0].ID, Amount: portfolioTotal, Percentage: totalPercentage, Breakdowns: breakdowns}
	})
}
