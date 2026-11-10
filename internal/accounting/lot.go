package accounting

import (
	"sort"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type split struct {
	date  time.Time
	ratio decimal.Decimal
}

// Lot represents the units acquired via a single posting. The units
// are as on the date of sale, i.e., after applying the stock splits,
// while the date and the total cost are that of the original purchase.
type Lot struct {
	posting.Posting
	splits []split
}

// SplitRatioAfter returns the number of units held now for each unit
// held on the given date.
func (l Lot) SplitRatioAfter(date time.Time) decimal.Decimal {
	ratio := decimal.NewFromInt(1)
	for _, s := range l.splits {
		if s.date.After(date) {
			ratio = ratio.Mul(s.ratio)
		}
	}
	return ratio
}

type Sale struct {
	Posting posting.Posting
	Lots    []Lot
}

type lotMatcher struct {
	db        *gorm.DB
	available []Lot
	sales     []Sale
}

// MatchLots matches the sales of an account with the purchases based
// on FIFO. It returns the sales along with the matched lots and the
// lots that are still available. Stock split carries the existing lots
// forward with new quantity and bonus units are acquired at zero cost.
func MatchLots(db *gorm.DB, postings []posting.Posting) ([]Sale, []Lot) {
	m := lotMatcher{db: db}
	byDate := utils.GroupByDate(postings)
	for _, date := range utils.SortedKeys(byDate) {
		m.processDay(byDate[date])
	}
	return m.sales, m.available
}

// processDay applies the splits and bonuses before the trades, as a
// stock trades at the adjusted price from the beginning of the day.
func (m *lotMatcher) processDay(postings []posting.Posting) {
	var actions, trades, pendingSplits []posting.Posting
	for _, p := range postings {
		if _, _, isSplit := service.StockSplitUnits(m.db, p); isSplit {
			actions = append(actions, p)
		} else if service.IsBonus(m.db, p) {
			p.Amount = decimal.Zero
			actions = append(actions, p)
		} else {
			trades = append(trades, p)
		}
	}

	// retain the order in which they are specified in the journal.
	sort.SliceStable(actions, func(i, j int) bool { return actions[i].ID < actions[j].ID })
	splitTransactions := make(map[string]bool)
	for _, p := range actions {
		removed, added, isSplit := service.StockSplitUnits(m.db, p)
		if !isSplit {
			m.trade(p)
			continue
		}

		if splitTransactions[p.TransactionID] {
			continue
		}
		splitTransactions[p.TransactionID] = true

		if m.holding(p.Commodity).LessThan(removed) {
			// units might have been bought on the same day
			pendingSplits = append(pendingSplits, p)
		} else {
			m.split(p, removed, added)
		}
	}

	for _, p := range trades {
		m.trade(p)
	}

	for _, p := range pendingSplits {
		removed, added, _ := service.StockSplitUnits(m.db, p)
		m.split(p, removed, added)
	}
}

func (m *lotMatcher) holding(commodity string) decimal.Decimal {
	return utils.SumBy(m.available, func(l Lot) decimal.Decimal {
		if l.Commodity == commodity {
			return l.Quantity
		}
		return decimal.Zero
	})
}

// split distributes the units among the existing lots in proportion to
// their quantity. A split always applies to all the units held, so it
// doesn't matter whether the split is specified for the whole holding
// or for each lot via separate transaction.
func (m *lotMatcher) split(p posting.Posting, removed decimal.Decimal, added decimal.Decimal) {
	before := m.holding(p.Commodity)
	if before.IsZero() {
		return
	}

	after := before.Sub(removed).Add(added)
	if before.LessThan(removed) {
		after = before.Mul(added).Div(removed)
	}
	ratio := after.Div(before)

	remaining := after
	_, last, _ := lo.FindLastIndexOf(m.available, func(l Lot) bool { return l.Commodity == p.Commodity })
	for i := range m.available {
		l := &m.available[i]
		if l.Commodity != p.Commodity {
			continue
		}

		if i == last {
			// avoid losing units due to rounding
			l.Quantity = remaining
		} else {
			l.Quantity = l.Quantity.Mul(after).Div(before)
			remaining = remaining.Sub(l.Quantity)
		}
		l.splits = append(l.splits, split{date: p.Date, ratio: ratio})
	}
}

func (m *lotMatcher) trade(p posting.Posting) {
	if p.Quantity.IsPositive() {
		m.available = append(m.available, Lot{Posting: p})
	} else if p.Quantity.IsNegative() {
		m.sell(p)
	}
}

func (m *lotMatcher) sell(p posting.Posting) {
	p.Amount = service.SellAmount(m.db, p)
	sale := Sale{Posting: p}
	quantity := p.Quantity.Neg()
	for quantity.IsPositive() && len(m.available) > 0 {
		sold := m.available[0]
		if sold.Quantity.GreaterThan(quantity) {
			sold.Posting = sold.WithQuantity(quantity)
			m.available[0].AddQuantity(quantity.Neg())
		} else {
			m.available = m.available[1:]
		}
		quantity = quantity.Sub(sold.Quantity)
		sale.Lots = append(sale.Lots, sold)
	}
	m.sales = append(m.sales, sale)
}
