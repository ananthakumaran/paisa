package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/google/btree"
	"github.com/onrik/gorm-logrus"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/constraints"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func YearHumanCutOffAt(date time.Time, cutoff time.Time) string {
	if date.Month() < cutoff.Month() || date.Month() == cutoff.Month() && date.Day() < cutoff.Day() {
		return fmt.Sprintf("%d - %d", date.Year()-1, date.Year()%100)
	} else {
		return fmt.Sprintf("%d - %d", date.Year(), (date.Year()+1)%100)
	}
}

func ParseFY(fy string) (time.Time, time.Time) {
	start, _ := time.ParseInLocation("2006", strings.Split(fy, " ")[0], config.TimeZone())
	start = start.AddDate(0, int(config.GetConfig().FinancialYearStartingMonth-time.January), 0)
	return BeginningOfFinancialYear(start), EndOfFinancialYear(start)
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
	return EndOfDay(toDate(date.AddDate(0, 1, -date.Day())))
}

func IsWithDate(date time.Time, start time.Time, end time.Time) bool {
	return (date.Equal(start) || date.After(start)) && (date.Before(end) || date.Equal(end))
}

func IsSameDate(a time.Time, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func EndOfDay(date time.Time) time.Time {
	return toDate(date).AddDate(0, 0, 1).Add(-time.Nanosecond)
}

var now time.Time

func SetNow(date string) {
	t, err := time.ParseInLocation("2006-01-02", date, config.TimeZone())
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Setting now to %s", t)
	now = t
}

func Now() time.Time {
	if !now.Equal(time.Time{}) {
		return now
	}
	return time.Now()
}

func IsNowDefined() bool {
	return !now.Equal(time.Time{})
}

func EndOfToday() time.Time {
	return EndOfDay(Now())
}

func toDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, config.TimeZone())
}

func IsSameOrParent(account string, comparison string) bool {
	if account == comparison {
		return true
	}

	return strings.HasPrefix(account, comparison+":")
}

func FirstName(account string) string {
	return strings.Split(account, ":")[0]
}

func IsParent(account string, comparison string) bool {
	return strings.HasPrefix(account, comparison+":")
}

func IsCurrency(currency string) bool {
	return currency == config.DefaultCurrency()
}

func IsCheckingAccount(account string) bool {
	return IsSameOrParent(account, "Assets:Checking")
}

func IsExpenseInterestAccount(account string) bool {
	return IsSameOrParent(account, "Expenses:Interest")
}

func MaxTime(a time.Time, b time.Time) time.Time {
	if a.After(b) {
		return a
	} else {
		return b
	}
}

type GroupableByDate interface {
	GroupDate() time.Time
}

func GroupByDate[G GroupableByDate](groupables []G) map[string][]G {
	grouped := make(map[string][]G)
	for _, g := range groupables {
		key := g.GroupDate().Format("2006-01-02")
		ps, ok := grouped[key]
		if ok {
			grouped[key] = append(ps, g)
		} else {
			grouped[key] = []G{g}
		}

	}
	return grouped
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

func GroupByYearCutoffAt[G GroupableByDate](groupables []G, date time.Time) map[string][]G {
	grouped := make(map[string][]G)
	for _, g := range groupables {
		key := YearHumanCutOffAt(g.GroupDate(), date)
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

func UnQuote(str string) string {
	if len(str) < 2 {
		return str
	}

	if str[0] == '"' && str[len(str)-1] == '"' {
		return str[1 : len(str)-1]
	}
	return str
}

func OpenDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(config.GetDBPath()), &gorm.Config{Logger: gorm_logrus.New()})
	return db, err
}

func Dos2Unix(str string) string {
	return strings.ReplaceAll(str, "\r\n", "\n")
}

func ReplaceLast(haystack, needle, replacement string) string {
	i := strings.LastIndex(haystack, needle)
	if i == -1 {
		return haystack
	}
	return haystack[:i] + replacement + haystack[i+len(needle):]
}

func Sha256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func BuildSubPath(baseDirectory string, path string) (string, error) {
	baseDirectory = filepath.Clean(baseDirectory)
	fullpath := filepath.Clean(filepath.Join(baseDirectory, filepath.Clean(path)))

	relpath, err := filepath.Rel(baseDirectory, fullpath)
	if err != nil {
		return "", err
	}

	if strings.Contains(relpath, "..") {
		return "", errors.New("Not allowed to refer path outside the base directory")
	}

	return fullpath, nil
}
