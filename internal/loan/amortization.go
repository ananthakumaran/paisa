// Package loan provides pure math helpers for amortizing loan schedules
// (equal payment / equal principal). No DB access, no posting awareness.
package loan

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// ScheduleKind selects the amortization method.
type ScheduleKind string

const (
	// EqualPayment is 等额本息: same total payment every month, interest declines.
	EqualPayment ScheduleKind = "equal_payment"
	// EqualPrincipal is 等额本金: same principal portion every month, payment declines.
	EqualPrincipal ScheduleKind = "equal_principal"
)

// Month represents one row of the amortization table. All amounts are rounded
// to 2dp. Invariants enforced by Amortize across the full schedule:
//
//	row.Payment == row.Principal + row.Interest          (every row)
//	sum(row.Principal for row in months) == Principal     (initial principal)
//	last month Balance == 0
type Month struct {
	Index     int             `json:"index"`     // 1-based month number
	Payment   decimal.Decimal `json:"payment"`   // total payment in this month
	Principal decimal.Decimal `json:"principal"` // principal portion
	Interest  decimal.Decimal `json:"interest"`  // interest portion
	Balance   decimal.Decimal `json:"balance"`   // remaining balance after this payment
}

// Schedule is the full amortization output.
type Schedule struct {
	Kind           ScheduleKind    `json:"kind"`
	Principal      decimal.Decimal `json:"principal"`
	APR            decimal.Decimal `json:"apr"`
	TermMonths     int             `json:"term_months"`
	MonthlyRate    decimal.Decimal `json:"monthly_rate"`
	MonthlyPayment decimal.Decimal `json:"monthly_payment"` // for equal_payment: constant payment; for equal_principal: first month's payment
	TotalPayment   decimal.Decimal `json:"total_payment"`
	TotalPrincipal decimal.Decimal `json:"total_principal"`
	TotalInterest  decimal.Decimal `json:"total_interest"`
	Months         []Month         `json:"months"`
}

// decimal helpers
var (
	twelve  = decimal.NewFromInt(12)
	one     = decimal.NewFromInt(1)
	hundred = decimal.NewFromInt(100)
)

// Amortize builds the schedule. Inputs:
//   - principal: initial loan principal (must be > 0)
//   - aprPercent: annual percentage rate, expressed as a percent (e.g. 4.9 for 4.9%)
//   - termMonths: total number of monthly installments (must be > 0)
//   - kind: EqualPayment or EqualPrincipal
//
// Reconciliation strategy: per-row arithmetic runs in full decimal precision.
// Rounding to 2dp happens only at the row-emit step. The last row is then
// adjusted in a reconciliation pass so the following invariants hold exactly:
//
//   - row.Payment == row.Principal + row.Interest (every row, including the last)
//   - sum(row.Principal) == initial principal     (no drift loss)
//   - last row Balance == 0
//
// All arithmetic uses shopspring/decimal; no float64 round-trips.
func Amortize(principal decimal.Decimal, aprPercent decimal.Decimal, termMonths int, kind ScheduleKind) (*Schedule, error) {
	if termMonths <= 0 {
		return nil, errors.New("term_months must be > 0")
	}
	if principal.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("principal must be > 0")
	}
	if aprPercent.LessThan(decimal.Zero) {
		return nil, errors.New("rate must be >= 0")
	}
	if kind != EqualPayment && kind != EqualPrincipal {
		return nil, fmt.Errorf("unknown schedule kind: %s", string(kind))
	}

	// monthly rate as a fraction: apr / 100 / 12
	monthlyRate := aprPercent.Div(hundred).Div(twelve)

	out := &Schedule{
		Kind:        kind,
		Principal:   principal,
		APR:         aprPercent,
		TermMonths:  termMonths,
		MonthlyRate: monthlyRate,
		Months:      make([]Month, 0, termMonths),
	}

	switch kind {
	case EqualPayment:
		buildEqualPayment(out, principal, monthlyRate, termMonths)
	case EqualPrincipal:
		buildEqualPrincipal(out, principal, monthlyRate, termMonths)
	}

	reconcile(out, principal)

	// Totals are summed from the (now reconciled) rounded rows so the UI sees
	// the exact same numbers as the rows.
	totalPrincipal := decimal.Zero
	totalInterest := decimal.Zero
	totalPayment := decimal.Zero
	for _, m := range out.Months {
		totalPrincipal = totalPrincipal.Add(m.Principal)
		totalInterest = totalInterest.Add(m.Interest)
		totalPayment = totalPayment.Add(m.Payment)
	}
	out.TotalPrincipal = totalPrincipal
	out.TotalInterest = totalInterest
	out.TotalPayment = totalPayment

	return out, nil
}

// buildEqualPayment: M = P * r * (1+r)^n / ((1+r)^n - 1) (or P/n when r==0).
//
// All per-row arithmetic stays in exact decimal. Each row's interest is rounded
// from the exact value, and payment is the rounded constant M. Principal is
// then derived as Payment - Interest so the per-row invariant
// `payment == principal + interest` holds by construction. The last-row
// reconciliation pass fixes up cumulative drift.
func buildEqualPayment(out *Schedule, principal, r decimal.Decimal, n int) {
	var monthlyExact decimal.Decimal
	nDec := decimal.NewFromInt(int64(n))

	if r.IsZero() {
		monthlyExact = principal.Div(nDec)
	} else {
		onePlusR := one.Add(r)
		pow := decimalPow(onePlusR, n)
		num := principal.Mul(r).Mul(pow)
		den := pow.Sub(one)
		monthlyExact = num.Div(den)
	}

	paymentR := monthlyExact.Round(2)
	out.MonthlyPayment = paymentR

	// Track remaining principal as a rounded running total so a row never pays
	// down more than is left owed (which would push the reconciled last-row
	// principal negative on tiny-principal corner cases).
	remaining := principal
	balanceExact := principal
	for i := 1; i <= n; i++ {
		interestExact := balanceExact.Mul(r)
		// In exact arithmetic the principal portion is monthlyExact - interestExact;
		// stepping the balance by that exact value keeps every subsequent interest
		// calculation precise.
		principalExact := monthlyExact.Sub(interestExact)
		balanceExact = balanceExact.Sub(principalExact)

		interestR := interestExact.Round(2)
		// Derive principal from payment - interest so the per-row invariant
		// `payment == principal + interest` holds exactly.
		principalR := paymentR.Sub(interestR)
		// Cap principal payment at the remaining balance so we never overpay.
		// reconcile() will absorb any residual mismatch into the final row.
		if principalR.GreaterThan(remaining) {
			principalR = remaining
		}
		if principalR.IsNegative() {
			principalR = decimal.Zero
		}
		remaining = remaining.Sub(principalR)
		// Re-derive payment from the (possibly capped) principal so the per-row
		// invariant still holds.
		thisPaymentR := principalR.Add(interestR)

		out.Months = append(out.Months, Month{
			Index:     i,
			Payment:   thisPaymentR,
			Principal: principalR,
			Interest:  interestR,
			// Balance is recomputed in reconcile() as a running running-sum to
			// guarantee the final balance is exactly zero.
		})
	}
}

// buildEqualPrincipal: principalPart = P/n; interest_k = balance_{k-1} * r.
//
// Per-row principal is computed as a delta of the rounded running cumulative
// target (i = P * k / n, rounded). This guarantees sum(principal) tracks P
// to within at most one cent across the whole schedule (corner cases where
// P/n < 0.005 no longer drift catastrophically). Payment is then derived as
// Principal + Interest so the per-row invariant holds by construction.
// The last-row reconciliation pass absorbs any final cent of drift.
func buildEqualPrincipal(out *Schedule, principal, r decimal.Decimal, n int) {
	nDec := decimal.NewFromInt(int64(n))
	principalPartExact := principal.Div(nDec)

	// Summary "monthly payment" shown to the user is the first month's payment.
	firstInterestExact := principal.Mul(r)
	out.MonthlyPayment = principalPartExact.Add(firstInterestExact).Round(2)

	balanceExact := principal
	prevCumPrincipalR := decimal.Zero
	remaining := principal
	for i := 1; i <= n; i++ {
		interestExact := balanceExact.Mul(r)
		// Step balance by the exact principalPart so subsequent interests stay accurate.
		balanceExact = balanceExact.Sub(principalPartExact)

		// Principal for this row = rounded-cumulative-target - sum-of-previous-principal.
		// This pulls the rounding error back toward zero each row instead of letting
		// it accumulate in one direction.
		cumTargetExact := principalPartExact.Mul(decimal.NewFromInt(int64(i)))
		cumTargetR := cumTargetExact.Round(2)
		principalR := cumTargetR.Sub(prevCumPrincipalR)
		// Cap principal at remaining balance so we never overpay.
		if principalR.GreaterThan(remaining) {
			principalR = remaining
		}
		if principalR.IsNegative() {
			principalR = decimal.Zero
		}
		prevCumPrincipalR = prevCumPrincipalR.Add(principalR)
		remaining = remaining.Sub(principalR)

		interestR := interestExact.Round(2)
		// Derive payment from principal + interest so the per-row invariant holds.
		paymentR := principalR.Add(interestR)

		out.Months = append(out.Months, Month{
			Index:     i,
			Payment:   paymentR,
			Principal: principalR,
			Interest:  interestR,
		})
	}
}

// reconcile performs the final pass that guarantees:
//
//	sum(month.Principal) == initial principal
//	last month Balance    == 0
//	month.Payment         == month.Principal + month.Interest (still holds after fixups)
//
// It does this by overwriting the LAST row's Principal so the principal total
// reconciles to the input, then recomputing Payment as Principal + Interest for
// that row. Balance is then filled in as a running sum so the final value is
// exactly zero.
func reconcile(s *Schedule, principal decimal.Decimal) {
	n := len(s.Months)
	if n == 0 {
		return
	}

	// 1. Absorb principal-rounding drift into the last row.
	sumPrincipalExceptLast := decimal.Zero
	for i := 0; i < n-1; i++ {
		sumPrincipalExceptLast = sumPrincipalExceptLast.Add(s.Months[i].Principal)
	}
	lastPrincipal := principal.Sub(sumPrincipalExceptLast)
	s.Months[n-1].Principal = lastPrincipal
	// Keep the per-row invariant: payment == principal + interest. Interest in
	// the last row is whatever was computed from the exact balance; payment is
	// now derived from the reconciled principal.
	s.Months[n-1].Payment = lastPrincipal.Add(s.Months[n-1].Interest)

	// 2. Fill Balance as a running deduction so the final balance is exactly 0.
	running := principal
	for i := range s.Months {
		running = running.Sub(s.Months[i].Principal)
		s.Months[i].Balance = running
	}
	// Guard against tiny rounding noise in the final balance (e.g. -0.00).
	if !s.Months[n-1].Balance.IsZero() {
		s.Months[n-1].Balance = decimal.Zero
	}
}

// decimalPow computes base^exp for integer exp >= 0 using repeated squaring,
// keeping everything in decimal so we don't lose precision through float64.
func decimalPow(base decimal.Decimal, exp int) decimal.Decimal {
	if exp == 0 {
		return one
	}
	result := one
	b := base
	e := exp
	for e > 0 {
		if e&1 == 1 {
			result = result.Mul(b)
		}
		b = b.Mul(b)
		e >>= 1
	}
	return result
}
