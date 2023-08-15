package transaction

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/samber/lo"
)

type Transaction struct {
	ID           string            `json:"id"`
	Date         time.Time         `json:"date"`
	Payee        string            `json:"payee"`
	Postings     []posting.Posting `json:"postings"`
	TagRecurring string            `json:"tag_recurring"`
	BeginLine    uint64            `json:"beginLine"`
	EndLine      uint64            `json:"endLine"`
	FileName     string            `json:"fileName"`
}

func Build(postings []posting.Posting) []Transaction {
	grouped := lo.GroupBy(postings, func(p posting.Posting) string { return p.TransactionID })
	return lo.Map(lo.Values(grouped), func(ps []posting.Posting, _ int) Transaction {
		sample := ps[0]
		var tagRecurring string
		for _, p := range ps {
			if p.TagRecurring != "" {
				tagRecurring = p.TagRecurring
				break
			}
		}
		return Transaction{ID: sample.TransactionID, Date: sample.Date, Payee: sample.Payee, Postings: ps, TagRecurring: tagRecurring, BeginLine: sample.TransactionBeginLine, EndLine: sample.TransactionEndLine, FileName: sample.FileName}
	})

}
