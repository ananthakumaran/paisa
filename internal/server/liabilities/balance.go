package liabilities

import (
	"sort"
	"strings"

	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AssetBreakdown struct {
	Group          string          `json:"group"`
	DrawnAmount    decimal.Decimal `json:"drawn_amount"`
	RepaidAmount   decimal.Decimal `json:"repaid_amount"`
	InterestAmount decimal.Decimal `json:"interest_amount"`
	BalanceAmount  decimal.Decimal `json:"balance_amount"`
	APR            decimal.Decimal `json:"apr"`
}

func GetBalance(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Liabilities:%").All()
	expenses := query.Init(db).Like("Expenses:Interest:%").All()
	postings = service.PopulateMarketPrice(db, postings)
	breakdowns := computeBreakdown(db, postings, expenses)
	return gin.H{"liability_breakdowns": breakdowns}
}

func computeBreakdown(db *gorm.DB, postings, expenses []posting.Posting) map[string]AssetBreakdown {
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

	for group := range accounts {
		ps := lo.Filter(postings, func(p posting.Posting, _ int) bool { return utils.IsSameOrParent(p.Account, group) })
		es := lo.Filter(expenses, func(e posting.Posting, _ int) bool { return utils.IsSameOrParent("Liabilities:"+e.RestName(2), group) })
		sort.Slice(ps, func(i, j int) bool { return ps[i].Date.Before(ps[j].Date) })
		ps = append(ps, es...)

		drawn := lo.Reduce(ps, func(agg decimal.Decimal, p posting.Posting, _ int) decimal.Decimal {
			if p.Amount.GreaterThan(decimal.Zero) || service.IsInterest(db, p) {
				return agg
			} else {
				return p.Amount.Neg().Add(agg)
			}
		}, decimal.Zero)

		repaid := lo.Reduce(ps, func(agg decimal.Decimal, p posting.Posting, _ int) decimal.Decimal {
			if p.Amount.LessThan(decimal.Zero) {
				return agg
			} else {
				return p.Amount.Add(agg)
			}
		}, decimal.Zero)

		balance := lo.Reduce(ps, func(agg decimal.Decimal, p posting.Posting, _ int) decimal.Decimal {
			if service.IsInterest(db, p) {
				return agg
			} else {
				return p.MarketAmount.Neg().Add(agg)
			}
		}, decimal.Zero)

		interest := balance.Add(repaid).Sub(drawn)

		apr := service.APR(db, ps)
		breakdown := AssetBreakdown{DrawnAmount: drawn, RepaidAmount: repaid, BalanceAmount: balance, APR: apr, Group: group, InterestAmount: interest}
		result[group] = breakdown
	}

	return result
}
