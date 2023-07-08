package server

import (
	"strings"

	"github.com/samber/lo"

	"sort"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/model/commodity"
	"github.com/ananthakumaran/paisa/internal/model/portfolio"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PortfolioDimension struct {
	Group      string
	GroupFn    func(CommodityBreakdown) string
	SubGroup   string
	SubGroupFn func(CommodityBreakdown) string
	FilterFn   func(CommodityBreakdown, int) bool
}

type CommodityBreakdown struct {
	ParentCommodityID string  `json:"parent_commodity_id"`
	CommodityName     string  `json:"commodity_name"`
	SecurityName      string  `json:"security_name"`
	SecurityRating    string  `json:"security_rating"`
	SecurityIndustry  string  `json:"security_industry"`
	Percentage        float64 `json:"percentage"`
	SecurityID        string  `json:"security_id"`
	SecurityType      string  `json:"security_type"`
	Amount            float64 `json:"amount"`
}

type PortfolioAggregate struct {
	Group      string               `json:"group"`
	SubGroup   string               `json:"sub_group"`
	ID         string               `json:"id"`
	Percentage float64              `json:"percentage"`
	Amount     float64              `json:"amount"`
	Breakdowns []CommodityBreakdown `json:"breakdowns"`
}

type PortfolioAllocationGroups struct {
	Commomdities        []string             `json:"commodities"`
	NameAndSecurityType []PortfolioAggregate `json:"name_and_security_type"`
	SecurityType        []PortfolioAggregate `json:"security_type"`
	Rating              []PortfolioAggregate `json:"rating"`
	Industry            []PortfolioAggregate `json:"industry"`
}

func GetPortfolioAllocation(db *gorm.DB) gin.H {
	groups := GetAccountPortfolioAllocation(db, "Assets")
	return gin.H{
		"commodities":            groups.Commomdities,
		"name_and_security_type": groups.NameAndSecurityType,
		"security_type":          groups.SecurityType,
		"rating":                 groups.Rating,
		"industry":               groups.Industry,
	}
}

func GetAccountPortfolioAllocation(db *gorm.DB, account string) PortfolioAllocationGroups {
	commoditieCodes := portfolio.GetAllParentCommodityIDs(db)
	commodities := lo.Map(commoditieCodes, func(code string, _ int) commodity.Commodity { return commodity.FindByCode(code) })
	postings := query.Init(db).AccountPrefix(account).Commodities(commodities).All()
	postings = service.PopulateMarketPrice(db, postings)
	byCommodity := lo.GroupBy(postings, func(p posting.Posting) string { return p.Commodity })

	supportedCommodities := lo.Keys(byCommodity)

	cbs := lo.FlatMap(lo.Keys(byCommodity), func(commodity string, _ int) []CommodityBreakdown {
		ps := byCommodity[commodity]
		balance := accounting.CurrentBalance(ps)
		if balance <= 0.0001 {
			supportedCommodities = lo.Without(supportedCommodities, commodity)
			return []CommodityBreakdown{}
		}
		return computePortfolioAggregate(db, commodity, balance)
	})

	return PortfolioAllocationGroups{
		Commomdities: supportedCommodities,
		NameAndSecurityType: rollupPortfolioAggregate(PortfolioDimension{
			FilterFn: func(c CommodityBreakdown, _ int) bool {
				return true
			},
			Group: "SecurityName",
			GroupFn: func(c CommodityBreakdown) string {
				return c.SecurityName
			},
			SubGroup: "SecurityType",
			SubGroupFn: func(c CommodityBreakdown) string {
				return orUnknown(c.SecurityType)
			}}, cbs),
		SecurityType: rollupPortfolioAggregate(PortfolioDimension{
			FilterFn: func(c CommodityBreakdown, _ int) bool {
				return true
			},
			Group: "SecurityType",
			GroupFn: func(c CommodityBreakdown) string {
				return orUnknown(c.SecurityType)
			},
			SubGroup: "SecurityType",
			SubGroupFn: func(c CommodityBreakdown) string {
				return orUnknown(c.SecurityType)
			}}, cbs),
		Rating: rollupPortfolioAggregate(PortfolioDimension{
			FilterFn: func(c CommodityBreakdown, _ int) bool {
				return c.SecurityType == "debt"
			},
			Group: "Rating",
			GroupFn: func(c CommodityBreakdown) string {
				return orUnknown(c.SecurityRating)
			},
			SubGroup: "Rating",
			SubGroupFn: func(c CommodityBreakdown) string {
				return orUnknown(c.SecurityRating)
			}}, cbs),
		Industry: rollupPortfolioAggregate(PortfolioDimension{
			FilterFn: func(c CommodityBreakdown, _ int) bool {
				return c.SecurityType == "equity"
			},
			Group: "Industry",
			GroupFn: func(c CommodityBreakdown) string {
				return orUnknown(c.SecurityIndustry)
			},
			SubGroup: "Industry",
			SubGroupFn: func(c CommodityBreakdown) string {
				return orUnknown(c.SecurityIndustry)
			}}, cbs),
	}
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
			SecurityRating:    p.SecurityRating,
			SecurityIndustry:  p.SecurityIndustry,
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
			SecurityRating:    bs[0].SecurityRating,
			SecurityIndustry:  bs[0].SecurityIndustry,
			SecurityType:      bs[0].SecurityType}
	})
}

func rollupPortfolioAggregate(dimension PortfolioDimension, cbs []CommodityBreakdown) []PortfolioAggregate {
	cbs = lo.Filter(cbs, dimension.FilterFn)
	total := lo.SumBy(cbs, func(b CommodityBreakdown) float64 { return b.Amount })
	grouped := lo.GroupBy(cbs, func(c CommodityBreakdown) string {
		return strings.Join([]string{dimension.GroupFn(c), dimension.SubGroupFn(c)}, ":")
	})
	pas := lo.Map(lo.Keys(grouped), func(key string, _ int) PortfolioAggregate {
		breakdowns := mergeBreakdowns(grouped[key])
		portfolioTotal := lo.SumBy(breakdowns, func(b CommodityBreakdown) float64 { return b.Amount })
		breakdowns = lo.Map(breakdowns, func(breakdown CommodityBreakdown, _ int) CommodityBreakdown {
			breakdown.Percentage = (breakdown.Amount / portfolioTotal) * 100
			return breakdown
		})
		totalPercentage := (portfolioTotal / total) * 100
		return PortfolioAggregate{Group: dimension.GroupFn(breakdowns[0]), SubGroup: dimension.SubGroupFn(breakdowns[0]), ID: key, Amount: portfolioTotal, Percentage: totalPercentage, Breakdowns: breakdowns}
	})

	sort.Slice(pas, func(i, j int) bool { return pas[i].Percentage > pas[j].Percentage })
	return pas

}

func orUnknown(str string) string {
	if str == "" {
		return "unknown"
	}
	return str
}
