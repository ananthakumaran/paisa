package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

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

// When the commodity has no tax_category (zero-value Commodity, e.g.
// untagged境外 holdings such as USD/HKD cash or UBER/0700.HK shares),
// computeCapitalGains must still produce a well-formed CapitalGain
// struct: TaxCategory is the empty string, FY map is non-nil, and the
// resulting JSON has no nil collections (so the frontend does not crash
// when it iterates over them).
//
// See issue #1.
func TestComputeCapitalGains_EmptyTaxCategoryProducesWellFormedOutput(t *testing.T) {
	// zero-value commodity → TaxCategory is "" (empty)
	untagged := config.Commodity{Name: "UBER", Type: config.Stock}

	// One buy on 2023-01-01, one sell on 2024-01-01 — even with no tax
	// classification, the handler must not panic and must return a
	// JSON-serialisable structure with a non-nil fy map.
	postings := []posting.Posting{
		makePosting("2023/01/01", "Assets:Brokerage:IBKR", "UBER", 3, 128.02),
		makePosting("2024/01/01", "Assets:Brokerage:IBKR", "UBER", -3, 150.00),
	}

	// Don't actually need the db for these untagged postings — the tax
	// engine will produce zeros — but the function takes one regardless.
	gain := computeCapitalGains(nil, "Assets:Brokerage:IBKR", untagged, postings)

	assert.Equal(t, "Assets:Brokerage:IBKR", gain.Account)
	assert.Equal(t, "", gain.TaxCategory)
	assert.NotNil(t, gain.FY, "FY map must be non-nil so json encodes as {} not null")

	// The JSON encoding must succeed and must not contain `null` for the
	// fy field (frontend uses cg.fy[financialYear] which requires an
	// object, not null).
	b, err := json.Marshal(gain)
	assert.NoError(t, err)
	var decoded map[string]interface{}
	assert.NoError(t, json.Unmarshal(b, &decoded))
	assert.NotNil(t, decoded["fy"], "fy must serialise to an object, not null")
}

// When GetCapitalGains is fed postings whose commodity lacks a
// tax_category, the resulting map must be well-formed JSON. We test
// this at the unit level on the helper because spinning a full gin
// router needs a working sqlite + ledger CLI.
func TestComputeCapitalGains_NoSellsProducesEmptyFY(t *testing.T) {
	untagged := config.Commodity{Name: "HKD", Type: config.Stock}
	postings := []posting.Posting{
		makePosting("2023/01/01", "Assets:Brokerage:IBKR", "HKD", 1000, 1.0),
	}

	gain := computeCapitalGains(nil, "Assets:Brokerage:IBKR", untagged, postings)
	assert.Equal(t, "", gain.TaxCategory)
	assert.NotNil(t, gain.FY)
	assert.Empty(t, gain.FY)
}
