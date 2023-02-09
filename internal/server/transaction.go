package server

import (
	"sort"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Transaction struct {
	ID       string            `json:"id"`
	Date     time.Time         `json:"date"`
	Payee    string            `json:"payee"`
	Postings []posting.Posting `json:"postings"`
}

func GetTransactions(db *gorm.DB) gin.H {
	postings := query.Init(db).Desc().All()

	grouped := lo.GroupBy(postings, func(p posting.Posting) string { return p.TransactionID })
	transactions := lo.Map(lo.Values(grouped), func(ps []posting.Posting, _ int) Transaction {
		sample := ps[0]
		return Transaction{ID: sample.TransactionID, Date: sample.Date, Payee: sample.Payee, Postings: ps}
	})

	sort.Slice(transactions, func(i, j int) bool { return transactions[i].ID > transactions[j].ID })
	sort.SliceStable(transactions, func(i, j int) bool { return transactions[i].Date.After(transactions[j].Date) })

	return gin.H{"transactions": transactions}
}
