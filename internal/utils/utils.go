package utils

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/google/btree"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"golang.org/x/exp/constraints"
)

func BTreeDescendFirstLessOrEqual[I btree.Item](tree *btree.BTree, item I) I {
	var hit I
	tree.DescendLessOrEqual(item, func(item btree.Item) bool {
		hit = item.(I)
		return false
	})

	return hit
}

func BTreeToSlice[I btree.Item](tree *btree.BTree) []I {
	var items []I = make([]I, 0)
	tree.Descend(func(item btree.Item) bool {
		items = append(items, item.(I))
		return true
	})

	return items
}

func FY(date time.Time) string {
	if config.GetConfig().FinancialYearStartingMonth == time.January {
		return fmt.Sprintf("%d", date.Year())
	}

	if date.Month() < config.GetConfig().FinancialYearStartingMonth {
		return fmt.Sprintf("%d-%d", date.Year()-1, date.Year()%100)
	} else {
		return fmt.Sprintf("%d-%d", date.Year(), (date.Year()+1)%100)
	}
}

func FYHuman(date time.Time) string {
	if config.GetConfig().FinancialYearStartingMonth == time.January {
		return fmt.Sprintf("%d", date.Year())
	}

	if date.Month() < config.GetConfig().FinancialYearStartingMonth {
		return fmt.Sprintf("%d - %d", date.Year()-1, date.Year()%100)
	} else {
		return fmt.Sprintf("%d - %d", date.Year(), (date.Year()+1)%100)
	}
}

func BeginningOfFinancialYear(date time.Time) time.Time {
	beginningOfMonth := BeginningOfMonth(date)
	if beginningOfMonth.Month() < config.GetConfig().FinancialYearStartingMonth {
		return beginningOfMonth.AddDate(-1, int(config.GetConfig().FinancialYearStartingMonth-beginningOfMonth.Month()), 0)
	} else {
		return beginningOfMonth.AddDate(0, -int(beginningOfMonth.Month()-config.GetConfig().FinancialYearStartingMonth), 0)
	}
}

func EndOfFinancialYear(date time.Time) time.Time {
	return EndOfMonth(BeginningOfFinancialYear(date).AddDate(0, 11, 0))
}

func BeginningOfMonth(date time.Time) time.Time {
	return toDate(date.AddDate(0, 0, -date.Day()+1))
}

func EndOfMonth(date time.Time) time.Time {
	return toDate(date.AddDate(0, 1, -date.Day()))
}

func IsWithDate(date time.Time, start time.Time, end time.Time) bool {
	return (date.Equal(start) || date.After(start)) && (date.Before(end) || date.Equal(end))
}

func BeginingOfDay(date time.Time) time.Time {
	return toDate(date)
}

func toDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
}

func IsSameOrParent(account string, comparison string) bool {
	if account == comparison {
		return true
	}

	return strings.HasPrefix(account, comparison+":")
}

func IsCurrency(currency string) bool {
	return currency == config.DefaultCurrency()
}

func IsCheckingAccount(account string) bool {
	return IsSameOrParent(account, "Assets:Checking")
}

func MaxTime(a time.Time, b time.Time) time.Time {
	if a.After(b) {
		return a
	} else {
		return b
	}
}

func GroupByAccount(posts []posting.Posting) map[string][]posting.Posting {
	return lo.GroupBy(posts, func(post posting.Posting) string {
		return post.Account
	})
}

type GroupableByDate interface {
	GroupDate() time.Time
}

func GroupByMonth[G GroupableByDate](groupables []G) map[string][]G {
	grouped := make(map[string][]G)
	for _, g := range groupables {
		key := g.GroupDate().Format("2006-01")
		ps, ok := grouped[key]
		if ok {
			grouped[key] = append(ps, g)
		} else {
			grouped[key] = []G{g}
		}

	}
	return grouped
}

func GroupByFY[G GroupableByDate](groupables []G) map[string][]G {
	grouped := make(map[string][]G)
	for _, g := range groupables {
		key := FYHuman(g.GroupDate())
		ps, ok := grouped[key]
		if ok {
			grouped[key] = append(ps, g)
		} else {
			grouped[key] = []G{g}
		}

	}
	return grouped
}

func SumBy[C any](collection []C, iteratee func(item C) decimal.Decimal) decimal.Decimal {
	return lo.Reduce(collection, func(acc decimal.Decimal, item C, _ int) decimal.Decimal {
		return iteratee(item).Add(acc)
	}, decimal.Zero)
}

func SortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	keys := lo.Keys(m)
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}
