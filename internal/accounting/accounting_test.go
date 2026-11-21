package accounting

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func lot(date string, quantity, amount int64) posting.Posting {
	d, _ := time.Parse("2006-01-02", date)
	return posting.Posting{Date: d, Commodity: "ABC", Quantity: decimal.NewFromInt(quantity), Amount: decimal.NewFromInt(amount)}
}

func TestApplySplit(t *testing.T) {
	available := []posting.Posting{lot("2021-02-01", 10, 1000), lot("2021-08-01", 10, 2000)}

	// 1:2 split booked as a sell and buy
	adjusted, ok := applySplit(available, []posting.Posting{lot("2022-01-01", -20, -3000), lot("2022-01-01", 40, 3000)})
	assert.True(t, ok)
	assert.Equal(t, "20", adjusted[0].Quantity.String())
	assert.Equal(t, "1000", adjusted[0].Amount.String())
	assert.Equal(t, "2021-02-01", adjusted[0].Date.Format("2006-01-02"))
	assert.Equal(t, "20", adjusted[1].Quantity.String())
	assert.Equal(t, "2000", adjusted[1].Amount.String())
	assert.Equal(t, "10", available[0].Quantity.String(), "should not modify the original lots")

	// 1:3 split booked as just the additional units
	adjusted, ok = applySplit(available, []posting.Posting{lot("2022-01-01", 40, 0)})
	assert.True(t, ok)
	assert.Equal(t, "30", adjusted[0].Quantity.String())
	assert.Equal(t, "30", adjusted[1].Quantity.String())

	// uneven ratio should still add up to the holding after the split
	adjusted, ok = applySplit([]posting.Posting{lot("2021-02-01", 1, 100), lot("2021-03-01", 1, 100), lot("2021-04-01", 1, 100)}, []posting.Posting{lot("2022-01-01", -2, 0)})
	assert.True(t, ok)
	assert.Equal(t, "1", adjusted[0].Quantity.Add(adjusted[1].Quantity).Add(adjusted[2].Quantity).String())

	_, ok = applySplit([]posting.Posting{}, []posting.Posting{lot("2022-01-01", 40, 0)})
	assert.False(t, ok)
}

func TestFIFOApplyConsumed(t *testing.T) {
	available := []posting.Posting{lot("2021-02-01", 10, 1000), lot("2021-08-01", 10, 2000)}
	available, consumed := fifoApply(available, lot("2022-01-01", -15, -3000))
	assert.Len(t, consumed, 2)
	assert.Equal(t, "10", consumed[0].Quantity.String())
	assert.Equal(t, "1000", consumed[0].Amount.String())
	assert.Equal(t, "5", consumed[1].Quantity.String())
	assert.Equal(t, "1000", consumed[1].Amount.String())
	assert.Len(t, available, 1)
	assert.Equal(t, "5", available[0].Quantity.String())
	assert.Equal(t, "1000", available[0].Amount.String())
}
