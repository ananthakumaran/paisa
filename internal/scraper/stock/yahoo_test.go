package stock

import (
	"testing"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeYahooPrice(t *testing.T) {
	tests := []struct {
		name             string
		inputValue       float64
		inputCurrency    string
		expectedValue    float64
		expectedCurrency string
	}{
		{"pounds", 140.14, "GBP", 140.14, "GBP"},
		{"pence", 690.7, "GBp", 6.907, "GBP"},
		{"pence alternate code", 21.5, "GBX", 0.215, "GBP"},
		{"other currency", 10.5, "USD", 10.5, "USD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, currency := normalizeYahooPrice(tt.inputValue, tt.inputCurrency)

			assert.InDelta(t, tt.expectedValue, value, 0.000001)
			assert.Equal(t, tt.expectedCurrency, currency)
		})
	}
}

func TestExchangeRateAt(t *testing.T) {
	exchangePrice := btree.New(2)
	exchangePrice.ReplaceOrInsert(ExchangePrice{Timestamp: 100, Close: 1.1})
	exchangePrice.ReplaceOrInsert(ExchangePrice{Timestamp: 200, Close: 1.2})

	rate, err := exchangeRateAt(exchangePrice, 250)

	assert.Nil(t, err)
	assert.InDelta(t, 1.2, rate, 0.000001)
}

func TestExchangeRateAtReturnsErrorWhenMissing(t *testing.T) {
	exchangePrice := btree.New(2)
	exchangePrice.ReplaceOrInsert(ExchangePrice{Timestamp: 200, Close: 1.2})

	_, err := exchangeRateAt(exchangePrice, 100)
	assert.Error(t, err)

	_, err = exchangeRateAt(btree.New(2), 100)
	assert.Error(t, err)
}

func TestExchangeRateAtReturnsErrorForZeroRate(t *testing.T) {
	exchangePrice := btree.New(2)
	exchangePrice.ReplaceOrInsert(ExchangePrice{Timestamp: 100, Close: 0})

	_, err := exchangeRateAt(exchangePrice, 100)

	assert.Error(t, err)
}
