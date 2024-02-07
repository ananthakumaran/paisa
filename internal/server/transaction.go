package server

import (
	"sort"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

func GetTransactions(db *gorm.DB) gin.H {
	postings := query.Init(db).Desc().All()
	transactions := transaction.Build(postings)

	sort.Slice(transactions, func(i, j int) bool { return transactions[i].ID > transactions[j].ID })
	sort.SliceStable(transactions, func(i, j int) bool { return transactions[i].Date.After(transactions[j].Date) })

	return gin.H{"transactions": transactions}
}

func GetBalancedPostings(db *gorm.DB) gin.H {
	postings := query.Init(db).Desc().All()
	transactions := transaction.Build(postings)
	balancePostings := accounting.BuildBalancedPostings(transactions)

	return gin.H{"balancedPostings": balancePostings}
}

func GetLatestTransactions(db *gorm.DB) []transaction.Transaction {
	postings := query.Init(db).Desc().Limit(200).All()
	transactions := transaction.Build(postings)

	sort.Slice(transactions, func(i, j int) bool { return transactions[i].ID > transactions[j].ID })
	sort.SliceStable(transactions, func(i, j int) bool { return transactions[i].Date.After(transactions[j].Date) })

	return transactions
}
