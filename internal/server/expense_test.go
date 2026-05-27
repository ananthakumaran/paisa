package server

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// makeExpensePosting builds a minimal expense posting fixture. In paisa's
// convention an Expenses:* posting is positive when it is a real expense
// and negative when it is a refund (红冲/退款) that cancels part of a prior
// expense. This is the M0-C / M3-G refund discipline.
func makeExpensePosting(account string, amount float64, date time.Time) posting.Posting {
	return posting.Posting{
		Account:      account,
		Amount:       decimal.NewFromFloat(amount),
		MarketAmount: decimal.NewFromFloat(amount),
		Quantity:     decimal.NewFromFloat(amount),
		Commodity:    "CNY",
		Date:         date,
	}
}

// TestComputeExpenseSummary_RefundReducesNet locks in the issue #25
// invariant: when a month has a +30000 Expenses:Train purchase and a
// -2990 Expenses:Train refund (style A), the per-month summary must
// expose gross=30000, refunds=-2990, net=27010 so the UI toggle has
// real numbers to switch between.
func TestComputeExpenseSummary_RefundReducesNet(t *testing.T) {
	d := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	postings := []posting.Posting{
		makeExpensePosting("Expenses:Transport:Train", 30000, d),
		makeExpensePosting("Expenses:Transport:Train", -2990, d.AddDate(0, 0, 3)),
	}

	summaries := computeExpenseSummary(postings, "2006-01")

	got, ok := summaries["2024-06"]
	assert.True(t, ok, "expected 2024-06 summary, keys=%v", summaries)
	assert.True(t, got.Gross.Equal(decimal.NewFromInt(30000)),
		"gross: got %s want 30000", got.Gross.String())
	assert.True(t, got.Refunds.Equal(decimal.NewFromInt(-2990)),
		"refunds: got %s want -2990", got.Refunds.String())
	assert.True(t, got.Net.Equal(decimal.NewFromInt(27010)),
		"net: got %s want 27010", got.Net.String())
}

// TestComputeExpenseSummary_NoRefund — a month with only forward
// expenses has refunds=0 and gross == net. This is the common path
// and must not regress.
func TestComputeExpenseSummary_NoRefund(t *testing.T) {
	d := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	postings := []posting.Posting{
		makeExpensePosting("Expenses:Food", 100, d),
		makeExpensePosting("Expenses:Food", 250, d.AddDate(0, 0, 5)),
	}

	summaries := computeExpenseSummary(postings, "2006-01")

	got := summaries["2024-07"]
	assert.True(t, got.Gross.Equal(decimal.NewFromInt(350)))
	assert.True(t, got.Refunds.IsZero(), "no refund → refunds must be zero, got %s", got.Refunds.String())
	assert.True(t, got.Net.Equal(decimal.NewFromInt(350)))
}

// TestComputeExpenseSummary_YearlyKey verifies the helper honors the
// supplied date layout so it works for both month_wise (2006-01) and
// year_wise (2006) aggregation.
func TestComputeExpenseSummary_YearlyKey(t *testing.T) {
	postings := []posting.Posting{
		makeExpensePosting("Expenses:Travel", 5000, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)),
		makeExpensePosting("Expenses:Travel", -500, time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC)),
		makeExpensePosting("Expenses:Travel", 1000, time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)),
	}

	summaries := computeExpenseSummary(postings, "2006")

	y2024 := summaries["2024"]
	assert.True(t, y2024.Gross.Equal(decimal.NewFromInt(5000)))
	assert.True(t, y2024.Refunds.Equal(decimal.NewFromInt(-500)))
	assert.True(t, y2024.Net.Equal(decimal.NewFromInt(4500)))

	y2023 := summaries["2023"]
	assert.True(t, y2023.Gross.Equal(decimal.NewFromInt(1000)))
	assert.True(t, y2023.Refunds.IsZero())
	assert.True(t, y2023.Net.Equal(decimal.NewFromInt(1000)))
}
