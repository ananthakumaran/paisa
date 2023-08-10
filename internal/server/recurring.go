package server

import (
	"sort"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func GetRecurringTransactions(db *gorm.DB) gin.H {
	return gin.H{"transaction_sequences": ComputeRecurringTransactions(query.Init(db).All())}
}

type TransactionSequence struct {
	Transactions             []transaction.Transaction `json:"transactions"`
	Key                      TransactionSequenceKey    `json:"key"`
	Interval                 int                       `json:"interval"`
	DaysSinceLastTransaction int                       `json:"days_since_last_transaction"`
}

type TransactionSequenceKey struct {
	TagRecurring string `json:"tagRecurring"`
}

func ComputeRecurringTransactions(postings []posting.Posting) []TransactionSequence {
	now := time.Now()

	postings = lo.Filter(postings, func(p posting.Posting, _ int) bool {
		return p.Date.Before(now)
	})

	transactions := transaction.Build(postings)
	transactions = lo.Filter(transactions, func(t transaction.Transaction, _ int) bool {
		return t.TagRecurring != ""
	})
	transactionsGrouped := lo.GroupBy(transactions, func(t transaction.Transaction) TransactionSequenceKey {
		return TransactionSequenceKey{TagRecurring: t.TagRecurring}
	})

	transaction_sequences := lo.MapToSlice(transactionsGrouped, func(key TransactionSequenceKey, ts []transaction.Transaction) TransactionSequence {
		sort.SliceStable(ts, func(i, j int) bool {
			return ts[i].Date.After(ts[j].Date)
		})

		interval := 0
		daysSinceLastTransaction := 0
		if len(ts) > 1 {
			interval = int(ts[0].Date.Sub(ts[1].Date).Hours() / 24)
			daysSinceLastTransaction = int(now.Sub(ts[0].Date).Hours() / 24)
		}

		return TransactionSequence{Transactions: ts, Key: key, Interval: interval, DaysSinceLastTransaction: daysSinceLastTransaction}
	})

	sort.SliceStable(transaction_sequences, func(i, j int) bool {
		return len(transaction_sequences[i].Transactions) > len(transaction_sequences[j].Transactions)
	})

	return lo.Filter(transaction_sequences, func(ts TransactionSequence, _ int) bool {
		return len(ts.Transactions) > 1
	})
}
