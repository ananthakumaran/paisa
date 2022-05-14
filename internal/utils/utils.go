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

func BeginningOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 0, -date.Day()+1)
}

func EndOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 1, -date.Day())
}
