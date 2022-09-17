package accounting

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
)

type Balance struct {
	Date      time.Time
	Commodity string
	Quantity  float64
}

func Register(postings []posting.Posting) []Balance {
	balances := make([]Balance, 0)
	current := Balance{Quantity: 0}
	for _, p := range postings {
		sameDay := p.Date == current.Date
		current = Balance{Date: p.Date, Quantity: p.Quantity + current.Quantity, Commodity: p.Commodity}
		if sameDay {
			balances = balances[:len(balances)-1]
		}
		balances = append(balances, current)

	}
	return balances
}
