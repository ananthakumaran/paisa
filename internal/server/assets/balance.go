package assets

import (
	"strings"

	"github.com/samber/lo"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AssetBreakdown struct {
	Group            string  `json:"group"`
	InvestmentAmount float64 `json:"investment_amount"`
	WithdrawalAmount float64 `json:"withdrawal_amount"`
	MarketAmount     float64 `json:"market_amount"`
	BalanceUnits     float64 `json:"balance_units"`
	LatestPrice      float64 `json:"latest_price"`
	XIRR             float64 `json:"xirr"`
}

func GetBalance(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%").All()
	postings = service.PopulateMarketPrice(db, postings)
	breakdowns := computeBreakdown(db, postings)
	return gin.H{"asset_breakdowns": breakdowns}
}

func computeBreakdown(db *gorm.DB, postings []posting.Posting) map[string]AssetBreakdown {
	accounts := make(map[string]bool)
	for _, p := range postings {
		var parts []string
		for _, part := range strings.Split(p.Account, ":") {
			parts = append(parts, part)
			accounts[strings.Join(parts, ":")] = false
		}
		accounts[p.Account] = true

	}

	result := make(map[string]AssetBreakdown)

	for group, leaf := range accounts {
		ps := lo.Filter(postings, func(p posting.Posting, _ int) bool { return utils.IsSameOrParent(p.Account, group) })
		investmentAmount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 {
			if p.Account == "Assets:Checking" || p.Amount < 0 || service.IsInterest(db, p) {
				return acc
			} else {
				return acc + p.Amount
			}
		}, 0.0)
		withdrawalAmount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 {
			if p.Account == "Assets:Checking" || p.Amount > 0 || service.IsInterest(db, p) {
				return acc
			} else {
				return acc + -p.Amount
			}
		}, 0.0)
		marketAmount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 { return acc + p.MarketAmount }, 0.0)
		var balanceUnits float64
		if leaf {
			balanceUnits = lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 {
				if !utils.IsCurrency(p.Commodity) {
					return acc + p.Quantity
				}
				return 0.0
			}, 0.0)
		}

		xirr := service.XIRR(db, ps)
		breakdown := AssetBreakdown{InvestmentAmount: investmentAmount, WithdrawalAmount: withdrawalAmount, MarketAmount: marketAmount, XIRR: xirr, Group: group, BalanceUnits: balanceUnits}
		result[group] = breakdown
	}

	return result
}
