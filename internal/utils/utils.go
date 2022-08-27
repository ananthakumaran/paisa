package utils

import (
	"github.com/google/btree"
	"time"
)

func BTreeDescendFirstLessOrEqual[I btree.Item](tree *btree.BTree, item I) I {
	var hit I
	tree.DescendLessOrEqual(item, func(item btree.Item) bool {
		hit = item.(I)
		return false
	})

	return hit
}

func BeginningOfFinancialYear(date time.Time) time.Time {
	beginningOfMonth := BeginningOfMonth(date)
	if beginningOfMonth.Month() < time.April {
		return beginningOfMonth.AddDate(-1, int(time.April-beginningOfMonth.Month()), 0)
	} else {
		return beginningOfMonth.AddDate(0, -int(beginningOfMonth.Month()-time.April), 0)
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

func toDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}
