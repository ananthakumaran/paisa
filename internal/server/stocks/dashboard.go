package stocks

import (
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Stock struct {
	Symbol           string  `json:"symbol"`
	AveragePrice     float64 `json:"averagePrice"`
	LastTradedPrice  float64 `json:"lastTradedPrice"`
	TargetPrice      float64 `json:"targetPrice"`
	Shares           int     `json:"shares"`
	TotalInvestment  float64 `json:"totalInvestment"`
	GainPercent      float64 `json:"gainPercent"`
	GainAmount       float64 `json:"gainAmount"`
	DrawdownFromPeak float64 `json:"drawdownFromPeak"`
	LastPurchaseDate string  `json:"lastPurchaseDate"`
}

type AssetBreakdown struct {
	Group            string          `json:"group"`
	InvestmentAmount decimal.Decimal `json:"investmentAmount"`
	WithdrawalAmount decimal.Decimal `json:"withdrawalAmount"`
	MarketAmount     decimal.Decimal `json:"marketAmount"`
	BalanceUnits     decimal.Decimal `json:"balanceUnits"`
	LatestPrice      decimal.Decimal `json:"latestPrice"`
	XIRR             decimal.Decimal `json:"xirr"`
	GainAmount       decimal.Decimal `json:"gainAmount"`
	AbsoluteReturn   decimal.Decimal `json:"absoluteReturn"`
	LastPurchaseDate time.Time       `json:"lastPurchaseDate"`
}

type UpdateTargetPriceRequest struct {
	Symbol      string  `json:"symbol"`
	TargetPrice float64 `json:"targetPrice"`
}

func GetDashboard(db *gorm.DB) gin.H {
	// stocks := []Stock{
	// 	{
	// 		Symbol:           "AAPL",
	// 		AveragePrice:     150.25,
	// 		LastTradedPrice:  175.50,
	// 		TargetPrice:      200.00,
	// 		Shares:           10,
	// 		TotalInvestment:  1502.50,
	// 		GainPercent:      16.80,
	// 		GainAmount:       252.50,
	// 		DrawdownFromPeak: -5.20,
	// 		LastPurchaseDate: time.Now().AddDate(0, -2, 0).Format("2006-01-02"),
	// 	},
	// }

	return GetBalance(db)
}

func GetBalance(db *gorm.DB) gin.H {
	return doGetBalance(db, "Assets:%", true)
}

func doGetBalance(db *gorm.DB, pattern string, rollup bool) gin.H {
	postings := query.Init(db).Like(pattern, "Income:CapitalGains:%").All()
	postings = service.PopulateMarketPrice(db, postings)
	breakdowns := ComputeBreakdowns(db, postings, rollup)

	stocks := make([]Stock, 0)
	for _, breakdown := range breakdowns {
		// Extract symbol from the group path (e.g., "Assets:Equity:Stocks:AAPL" -> "AAPL")
		parts := strings.Split(breakdown.Group, ":")
		symbol := parts[len(parts)-1]

		// Calculate average price per share
		averagePrice := decimal.Zero
		if !breakdown.BalanceUnits.IsZero() {
			netInvestment := breakdown.InvestmentAmount.Sub(breakdown.WithdrawalAmount)
			averagePrice = netInvestment.Div(breakdown.BalanceUnits)
		}

		stock := Stock{
			Symbol:           symbol,
			AveragePrice:     averagePrice.Round(2).InexactFloat64(),
			LastTradedPrice:  breakdown.LatestPrice.Round(2).InexactFloat64(),
			TargetPrice:      0,
			Shares:           int(breakdown.BalanceUnits.InexactFloat64()),
			TotalInvestment:  breakdown.InvestmentAmount.Sub(breakdown.WithdrawalAmount).Round(2).InexactFloat64(),
			GainPercent:      breakdown.GainAmount.Div(breakdown.InvestmentAmount).Mul(decimal.NewFromInt(100)).Round(2).InexactFloat64(),
			GainAmount:       breakdown.GainAmount.Round(2).InexactFloat64(),
			DrawdownFromPeak: 0,
			LastPurchaseDate: breakdown.LastPurchaseDate.Format("2006-01-02"),
		}
		stocks = append(stocks, stock)
	}

	return gin.H{"stocks": stocks}
}

func ComputeBreakdowns(db *gorm.DB, postings []posting.Posting, rollup bool) map[string]AssetBreakdown {
	accounts := make(map[string]bool)
	for _, p := range postings {
		if service.IsCapitalGains(p) {
			continue
		}

		if rollup {
			var parts []string
			for _, part := range strings.Split(p.Account, ":") {
				parts = append(parts, part)
				accounts[strings.Join(parts, ":")] = false
			}
		}
		accounts[p.Account] = true

	}

	result := make(map[string]AssetBreakdown)

	for group, leaf := range accounts {
		ps := lo.Filter(postings, func(p posting.Posting, _ int) bool {
			account := p.Account
			if service.IsCapitalGains(p) {
				account = service.CapitalGainsSourceAccount(p.Account)
			}
			return utils.IsSameOrParent(account, group)
		})
		breakdown := ComputeBreakdown(db, ps, leaf, group)
		if breakdown.BalanceUnits.GreaterThan(decimal.Zero) && strings.HasPrefix(breakdown.Group, "Assets:Equity:Stocks") {
			result[group] = breakdown
		}
	}

	return result
}

func ComputeBreakdown(db *gorm.DB, ps []posting.Posting, leaf bool, group string) AssetBreakdown {
	investmentAmount := lo.Reduce(ps, func(acc decimal.Decimal, p posting.Posting, _ int) decimal.Decimal {
		if utils.IsCheckingAccount(p.Account) || p.Amount.LessThan(decimal.Zero) || service.IsInterest(db, p) || service.IsStockSplit(db, p) || service.IsCapitalGains(p) {
			return acc
		} else {
			return acc.Add(p.Amount)
		}
	}, decimal.Zero)
	withdrawalAmount := lo.Reduce(ps, func(acc decimal.Decimal, p posting.Posting, _ int) decimal.Decimal {
		if !service.IsCapitalGains(p) && (utils.IsCheckingAccount(p.Account) || p.Amount.GreaterThan(decimal.Zero) || service.IsInterest(db, p) || service.IsStockSplit(db, p)) {
			return acc
		} else {
			return acc.Add(p.Amount.Neg())
		}
	}, decimal.Zero)
	psWithoutCapitalGains := lo.Filter(ps, func(p posting.Posting, _ int) bool {
		return !service.IsCapitalGains(p)
	})
	marketAmount := accounting.CurrentBalance(psWithoutCapitalGains)
	var balanceUnits decimal.Decimal
	if leaf {
		balanceUnits = lo.Reduce(ps, func(acc decimal.Decimal, p posting.Posting, _ int) decimal.Decimal {
			if !utils.IsCurrency(p.Commodity) {
				return acc.Add(p.Quantity)
			}
			return decimal.Zero
		}, decimal.Zero)
	}

	xirr := service.XIRR(db, ps)
	netInvestment := investmentAmount.Sub(withdrawalAmount)
	gainAmount := marketAmount.Sub(netInvestment)
	absoluteReturn := decimal.Zero
	if !investmentAmount.IsZero() {
		absoluteReturn = marketAmount.Sub(netInvestment).Div(investmentAmount)
	}
	lastPurchaseDate := time.Time{}
	for _, p := range ps {
		if p.Date.After(lastPurchaseDate) {
			lastPurchaseDate = p.Date
		}
	}
	return AssetBreakdown{
		InvestmentAmount: investmentAmount,
		WithdrawalAmount: withdrawalAmount,
		MarketAmount:     marketAmount,
		XIRR:             xirr,
		Group:            group,
		BalanceUnits:     balanceUnits,
		GainAmount:       gainAmount,
		AbsoluteReturn:   absoluteReturn,
		LastPurchaseDate: lastPurchaseDate,
	}
}

func UpdateTargetPrice(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateTargetPriceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		// TODO: Implement actual database update
		// For now, we'll just return the updated value
		c.JSON(200, gin.H{
			"symbol":      req.Symbol,
			"targetPrice": req.TargetPrice,
		})
	}
}
