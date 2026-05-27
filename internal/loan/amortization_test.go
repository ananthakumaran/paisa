package loan

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertInvariants enforces the reconciliation contract on the schedule:
//
//   - every row: payment == principal + interest
//   - sum(principal) == initial principal (no drift)
//   - sum(payment)   == sum(principal) + sum(interest)
//   - last row balance == 0
//   - schedule.TotalPrincipal / TotalInterest / TotalPayment match the row sums
func assertInvariants(t *testing.T, principal decimal.Decimal, s *Schedule) {
	t.Helper()
	require.NotNil(t, s)
	require.NotEmpty(t, s.Months, "schedule must have at least one row")

	sumPrincipal := decimal.Zero
	sumInterest := decimal.Zero
	sumPayment := decimal.Zero
	for i, m := range s.Months {
		// Per-row invariant: payment == principal + interest, exactly.
		require.True(t, m.Principal.Add(m.Interest).Equal(m.Payment),
			"row %d: payment=%s != principal=%s + interest=%s (sum=%s)",
			i+1, m.Payment, m.Principal, m.Interest, m.Principal.Add(m.Interest))
		// Per-row: amounts are non-negative.
		require.False(t, m.Principal.IsNegative(), "row %d: principal must be non-negative", i+1)
		require.False(t, m.Interest.IsNegative(), "row %d: interest must be non-negative", i+1)
		require.False(t, m.Payment.IsNegative(), "row %d: payment must be non-negative", i+1)

		sumPrincipal = sumPrincipal.Add(m.Principal)
		sumInterest = sumInterest.Add(m.Interest)
		sumPayment = sumPayment.Add(m.Payment)
	}

	// Reconciliation invariant: total principal equals the input principal.
	require.True(t, sumPrincipal.Equal(principal),
		"sum(principal)=%s must equal initial principal=%s", sumPrincipal, principal)

	// Reconciliation invariant: total payment equals total principal + total interest.
	require.True(t, sumPayment.Equal(sumPrincipal.Add(sumInterest)),
		"sum(payment)=%s must equal sum(principal)=%s + sum(interest)=%s",
		sumPayment, sumPrincipal, sumInterest)

	// Last row balance is exactly zero.
	last := s.Months[len(s.Months)-1]
	require.True(t, last.Balance.IsZero(), "last balance must be 0, got %s", last.Balance)

	// Schedule totals match what we summed from rows.
	require.True(t, s.TotalPrincipal.Equal(sumPrincipal),
		"schedule.TotalPrincipal=%s != sum=%s", s.TotalPrincipal, sumPrincipal)
	require.True(t, s.TotalInterest.Equal(sumInterest),
		"schedule.TotalInterest=%s != sum=%s", s.TotalInterest, sumInterest)
	require.True(t, s.TotalPayment.Equal(sumPayment),
		"schedule.TotalPayment=%s != sum=%s", s.TotalPayment, sumPayment)

	// Index is 1-based and contiguous.
	for i, m := range s.Months {
		require.Equal(t, i+1, m.Index, "row index mismatch at position %d", i)
	}
}

func TestAmortizeEqualPayment(t *testing.T) {
	principal := decimal.NewFromInt(1100000)
	apr := decimal.NewFromFloat(4.9)
	schedule, err := Amortize(principal, apr, 240, EqualPayment)
	require.NoError(t, err)

	assertInvariants(t, principal, schedule)

	// Known-good: M = P * r * (1+r)^n / ((1+r)^n - 1)
	// For P=1,100,000, APR=4.9, n=240 => 7198.884538...
	monthly, _ := schedule.MonthlyPayment.Float64()
	assert.InDelta(t, 7198.88, monthly, 0.01)

	assert.Equal(t, 240, len(schedule.Months))

	// First month
	first := schedule.Months[0]
	assert.Equal(t, 1, first.Index)
	firstInterest, _ := first.Interest.Float64()
	// First interest: 1,100,000 * 0.049/12 = 4491.6666... -> 4491.67
	assert.InDelta(t, 4491.67, firstInterest, 0.01)
}

func TestAmortizeEqualPrincipal(t *testing.T) {
	principal := decimal.NewFromInt(110000)
	apr := decimal.NewFromFloat(5.0)
	schedule, err := Amortize(principal, apr, 60, EqualPrincipal)
	require.NoError(t, err)

	assertInvariants(t, principal, schedule)

	assert.Equal(t, 60, len(schedule.Months))

	// Equal principal: principal portion = P/n = 110000/60 = 1833.3333... -> 1833.33
	first := schedule.Months[0]
	firstPrincipal, _ := first.Principal.Float64()
	assert.InDelta(t, 1833.33, firstPrincipal, 0.01)

	// First interest: 110000 * 0.05/12 = 458.333...
	firstInterest, _ := first.Interest.Float64()
	assert.InDelta(t, 458.33, firstInterest, 0.01)

	// First payment is principal + interest (1833.33 + 458.33 = 2291.66).
	firstPayment, _ := first.Payment.Float64()
	assert.InDelta(t, 2291.66, firstPayment, 0.01)

	// Payment declines over time in equal-principal.
	last := schedule.Months[len(schedule.Months)-1]
	lastPayment, _ := last.Payment.Float64()
	assert.True(t, lastPayment < firstPayment, "last payment should be less than first")
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
	require.NoError(t, err)
	assertInvariants(t, principal, schedule)

	monthly, _ := schedule.MonthlyPayment.Float64()
	assert.InDelta(t, 1000.0, monthly, 0.001)
	for _, m := range schedule.Months {
		assert.True(t, m.Interest.IsZero(), "zero-rate row must have zero interest")
	}

	// Equal principal at 0%.
	sched2, err := Amortize(principal, decimal.Zero, 12, EqualPrincipal)
	require.NoError(t, err)
	assertInvariants(t, principal, sched2)
	for _, m := range sched2.Months {
		assert.True(t, m.Interest.IsZero(), "zero-rate equal_principal row must have zero interest")
	}
}

// TestAmortizeReconciliation runs the invariant check across a representative
// matrix of (principal, APR, term, schedule) combinations. These are the
// concrete cases the reviewer flagged in the first review iteration; without
// the new reconciliation pass they would all fail.
func TestAmortizeReconciliation(t *testing.T) {
	cases := []struct {
		name      string
		principal decimal.Decimal
		apr       decimal.Decimal
		term      int
		kind      ScheduleKind
	}{
		// Reviewer's BLOCKER cases:
		{"EP_1.1M_4.9_240", decimal.NewFromInt(1100000), decimal.NewFromFloat(4.9), 240, EqualPrincipal},
		{"EP_110K_5_60", decimal.NewFromInt(110000), decimal.NewFromInt(5), 60, EqualPrincipal},
		{"EQ_5M_6.25_360", decimal.NewFromInt(5000000), decimal.NewFromFloat(6.25), 360, EqualPayment},
		{"EQ_110K_5_60", decimal.NewFromInt(110000), decimal.NewFromInt(5), 60, EqualPayment},
		// Extra coverage:
		{"EQ_1.1M_4.9_240", decimal.NewFromInt(1100000), decimal.NewFromFloat(4.9), 240, EqualPayment},
		{"EP_5M_6.25_360", decimal.NewFromInt(5000000), decimal.NewFromFloat(6.25), 360, EqualPrincipal},
		{"EQ_100_3.5_6", decimal.NewFromInt(100), decimal.NewFromFloat(3.5), 6, EqualPayment},
		{"EP_100_3.5_6", decimal.NewFromInt(100), decimal.NewFromFloat(3.5), 6, EqualPrincipal},
		{"EQ_1_5_1", decimal.NewFromInt(1), decimal.NewFromFloat(5.0), 1, EqualPayment},
		{"EP_1_5_1", decimal.NewFromInt(1), decimal.NewFromFloat(5.0), 1, EqualPrincipal},
		{"EQ_99999_7.5_84", decimal.NewFromInt(99999), decimal.NewFromFloat(7.5), 84, EqualPayment},
		{"EP_99999_7.5_84", decimal.NewFromInt(99999), decimal.NewFromFloat(7.5), 84, EqualPrincipal},
		// Fractional/odd-cent principal to exercise rounding paths.
		{"EQ_50000.07_3.65_120", decimal.RequireFromString("50000.07"), decimal.RequireFromString("3.65"), 120, EqualPayment},
		{"EP_50000.07_3.65_120", decimal.RequireFromString("50000.07"), decimal.RequireFromString("3.65"), 120, EqualPrincipal},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s, err := Amortize(tc.principal, tc.apr, tc.term, tc.kind)
			require.NoError(t, err)
			assertInvariants(t, tc.principal, s)
			assert.Equal(t, tc.term, len(s.Months),
				"%s: expected %d months, got %d", tc.name, tc.term, len(s.Months))
		})
	}
}

// TestAmortizeEqualPaymentRows spot-checks specific row values for the case
// the reviewer flagged (110K / 5% / 60mo equal_payment) to make sure the
// principal+interest=payment invariant holds in every row, not just on average.
func TestAmortizeEqualPaymentRows(t *testing.T) {
	principal := decimal.NewFromInt(110000)
	apr := decimal.NewFromInt(5)
	s, err := Amortize(principal, apr, 60, EqualPayment)
	require.NoError(t, err)
	assertInvariants(t, principal, s)

	// Reviewer specifically called out months 12, 14, 17, 20, 23, 26, 30, 31, 34,
	// 37, 38, 39, 41, 42, 43, 49, 51, 53, 56, 58, 59 as broken under the old
	// per-row-independent rounding. Force-check each one.
	suspectIndexes := []int{12, 14, 17, 20, 23, 26, 30, 31, 34, 37, 38, 39, 41, 42, 43, 49, 51, 53, 56, 58, 59}
	for _, idx := range suspectIndexes {
		m := s.Months[idx-1]
		require.True(t, m.Principal.Add(m.Interest).Equal(m.Payment),
			"row %d: payment=%s but principal+interest=%s", idx, m.Payment,
			m.Principal.Add(m.Interest))
	}
}

// TestAmortizeEqualPrincipalSumExact verifies that the EqualPrincipal sum
// reconciliation problem the reviewer found (1.1M / 4.9 / 240 off by 0.80) is
// now exactly zero.
func TestAmortizeEqualPrincipalSumExact(t *testing.T) {
	cases := []struct {
		name      string
		principal decimal.Decimal
		apr       decimal.Decimal
		term      int
	}{
		{"1.1M_4.9_240", decimal.NewFromInt(1100000), decimal.NewFromFloat(4.9), 240},
		{"110K_5_60", decimal.NewFromInt(110000), decimal.NewFromInt(5), 60},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s, err := Amortize(tc.principal, tc.apr, tc.term, EqualPrincipal)
			require.NoError(t, err)
			sum := decimal.Zero
			for _, m := range s.Months {
				sum = sum.Add(m.Principal)
			}
			require.True(t, sum.Equal(tc.principal),
				"%s: sum(principal)=%s, want %s (drift=%s)",
				tc.name, sum, tc.principal, tc.principal.Sub(sum))
		})
	}
}

// TestAmortizeEqualPaymentSumExact verifies that the EqualPayment sum
// reconciliation problem the reviewer found (5M / 6.25 / 360 off by 0.01) is
// now exactly zero.
func TestAmortizeEqualPaymentSumExact(t *testing.T) {
	principal := decimal.NewFromInt(5000000)
	apr := decimal.NewFromFloat(6.25)
	s, err := Amortize(principal, apr, 360, EqualPayment)
	require.NoError(t, err)

	sum := decimal.Zero
	for _, m := range s.Months {
		sum = sum.Add(m.Principal)
	}
	require.True(t, sum.Equal(principal),
		"sum(principal)=%s, want %s (drift=%s)", sum, principal, principal.Sub(sum))
}

// Ensure errors include the value when format-able.
func TestAmortizeErrorMessages(t *testing.T) {
	_, err := Amortize(decimal.NewFromInt(100), decimal.NewFromInt(5), 12, ScheduleKind("bogus"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogus", "error should include the unknown kind")
}

// Sanity check: balance is the running deduction of principal from the input.
func TestAmortizeBalanceRunningSum(t *testing.T) {
	principal := decimal.NewFromInt(12000)
	s, err := Amortize(principal, decimal.NewFromInt(6), 12, EqualPayment)
	require.NoError(t, err)
	assertInvariants(t, principal, s)

	running := principal
	for i, m := range s.Months {
		running = running.Sub(m.Principal)
		require.True(t, m.Balance.Equal(running),
			"row %d: balance=%s but running=%s", i+1, m.Balance, running)
	}
}

// Stress: a wide sweep of small parameter combinations should all satisfy
// invariants. This catches edge cases the named tests above might miss.
func TestAmortizeInvariantSweep(t *testing.T) {
	principals := []decimal.Decimal{
		decimal.NewFromInt(1),
		decimal.NewFromInt(100),
		decimal.NewFromInt(12345),
		decimal.NewFromInt(1000000),
	}
	rates := []decimal.Decimal{
		decimal.Zero,
		decimal.NewFromFloat(0.01),
		decimal.NewFromInt(3),
		decimal.NewFromFloat(7.25),
	}
	terms := []int{1, 2, 3, 6, 12, 60, 120}
	kinds := []ScheduleKind{EqualPayment, EqualPrincipal}

	for _, p := range principals {
		for _, r := range rates {
			for _, n := range terms {
				for _, k := range kinds {
					name := fmt.Sprintf("p=%s_r=%s_n=%d_k=%s", p, r, n, k)
					s, err := Amortize(p, r, n, k)
					require.NoErrorf(t, err, "%s: unexpected error", name)
					assertInvariants(t, p, s)
				}
			}
		}
	}
}
