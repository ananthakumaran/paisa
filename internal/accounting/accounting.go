package accounting

import (
	"sort"
	"time"

	"path/filepath"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Balance struct {
	Date      time.Time
	Commodity string
	Quantity  decimal.Decimal
}

func Register(postings []posting.Posting) []Balance {
	balances := make([]Balance, 0)
	current := Balance{Quantity: decimal.Zero}
	for _, p := range postings {
		sameDay := p.Date == current.Date
		current = Balance{Date: p.Date, Quantity: p.Quantity.Add(current.Quantity), Commodity: p.Commodity}
		if sameDay {
			balances = balances[:len(balances)-1]
		}
		balances = append(balances, current)

	}
	return balances
}

func FilterByGlob(postings []posting.Posting, accounts []string) []posting.Posting {
	negatePresent := lo.SomeBy(accounts, func(accountGlob string) bool {
		return accountGlob[0] == '!'
	})
	var combine func(collection []string, predicate func(item string) bool) bool
	if negatePresent {
		combine = lo.EveryBy[string]
	} else {
		combine = lo.SomeBy[string]
	}

	return lo.Filter(postings, func(p posting.Posting, _ int) bool {
		return combine(accounts, func(accountGlob string) bool {
			negative := false

			if accountGlob[0] == '!' {
				negative = true
				accountGlob = accountGlob[1:]
			}

			account := p.Account
			if service.IsCapitalGains(p) {
				account = service.CapitalGainsSourceAccount(p.Account)
			}
			match, err := filepath.Match(accountGlob, account)
			if err != nil {
				log.Fatal("Invalid account glob used for filtering", accountGlob, err)
			}

			if negative {
				return !match
			}
			return match
		})
	})
}

func FIFO(postings []posting.Posting) []posting.Posting {
	var available []posting.Posting
	for _, p := range postings {
		available, _ = fifoApply(available, p)
	}

	return available
}

// fifoApply adds the posting to the available lots, or if it reduces
// the holding, consumes the available lots in FIFO order. Returns the
// remaining lots and, for commodities, the consumed portion of each lot.
func fifoApply(available []posting.Posting, p posting.Posting) ([]posting.Posting, []posting.Posting) {
	var consumed []posting.Posting
	if utils.IsCurrency(p.Commodity) {
		if p.Amount.GreaterThan(decimal.Zero) {
			available = append(available, p)
		} else {
			amount := p.Amount.Neg()
			for amount.GreaterThan(decimal.Zero) && len(available) > 0 {
				first := available[0]
				if first.Amount.GreaterThan(amount) {
					first.AddAmount(amount.Neg())
					available[0] = first
					amount = decimal.Zero
				} else {
					amount = amount.Sub(first.Amount)
					available = available[1:]
				}
			}
		}
	} else {
		if p.Quantity.GreaterThan(decimal.Zero) {
			available = append(available, p)
		} else {
			quantity := p.Quantity.Neg()
			for quantity.GreaterThan(decimal.Zero) && len(available) > 0 {
				first := available[0]
				if first.Quantity.GreaterThan(quantity) {
					consumed = append(consumed, first.WithQuantity(quantity))
					first.AddQuantity(quantity.Neg())
					available[0] = first
					quantity = decimal.Zero
				} else {
					consumed = append(consumed, first)
					quantity = quantity.Sub(first.Quantity)
					available = available[1:]
				}
			}
		}
	}

	return available, consumed
}

type TaxSale struct {
	Sell      posting.Posting
	Purchases []posting.Posting
}

// TaxFIFO matches the sells of a single account against its purchases
// in FIFO order, for computing the cost basis and capital gains.
//
// Unlike FIFO, a stock split is not treated as a sell followed by a
// fresh purchase. Instead the units of the open lots are scaled by the
// split ratio, keeping their purchase date and cost. Bonus units issued
// against an Income account are acquired at zero cost on the allotment
// date, irrespective of the value booked against the Income account.
func TaxFIFO(db *gorm.DB, postings []posting.Posting) ([]posting.Posting, []TaxSale) {
	var available []posting.Posting
	var sales []TaxSale
	splits := make(map[string]bool)
	for i, p := range postings {
		if service.IsStockSplit(db, p) {
			if splits[p.TransactionID] {
				continue
			}
			splits[p.TransactionID] = true

			split := lo.Filter(postings[i:], func(sp posting.Posting, _ int) bool {
				return sp.TransactionID == p.TransactionID
			})
			if adjusted, ok := applySplit(available, split); ok {
				available = adjusted
			} else {
				for _, sp := range split {
					available, _ = fifoApply(available, sp)
				}
			}
			continue
		}

		if service.IsBonus(db, p) {
			p.Amount = decimal.Zero
		}

		var consumed []posting.Posting
		available, consumed = fifoApply(available, p)
		if !utils.IsCurrency(p.Commodity) && p.Quantity.IsNegative() {
			p.Amount = service.SellAmount(db, p)
			sales = append(sales, TaxSale{Sell: p, Purchases: consumed})
		}
	}

	return available, sales
}

// applySplit scales the units of the available lots so that they add
// up to the holding after the split, leaving the cost of each lot
// unchanged.
func applySplit(available []posting.Posting, split []posting.Posting) ([]posting.Posting, bool) {
	held := utils.SumBy(available, func(p posting.Posting) decimal.Decimal { return p.Quantity })
	after := held.Add(utils.SumBy(split, func(p posting.Posting) decimal.Decimal { return p.Quantity }))
	if !held.IsPositive() || !after.IsPositive() {
		return available, false
	}

	adjusted := make([]posting.Posting, len(available))
	remaining := after
	for i, p := range available {
		if i == len(available)-1 {
			// assign the remainder to the last lot to avoid rounding errors
			p.Quantity = remaining
		} else {
			p.Quantity = p.Quantity.Mul(after).Div(held)
			remaining = remaining.Sub(p.Quantity)
		}
		adjusted[i] = p
	}
	return adjusted, true
}

func CostBalance(postings []posting.Posting) decimal.Decimal {
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	return utils.SumBy(lo.Values(byAccount), func(ps []posting.Posting) decimal.Decimal {
		return utils.SumBy(FIFO(ps), func(p posting.Posting) decimal.Decimal {
			return p.Amount
		})
	})

}

// TaxCostBalance is like CostBalance, but handles stock splits and
// bonus issues the way TaxFIFO does.
func TaxCostBalance(db *gorm.DB, postings []posting.Posting) decimal.Decimal {
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	return utils.SumBy(lo.Values(byAccount), func(ps []posting.Posting) decimal.Decimal {
		available, _ := TaxFIFO(db, ps)
		return utils.SumBy(available, func(p posting.Posting) decimal.Decimal {
			return p.Amount
		})
	})
}

func CurrentBalance(postings []posting.Posting) decimal.Decimal {
	return utils.SumBy(postings, func(p posting.Posting) decimal.Decimal {
		return p.MarketAmount
	})
}

func CurrentBalanceOn(db *gorm.DB, postings []posting.Posting, date time.Time) decimal.Decimal {
	return utils.SumBy(postings, func(p posting.Posting) decimal.Decimal {
		return service.GetMarketPrice(db, p, date)
	})
}

func CostSum(postings []posting.Posting) decimal.Decimal {
	return utils.SumBy(postings, func(p posting.Posting) decimal.Decimal {
		return p.Amount
	})
}

type Point struct {
	Date  time.Time       `json:"date"`
	Value decimal.Decimal `json:"value"`
}

func RunningBalance(db *gorm.DB, postings []posting.Posting) []Point {
	SortAsc(postings)
	var series []Point

	if len(postings) == 0 {
		return series
	}

	var p posting.Posting
	accumulator := make(map[string]decimal.Decimal)

	end := utils.EndOfToday()
	for start := postings[0].Date; start.Before(end); start = start.AddDate(0, 0, 1) {
		for len(postings) > 0 && (postings[0].Date.Before(start) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			accumulator[p.Commodity] = accumulator[p.Commodity].Add(p.Quantity)
		}

		balance := decimal.Zero

		for commodity, quantity := range accumulator {
			if utils.IsCurrency(commodity) {
				balance = balance.Add(quantity)
			} else {
				price := service.GetUnitPrice(db, commodity, start)
				if !price.Value.Equal(decimal.Zero) {
					balance = balance.Add(quantity.Mul(price.Value))
				} else {
					balance = balance.Add(quantity)
				}
			}
		}
		series = append(series, Point{Date: start, Value: balance})
	}
	return series
}

func SortTransactionAsc(transactions []transaction.Transaction) []transaction.Transaction {
	sort.Slice(transactions, func(i, j int) bool { return transactions[i].Date.Before(transactions[j].Date) })
	return transactions
}

func SortAsc(postings []posting.Posting) []posting.Posting {
	sort.Slice(postings, func(i, j int) bool { return postings[i].Date.Before(postings[j].Date) })
	return postings
}

func SortDesc(postings []posting.Posting) []posting.Posting {
	sort.Slice(postings, func(i, j int) bool { return postings[i].Date.After(postings[j].Date) })
	return postings
}

func PopulateBalance(postings []posting.Posting) []posting.Posting {
	SortAsc(postings)
	accumulator := make(map[string]decimal.Decimal)

	for i, p := range postings {
		accumulator[p.Account] = accumulator[p.Account].Add(p.Quantity)
		postings[i].Balance = accumulator[p.Account]
	}
	return postings
}

func GroupByAccount(posts []posting.Posting) map[string][]posting.Posting {
	return lo.GroupBy(posts, func(post posting.Posting) string {
		return post.Account
	})
}

func GroupByMonthlyBillingCycle(postsings []posting.Posting, billDate int) map[string][]posting.Posting {
	return lo.GroupBy(postsings, func(p posting.Posting) string {
		if p.Date.Day() > billDate {
			return utils.BeginningOfMonth(p.Date).AddDate(0, 1, 0).Format("2006-01")
		} else {
			return p.Date.Format("2006-01")
		}
	})
}
