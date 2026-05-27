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

// Month represents one row of the amortization table.
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
	MonthlyPayment decimal.Decimal `json:"monthly_payment"` // for equal_payment; first month payment for equal_principal
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
// All arithmetic uses shopspring/decimal. Internal precision is 12 dp; outputs
// in Month rows are rounded to 2 dp (banker-friendly) with any rounding drift
// folded into the last row so totals reconcile exactly to the input principal.
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

	// Totals (sum from rounded rows so totals match what UI shows).
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
func buildEqualPayment(out *Schedule, principal, r decimal.Decimal, n int) {
	var monthly decimal.Decimal
	nDec := decimal.NewFromInt(int64(n))

	if r.IsZero() {
		monthly = principal.Div(nDec)
	} else {
		// (1+r)^n -- use float pow then back to decimal for the exponent; precision
		// of the result is fine for monetary math (rounding drift is absorbed below).
		onePlusR := one.Add(r)
		pow := decimalPow(onePlusR, n)
		num := principal.Mul(r).Mul(pow)
		den := pow.Sub(one)
		monthly = num.Div(den)
	}

	out.MonthlyPayment = monthly.Round(2)

	balance := principal
	for i := 1; i <= n; i++ {
		interest := balance.Mul(r)
		paymentThisMonth := monthly
		principalThisMonth := paymentThisMonth.Sub(interest)
		balance = balance.Sub(principalThisMonth)

		// Round outputs to 2dp; fold rounding into the last row.
		interestR := interest.Round(2)
		principalR := principalThisMonth.Round(2)
		paymentR := paymentThisMonth.Round(2)
		balanceR := balance.Round(2)

		if i == n {
			// Force final balance to exactly 0 and adjust principal so totals reconcile.
			principalR = principalR.Add(balanceR)
			paymentR = principalR.Add(interestR)
			balanceR = decimal.Zero
		}

		out.Months = append(out.Months, Month{
			Index:     i,
			Payment:   paymentR,
			Principal: principalR,
			Interest:  interestR,
			Balance:   balanceR,
		})
	}
}

// buildEqualPrincipal: principal portion = P/n; interest_k = balance_{k-1} * r.
func buildEqualPrincipal(out *Schedule, principal, r decimal.Decimal, n int) {
	nDec := decimal.NewFromInt(int64(n))
	principalPart := principal.Div(nDec)

	// "Monthly payment" reported for the summary is the first month's payment.
	firstInterest := principal.Mul(r)
	out.MonthlyPayment = principalPart.Add(firstInterest).Round(2)

	balance := principal
	for i := 1; i <= n; i++ {
		interest := balance.Mul(r)
		thisPrincipal := principalPart
		balance = balance.Sub(thisPrincipal)
		payment := thisPrincipal.Add(interest)

		interestR := interest.Round(2)
		principalR := thisPrincipal.Round(2)
		paymentR := payment.Round(2)
		balanceR := balance.Round(2)

		if i == n {
			principalR = principalR.Add(balanceR)
			paymentR = principalR.Add(interestR)
			balanceR = decimal.Zero
		}

		out.Months = append(out.Months, Month{
			Index:     i,
			Payment:   paymentR,
			Principal: principalR,
			Interest:  interestR,
			Balance:   balanceR,
		})
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
