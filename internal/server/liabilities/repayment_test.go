package liabilities

import (
	"testing"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildAmortizationsSkipsNonAmortizing(t *testing.T) {
	liabs := []config.Liability{
		{Name: "Liabilities:CreditCard:Chase"}, // no kind => skipped
		{
			Name:       "Liabilities:Mortgage:House",
			Kind:       config.AmortizingLoan,
			Principal:  1100000,
			Rate:       4.9,
			TermMonths: 240,
			Schedule:   config.LiabilityEqualPayment,
		},
	}
	out := buildAmortizations(liabs)
	assert.Len(t, out, 1)
	assert.Equal(t, "Liabilities:Mortgage:House", out[0].Name)
	assert.Equal(t, "amortizing_loan", out[0].Kind)
	assert.Equal(t, "equal_payment", out[0].Schedule)
	assert.Len(t, out[0].Months, 240)
	m, _ := out[0].MonthlyPayment.Float64()
	assert.InDelta(t, 7198.88, m, 0.05)
}

func TestBuildAmortizationsEqualPrincipal(t *testing.T) {
	liabs := []config.Liability{
		{
			Name:       "Liabilities:Loan:Car",
			Kind:       config.AmortizingLoan,
			Principal:  110000,
			Rate:       5.0,
			TermMonths: 60,
			Schedule:   config.LiabilityEqualPrincipal,
		},
	}
	out := buildAmortizations(liabs)
	assert.Len(t, out, 1)
	assert.Equal(t, "equal_principal", out[0].Schedule)
	assert.Len(t, out[0].Months, 60)
	p, _ := out[0].Months[0].Principal.Float64()
	assert.InDelta(t, 1833.33, p, 0.01)
}

func TestBuildAmortizationsEmpty(t *testing.T) {
	out := buildAmortizations(nil)
	assert.NotNil(t, out)
	assert.Len(t, out, 0)
}

func TestBuildAmortizationsBackwardCompat(t *testing.T) {
	// A liability with no kind set should be silently skipped (no error, no entry).
	liabs := []config.Liability{
		{Name: "Liabilities:CreditCard:Chase"},
	}
	out := buildAmortizations(liabs)
	assert.Len(t, out, 0)
}

// TestBuildAmortizationsReconcilesExactly: even when the YAML→float64 path
// would normally introduce IEEE-754 noise (e.g. 4.9 is not exact in binary),
// the amortization output must still reconcile exactly: sum(principal) ==
// declared principal, and per-row payment == principal + interest.
func TestBuildAmortizationsReconcilesExactly(t *testing.T) {
	liabs := []config.Liability{
		{
			Name:       "Liabilities:Mortgage:House",
			Kind:       config.AmortizingLoan,
			Principal:  1100000, // exactly representable, but goes through float64
			Rate:       4.9,     // NOT exactly representable in IEEE-754
			TermMonths: 240,
			Schedule:   config.LiabilityEqualPayment,
		},
	}
	out := buildAmortizations(liabs)
	require.Len(t, out, 1)
	a := out[0]

	// Per-row invariant.
	for i, m := range a.Months {
		require.True(t, m.Principal.Add(m.Interest).Equal(m.Payment),
			"row %d: payment=%s != principal=%s + interest=%s",
			i+1, m.Payment, m.Principal, m.Interest)
	}

	// Total principal reconciles to declared principal exactly.
	sumP := decimal.Zero
	for _, m := range a.Months {
		sumP = sumP.Add(m.Principal)
	}
	want := decimal.NewFromInt(1100000)
	require.True(t, sumP.Equal(want),
		"sum(principal)=%s, want %s", sumP, want)

	// Final balance is zero.
	require.True(t, a.Months[len(a.Months)-1].Balance.IsZero(),
		"last balance must be zero, got %s", a.Months[len(a.Months)-1].Balance)
}

// TestDecimalFromYAMLFloat confirms the float→decimal helper preserves the
// shortest decimal representation rather than the IEEE-754 expansion.
func TestDecimalFromYAMLFloat(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{4.9, "4.9"},
		{1100000, "1100000"},
		{0.1, "0.1"},
		{1.5, "1.5"},
		{0, "0"},
		{0.0001, "0.0001"},
	}
	for _, tc := range cases {
		got := decimalFromYAMLFloat(tc.in).String()
		assert.Equal(t, tc.want, got, "decimalFromYAMLFloat(%v)", tc.in)
	}
}
