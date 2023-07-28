package transaction

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/samber/lo"
)

type Transaction struct {
	ID       string            `json:"id"`
	Date     time.Time         `json:"date"`
	Payee    string            `json:"payee"`
	Postings []posting.Posting `json:"postings"`
}

func Build(postings []posting.Posting) []Transaction {
	grouped := lo.GroupBy(postings, func(p posting.Posting) string { return p.TransactionID })
	return lo.Map(lo.Values(grouped), func(ps []posting.Posting, _ int) Transaction {
		sample := ps[0]
		return Transaction{ID: sample.TransactionID, Date: sample.Date, Payee: sample.Payee, Postings: ps}
	})

}
