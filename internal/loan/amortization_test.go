package loan

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAmortizeEqualPayment(t *testing.T) {
	principal := decimal.NewFromInt(1100000)
	apr := decimal.NewFromFloat(4.9)
	schedule, err := Amortize(principal, apr, 240, EqualPayment)
	assert.NoError(t, err)

	// Known-good: M = P * r * (1+r)^n / ((1+r)^n - 1)
	// For P=1,100,000, APR=4.9, n=240 => 7198.884538...
	monthly, _ := schedule.MonthlyPayment.Float64()
	assert.InDelta(t, 7198.88, monthly, 0.05)

	// All 240 months present
	assert.Equal(t, 240, len(schedule.Months))

	// First month
	first := schedule.Months[0]
	assert.Equal(t, 1, first.Index)
	firstInterest, _ := first.Interest.Float64()
	firstPrincipal, _ := first.Principal.Float64()
	firstPayment, _ := first.Payment.Float64()
	// First interest: 1,100,000 * 0.049/12 = 4491.6666...
	assert.InDelta(t, 4491.67, firstInterest, 0.05)
	assert.InDelta(t, monthly-firstInterest, firstPrincipal, 0.05)
	assert.InDelta(t, monthly, firstPayment, 0.05)

	// Last month: balance should reach 0
	last := schedule.Months[len(schedule.Months)-1]
	assert.Equal(t, 240, last.Index)
	lastBalance, _ := last.Balance.Float64()
	assert.InDelta(t, 0, lastBalance, 0.5)

	// Totals
	totalPrincipal := decimal.Zero
	totalInterest := decimal.Zero
	for _, m := range schedule.Months {
		totalPrincipal = totalPrincipal.Add(m.Principal)
		totalInterest = totalInterest.Add(m.Interest)
	}
	tp, _ := totalPrincipal.Float64()
	ti, _ := totalInterest.Float64()
	assert.InDelta(t, 1100000.0, tp, 1.0)
	assert.True(t, ti > 0)
	tpi, _ := schedule.TotalPrincipal.Float64()
	tii, _ := schedule.TotalInterest.Float64()
	assert.InDelta(t, tp, tpi, 0.01)
	assert.InDelta(t, ti, tii, 0.01)
}

func TestAmortizeEqualPrincipal(t *testing.T) {
	principal := decimal.NewFromInt(110000)
	apr := decimal.NewFromFloat(5.0)
	schedule, err := Amortize(principal, apr, 60, EqualPrincipal)
	assert.NoError(t, err)

	assert.Equal(t, 60, len(schedule.Months))

	// Equal principal: principal portion = P/n = 110000/60 = 1833.3333...
	first := schedule.Months[0]
	firstPrincipal, _ := first.Principal.Float64()
	assert.InDelta(t, 1833.33, firstPrincipal, 0.01)

	// First interest: 110000 * 0.05/12 = 458.333...
	firstInterest, _ := first.Interest.Float64()
	assert.InDelta(t, 458.33, firstInterest, 0.01)

	// First payment
	firstPayment, _ := first.Payment.Float64()
	assert.InDelta(t, 2291.67, firstPayment, 0.01)

	// In equal principal, the principal portion is roughly constant; payment declines monthly.
	last := schedule.Months[len(schedule.Months)-1]
	lastPayment, _ := last.Payment.Float64()
	assert.True(t, lastPayment < firstPayment, "last payment should be less than first")

	// Final balance ~0
	lastBalance, _ := last.Balance.Float64()
	assert.InDelta(t, 0, lastBalance, 0.5)

	// Total principal == initial principal
	totalPrincipal := decimal.Zero
	for _, m := range schedule.Months {
		totalPrincipal = totalPrincipal.Add(m.Principal)
	}
	tp, _ := totalPrincipal.Float64()
	assert.InDelta(t, 110000.0, tp, 0.5)
}

func TestAmortizeValidation(t *testing.T) {
	principal := decimal.NewFromInt(100000)
	apr := decimal.NewFromFloat(5.0)

	_, err := Amortize(principal, apr, 0, EqualPayment)
	assert.Error(t, err, "term_months must be > 0")

	_, err = Amortize(decimal.Zero, apr, 12, EqualPayment)
	assert.Error(t, err, "principal must be > 0")

	_, err = Amortize(principal, decimal.NewFromInt(-1), 12, EqualPayment)
	assert.Error(t, err, "rate must be >= 0")

	_, err = Amortize(principal, apr, 12, ScheduleKind("nope"))
	assert.Error(t, err, "unknown schedule kind")
}

func TestAmortizeZeroRate(t *testing.T) {
	// Edge case: 0% interest -- equal payment becomes P/n with zero interest.
	principal := decimal.NewFromInt(12000)
	schedule, err := Amortize(principal, decimal.Zero, 12, EqualPayment)
	assert.NoError(t, err)
	monthly, _ := schedule.MonthlyPayment.Float64()
	assert.InDelta(t, 1000.0, monthly, 0.001)
	for _, m := range schedule.Months {
		i, _ := m.Interest.Float64()
		assert.InDelta(t, 0, i, 0.001)
	}
}
