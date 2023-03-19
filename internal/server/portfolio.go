package server

import (
	"sort"
	"strings"

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
	ParentCommodityID string  `json:"parent_commodity_id"`
	CommodityName     string  `json:"commodity_name"`
	SecurityName      string  `json:"security_name"`
	Percentage        float64 `json:"percentage"`
	SecurityID        string  `json:"security_id"`
	SecurityType      string  `json:"security_type"`
	Amount            float64 `json:"amount"`
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
	cbs := lo.FlatMap(lo.Keys(byCommodity), func(commodity string, _ int) []CommodityBreakdown {
		ps := byCommodity[commodity]
		balance := accounting.CurrentBalance(ps)
		return computePortfolioAggregate(db, commodity, balance)
	})

	pas := rollupPortfolioAggregate(cbs, accounting.CurrentBalance(postings))
	sort.Slice(pas, func(i, j int) bool { return pas[i].Percentage > pas[j].Percentage })

	return gin.H{"portfolio_aggregates": pas}
}

func computePortfolioAggregate(db *gorm.DB, commodityName string, total float64) []CommodityBreakdown {
	commodity := commodity.FindByName(commodityName)
	portfolios := portfolio.GetPortfolios(db, commodity.Code)
	return lo.Map(portfolios, func(p portfolio.Portfolio, _ int) CommodityBreakdown {
		amount := (total * p.Percentage) / 100
		return CommodityBreakdown{
			SecurityName:      p.SecurityName,
			CommodityName:     commodity.Name,
			ParentCommodityID: p.ParentCommodityID,
			Amount:            amount,
			SecurityID:        p.SecurityID,
			SecurityType:      p.SecurityType}
	})
}

func mergeBreakdowns(cbs []CommodityBreakdown) []CommodityBreakdown {
	grouped := lo.GroupBy(cbs, func(c CommodityBreakdown) string {
		return c.CommodityName
	})

	return lo.Map(lo.Keys(grouped), func(key string, _ int) CommodityBreakdown {
		bs := grouped[key]
		return CommodityBreakdown{
			SecurityName:      bs[0].SecurityName,
			CommodityName:     bs[0].CommodityName,
			ParentCommodityID: bs[0].ParentCommodityID,
			Amount:            lo.SumBy(bs, func(b CommodityBreakdown) float64 { return b.Amount }),
			SecurityID:        strings.Join(lo.Map(bs, func(b CommodityBreakdown, _ int) string { return b.SecurityID }), ","),
			SecurityType:      bs[0].SecurityType}
	})
}

func rollupPortfolioAggregate(cbs []CommodityBreakdown, total float64) []PortfolioAggregate {
	grouped := lo.GroupBy(cbs, func(c CommodityBreakdown) string {
		t := c.SecurityType
		if t == "" {
			t = c.SecurityID
		}
		return strings.Join([]string{c.SecurityName, t}, ":")
	})
	return lo.Map(lo.Keys(grouped), func(key string, _ int) PortfolioAggregate {
		breakdowns := mergeBreakdowns(grouped[key])
		portfolioTotal := lo.SumBy(breakdowns, func(b CommodityBreakdown) float64 { return b.Amount })
		breakdowns = lo.Map(breakdowns, func(breakdown CommodityBreakdown, _ int) CommodityBreakdown {
			breakdown.Percentage = (breakdown.Amount / portfolioTotal) * 100
			return breakdown
		})
		totalPercentage := (portfolioTotal / total) * 100
		return PortfolioAggregate{Name: breakdowns[0].SecurityName, ID: key, Amount: portfolioTotal, Percentage: totalPercentage, Breakdowns: breakdowns}
	})
}
