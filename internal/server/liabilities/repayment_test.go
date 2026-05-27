package liabilities

import (
	"testing"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/stretchr/testify/assert"
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
