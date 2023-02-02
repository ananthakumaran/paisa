package accounting

import (
	"sort"
	"time"

	"path/filepath"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Balance struct {
	Date      time.Time
	Commodity string
	Quantity  float64
}

func Register(postings []posting.Posting) []Balance {
	balances := make([]Balance, 0)
	current := Balance{Quantity: 0}
	for _, p := range postings {
		sameDay := p.Date == current.Date
		current = Balance{Date: p.Date, Quantity: p.Quantity + current.Quantity, Commodity: p.Commodity}
		if sameDay {
			balances = balances[:len(balances)-1]
		}
		balances = append(balances, current)

	}
	return balances
}

func FilterByGlob(postings []posting.Posting, accounts []string) []posting.Posting {
	return lo.Filter(postings, func(p posting.Posting, _ int) bool {
		return lo.SomeBy(accounts, func(accountGlob string) bool {
			match, err := filepath.Match(accountGlob, p.Account)
			if err != nil {
				log.Fatal("Invalid account glob used for filtering", accountGlob, err)
			}
			return match
		})
	})
}

func FIFO(postings []posting.Posting) []posting.Posting {
	var available []posting.Posting
	for _, p := range postings {
		if p.Commodity == "INR" {
			if p.Amount > 0 {
				available = append(available, p)
			} else {
				amount := -p.Amount
				for amount > 0 && len(available) > 0 {
					first := available[0]
					if first.Amount > amount {
						first.AddAmount(-amount)
						available[0] = first
						amount = 0
					} else {
						amount -= first.Amount
						available = available[1:]
					}
				}
			}
		} else {
			if p.Quantity > 0 {
				available = append(available, p)
			} else {
				quantity := -p.Quantity
				for quantity > 0 && len(available) > 0 {
					first := available[0]
					if first.Quantity > quantity {
						first.AddQuantity(-quantity)
						available[0] = first
						quantity = 0
					} else {
						quantity -= first.Quantity
						available = available[1:]
					}
				}
			}
		}
	}

	return available
}

func CostBalance(postings []posting.Posting) float64 {
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	return lo.SumBy(lo.Values(byAccount), func(ps []posting.Posting) float64 {
		return lo.SumBy(FIFO(ps), func(p posting.Posting) float64 {
			return p.Amount
		})
	})

}

func CurrentBalance(postings []posting.Posting) float64 {
	return lo.SumBy(postings, func(p posting.Posting) float64 {
		return p.MarketAmount
	})
}

type Point struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

func RunningBalance(db *gorm.DB, postings []posting.Posting) []Point {
	sort.Slice(postings, func(i, j int) bool { return postings[i].Date.Before(postings[j].Date) })
	var series []Point

	if len(postings) == 0 {
		return series
	}

	var p posting.Posting
	var pastPostings []posting.Posting

	end := time.Now()
	for start := postings[0].Date; start.Before(end); start = start.AddDate(0, 0, 1) {
		for len(postings) > 0 && (postings[0].Date.Before(start) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			pastPostings = append(pastPostings, p)
		}

		balance := lo.SumBy(pastPostings, func(p posting.Posting) float64 {
			return service.GetMarketPrice(db, p, start)
		})
		series = append(series, Point{Date: start, Value: balance})
	}
	return series
}
