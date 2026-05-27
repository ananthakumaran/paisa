package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// After the fiscal-year → calendar-year migration (issue #5), the JSON
// field that used to be `"fy"` must serialise as `"year"` and the
// per-year keys must be plain 4-digit strings like "2024" — never
// Indian fiscal-year ranges like "2024-25". These tests pin both
// shapes; the frontend (`cg.year[<key>]`) depends on them.

// makePosting returns a posting that purchases `qty` units of `commodity`
// at the given unit `price` on `date`. It mirrors what query.Init().All()
// would produce for a real ledger entry.
func makePosting(date string, account string, commodity string, qty float64, price float64) posting.Posting {
	d, _ := time.Parse("2006/01/02", date)
	q := decimal.NewFromFloat(qty)
	amount := decimal.NewFromFloat(qty * price)
	return posting.Posting{
		Date:      d,
		Account:   account,
		Commodity: commodity,
		Quantity:  q,
		Amount:    amount,
	}
}

// When a posting set has no tax classification (untagged境外 holdings
// such as USD/HKD cash or UBER/0700.HK shares), computeCapitalGains
// must still produce a well-formed CapitalGain struct: TaxCategory is
// the empty string, Year map is non-nil, and the resulting JSON has no
// nil collections (so the frontend does not crash when it iterates
// over them).
//
// See issue #1.
func TestComputeCapitalGains_EmptyTaxCategoryProducesWellFormedOutput(t *testing.T) {
	// One buy on 2023-01-01, one sell on 2024-01-01 — even with no tax
	// classification, the handler must not panic and must return a
	// JSON-serialisable structure with a non-nil year map.
	postings := []posting.Posting{
		makePosting("2023/01/01", "Assets:Brokerage:IBKR", "UBER", 3, 128.02),
		makePosting("2024/01/01", "Assets:Brokerage:IBKR", "UBER", -3, 150.00),
	}

	gain := computeCapitalGains(nil, "Assets:Brokerage:IBKR", "", postings)

	assert.Equal(t, "Assets:Brokerage:IBKR", gain.Account)
	assert.Equal(t, "", gain.TaxCategory)
	assert.NotNil(t, gain.Year, "Year map must be non-nil so json encodes as {} not null")

	// The JSON encoding must succeed and must not contain `null` for the
	// year field (frontend uses cg.year[<calendar-year>] which requires
	// an object, not null).
	b, err := json.Marshal(gain)
	assert.NoError(t, err)
	var decoded map[string]interface{}
	assert.NoError(t, json.Unmarshal(b, &decoded))
	assert.NotNil(t, decoded["year"], "year must serialise to an object, not null")
	_, fyPresent := decoded["fy"]
	assert.False(t, fyPresent, "legacy 'fy' field must NOT be present after the calendar-year migration")

	// Calendar-year keys are plain 4-digit strings like "2024",
	// never Indian fiscal-year ranges like "2023-24".
	for k := range gain.Year {
		assert.NotContains(t, k, "-", "calendar-year key must not contain a dash")
		assert.Len(t, k, 4, "calendar-year key must be 4 digits")
	}
}

// When computeCapitalGains is fed postings whose commodity lacks a
// tax_category, the resulting map must be well-formed JSON. We test
// this at the unit level on the helper because spinning a full gin
// router needs a working sqlite + ledger CLI.
func TestComputeCapitalGains_NoSellsProducesEmptyYear(t *testing.T) {
	postings := []posting.Posting{
		makePosting("2023/01/01", "Assets:Brokerage:IBKR", "HKD", 1000, 1.0),
	}

	gain := computeCapitalGains(nil, "Assets:Brokerage:IBKR", "", postings)
	assert.Equal(t, "", gain.TaxCategory)
	assert.NotNil(t, gain.Year)
	assert.Empty(t, gain.Year)
}

// After the FY → calendar-year migration, a buy/sell pair that
// straddles a calendar-year boundary in the Indian fiscal-year sense
// (buy in Apr, sell in Feb of the next calendar year) must aggregate
// under the SELL date's calendar year — not the FY range.
func TestComputeCapitalGains_AggregatesUnderCalendarYearOfSellDate(t *testing.T) {
	postings := []posting.Posting{
		// Buy in April 2023 (Indian FY 2023-24).
		makePosting("2023/04/15", "Assets:Brokerage:IBKR", "UBER", 5, 100.0),
		// Sell in February 2024 — still Indian FY 2023-24, but
		// calendar year 2024.
		makePosting("2024/02/15", "Assets:Brokerage:IBKR", "UBER", -5, 150.0),
	}

	gain := computeCapitalGains(nil, "Assets:Brokerage:IBKR", "", postings)
	_, has2024 := gain.Year["2024"]
	assert.True(t, has2024, "expected key '2024' (calendar year of sell), got %v", keysOf(gain.Year))
	for k := range gain.Year {
		assert.NotContains(t, k, "-", "no FY-range keys allowed after migration")
	}
}

func keysOf[V any](m map[string]V) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
