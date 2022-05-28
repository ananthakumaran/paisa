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
	return date.AddDate(0, 0, -date.Day()+1)
}

func EndOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 1, -date.Day())
}
