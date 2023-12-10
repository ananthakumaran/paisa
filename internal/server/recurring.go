package server

import (
	"sort"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func GetRecurringTransactions(db *gorm.DB) gin.H {
	return gin.H{"transaction_sequences": ComputeRecurringTransactions(query.Init(db).All())}
}

type TransactionSequence struct {
	Transactions []transaction.Transaction `json:"transactions"`
	Key          string                    `json:"key"`
	Period       string                    `json:"period"`
	Interval     int                       `json:"interval"`
}

func ComputeRecurringTransactions(postings []posting.Posting) []TransactionSequence {
	now := utils.EndOfToday()

	postings = lo.Filter(postings, func(p posting.Posting, _ int) bool {
		return p.Date.Before(now)
	})

	transactions := transaction.Build(postings)
	transactions = lo.Filter(transactions, func(t transaction.Transaction, _ int) bool {
		return t.TagRecurring != ""
	})
	transactionsGrouped := lo.GroupBy(transactions, func(t transaction.Transaction) string {
		return t.TagRecurring
	})

	transaction_sequences := lo.MapToSlice(transactionsGrouped, func(key string, ts []transaction.Transaction) TransactionSequence {
		sort.SliceStable(ts, func(i, j int) bool {
			return ts[i].Date.After(ts[j].Date)
		})

		interval := 0
		var period string
		if ts[0].TagPeriod != "" {
			period = ts[0].TagPeriod
		}

		if len(ts) > 1 {
			for l := 0; l < len(ts)-1; l++ {
				interval = int(ts[l].Date.Sub(ts[l+1].Date).Hours() / 24)
				if interval > 0 {
					break
				}
			}
		}

		return TransactionSequence{Transactions: ts, Key: key, Interval: interval, Period: period}
	})

	sort.SliceStable(transaction_sequences, func(i, j int) bool {
		return len(transaction_sequences[i].Transactions) > len(transaction_sequences[j].Transactions)
	})

	return lo.Filter(transaction_sequences, func(ts TransactionSequence, _ int) bool {
		return len(ts.Transactions) > 1 && ts.Interval > 0
	})
}
