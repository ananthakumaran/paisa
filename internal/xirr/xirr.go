package xirr

import (
	"math"
	"sort"
	"time"

	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

func daysBetween(start, end time.Time) float64 {
	const millisecondsPerDay = 1000 * 60 * 60 * 24
	millisBetween := end.Sub(start).Milliseconds()
	return float64(millisBetween) / millisecondsPerDay
}

type Transaction struct {
	Years  float64
	Amount float64
}

type Cashflow struct {
	Date   time.Time
	Amount float64
}

func newtonXIRR(transactions []Transaction, initialGuess float64) (float64, bool) {
	x := initialGuess
	const MAX_TRIES = 100
	const EPSILON = 1.0e-6

	for tries := 0; tries < MAX_TRIES; tries++ {
		fxs := 0.0
		dfxs := 0.0
		for _, tx := range transactions {
			fx := tx.Amount / (math.Pow(1.0+x, tx.Years))
			dfx := (-tx.Years * tx.Amount) / (math.Pow(1.0+x, tx.Years+1))
			fxs += fx
			dfxs += dfx
		}

		xNew := x - fxs/dfxs
		if math.IsNaN(xNew) {
			return 0, false
		}
		epsilon := math.Abs(xNew - x)
		if epsilon <= EPSILON {
			return x, true
		}
		x = xNew
	}
	return 0, false
}

func calculateXIRR(transactions []Transaction, initialGuess float64) float64 {
	if x, ok := newtonXIRR(transactions, initialGuess); ok {
		return x
	}

	guess := -0.99

	for guess < 1.0 {
		if x, ok := newtonXIRR(transactions, guess); ok {
			return x
		}
		guess += 0.01
	}

	log.Warn("XIRR didn't converge")
	return 0
}

func XIRR(cashflows []Cashflow) decimal.Decimal {
	if len(cashflows) == 0 {
		return decimal.Zero
	}

	sort.Slice(cashflows, func(i, j int) bool { return cashflows[i].Date.Before(cashflows[j].Date) })
	transactions := lo.Map(cashflows, func(cf Cashflow, _ int) Transaction {
		return Transaction{
			Years:  daysBetween(cashflows[0].Date, cf.Date) / 365,
			Amount: cf.Amount,
		}
	})
	return decimal.NewFromFloat(calculateXIRR(transactions, 0.1) * 100).Round(2)
}
