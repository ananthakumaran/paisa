package accounting

import (
	"sort"
	"time"

	"path/filepath"

	"github.com/ananthakumaran/paisa/internal/model/posting"
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
						first.AddQuantity(quantity.Neg())
						available[0] = first
						quantity = decimal.Zero
					} else {
						quantity = quantity.Sub(first.Quantity)
						available = available[1:]
					}
				}
			}
		}
	}

	return available
}

func CostBalance(postings []posting.Posting) decimal.Decimal {
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	return utils.SumBy(lo.Values(byAccount), func(ps []posting.Posting) decimal.Decimal {
		return utils.SumBy(FIFO(ps), func(p posting.Posting) decimal.Decimal {
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
		if p.Date.Day() >= billDate {
			return utils.BeginningOfMonth(p.Date).AddDate(0, 1, 0).Format("2006-01")
		} else {
			return p.Date.Format("2006-01")
		}
	})
}
