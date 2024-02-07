package accounting

import (
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type BalancedPosting struct {
	From posting.Posting `json:"from"`
	To   posting.Posting `json:"to"`
}

type PostingPair struct {
	Posting        posting.Posting
	CounterPosting posting.Posting
}

func balancePostings(postings []posting.Posting) []PostingPair {
	EPSILON := decimal.NewFromFloat(0.01)

	var pairs []PostingPair
	pending := slices.Clone(postings)
	for len(pending) > 0 {
		var pair PostingPair
		pair.Posting = pending[0]
		pending = pending[1:]
		found := false
		for i, p := range pending {
			if pair.Posting.Commodity == p.Commodity && pair.Posting.Quantity.Neg().Equal(p.Quantity) {
				pair.CounterPosting = p
				pending = slices.Delete(pending, i, i+1)
				pairs = append(pairs, pair)
				found = true
				break
			}
		}

		if !found {
			for i, p := range pending {
				if pair.Posting.Amount.Neg().Equal(p.Amount) {
					pair.CounterPosting = p
					pending = slices.Delete(pending, i, i+1)
					pairs = append(pairs, pair)
					found = true
					break
				}
			}
		}

		if !found {
		RESTART:
			for i, p := range pending {
				if (pair.Posting.Amount.Sign() == 1 && p.Amount.Sign() == -1) ||
					(pair.Posting.Amount.Sign() == -1 && p.Amount.Sign() == 1) {

					if pair.Posting.Amount.Abs().Equal(p.Amount.Abs()) {
						pair.CounterPosting = p
						pending = slices.Delete(pending, i, i+1)
						pairs = append(pairs, pair)
						found = true
						break
					} else if pair.Posting.Amount.Abs().LessThan(p.Amount.Abs()) {
						counter, remaining := p.Split(pair.Posting.Amount.Neg())
						pair.CounterPosting = counter
						pending[i] = remaining
						pairs = append(pairs, pair)
						found = true
						break
					} else {
						current, remaining := pair.Posting.Split(p.Amount.Neg())
						pair.Posting = current
						pair.CounterPosting = p
						pending = slices.Delete(pending, i, i+1)
						pairs = append(pairs, pair)

						pair = PostingPair{}
						pair.Posting = remaining
						goto RESTART
					}
				}
			}
		}

		if !found && pair.Posting.Amount.Abs().GreaterThan(EPSILON) {
			log.Infof("No counter posting found for %v \npending: %v \npairs: %v e: %v", pair.Posting, pending, pairs, EPSILON)
			break
		}
	}
	return pairs
}

func BuildBalancedPostings(transactions []transaction.Transaction) []BalancedPosting {
	var balancedPostings []BalancedPosting
	for _, t := range transactions {
		postings := t.Postings
		pairs := balancePostings(postings)
		for _, pair := range pairs {
			var from, to posting.Posting
			if pair.Posting.Quantity.IsPositive() {
				to = pair.Posting
				from = pair.CounterPosting
			} else {
				to = pair.CounterPosting
				from = pair.Posting
			}

			balancedPostings = append(balancedPostings, BalancedPosting{
				From: from,
				To:   to,
			})
		}
	}
	return balancedPostings
}
