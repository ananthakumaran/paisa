package server

import (
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PostingPair is a (purchase, sell) lot match-up produced by the
// FIFO walk in computeCapitalGains.
//
// After the M2-A delete of the India-specific taxation engine this
// struct no longer carries a Tax field — capital gains are reported
// in pure decimal currency terms; jurisdiction-specific tax rules
// (CN / US / etc.) will be reintroduced in M4 with a different shape.
type PostingPair struct {
	Purchase posting.Posting `json:"purchase"`
	Sell     posting.Posting `json:"sell"`
}

// YearCapitalGain holds aggregated capital-gain figures for a single
// calendar year. (Previously keyed by Indian fiscal year as
// FYCapitalGain — renamed for issue #5.)
type YearCapitalGain struct {
	Units         decimal.Decimal `json:"units"`
	PurchasePrice decimal.Decimal `json:"purchase_price"`
	SellPrice     decimal.Decimal `json:"sell_price"`
	PostingPairs  []PostingPair   `json:"posting_pairs"`
}

type CapitalGain struct {
	Account     string                     `json:"account"`
	TaxCategory string                     `json:"tax_category"`
	Year        map[string]YearCapitalGain `json:"year"`
}

// GetCapitalGains is a no-op stub after M2-A. The India-specific
// taxation engine (LTCG / STCG / CII indexation / Schedule AL) was
// removed; until M4 rebuilds capital-gains for the CN tax regime,
// this endpoint returns an empty map so the frontend can still call
// it without crashing.
func GetCapitalGains(db *gorm.DB) gin.H {
	return gin.H{"capital_gains": map[string]CapitalGain{}}
}

// computeCapitalGains walks a single account's postings FIFO and
// aggregates buy/sell pairs by the sell date's calendar year. It is
// retained (separately from GetCapitalGains) so that the unit tests
// covering the FY → calendar-year migration (issue #5) continue to
// pin the output shape. The full server endpoint stays a stub until
// M4 rebuilds capital gains for the CN tax regime.
func computeCapitalGains(_ *gorm.DB, account string, taxCategory string, postings []posting.Posting) CapitalGain {
	capitalGain := CapitalGain{Account: account, TaxCategory: taxCategory, Year: make(map[string]YearCapitalGain)}
	var available []posting.Posting
	for _, p := range postings {
		if p.Quantity.GreaterThan(decimal.Zero) {
			available = append(available, p)
		} else {
			quantity := p.Quantity.Neg()
			purchasePrice := decimal.Zero
			postingPairs := make([]PostingPair, 0)
			for quantity.GreaterThan(decimal.Zero) && len(available) > 0 {
				first := available[0]
				q := decimal.Zero

				if first.Quantity.GreaterThan(quantity) {
					first.AddQuantity(quantity.Neg())
					q = quantity
					available[0] = first
					quantity = decimal.Zero
				} else {
					quantity = quantity.Sub(first.Quantity)
					q = first.Quantity
					available = available[1:]
				}

				purchasePrice = purchasePrice.Add(q.Mul(first.Price()))
				postingPair := PostingPair{Purchase: first.WithQuantity(q), Sell: p.WithQuantity(q.Neg())}
				postingPairs = append(postingPairs, postingPair)

			}
			year := utils.CalendarYear(p.Date)
			yearGain := capitalGain.Year[year]
			yearGain.Units = yearGain.Units.Add(p.Quantity.Neg())
			yearGain.PurchasePrice = yearGain.PurchasePrice.Add(purchasePrice)
			yearGain.SellPrice = yearGain.SellPrice.Add(p.Amount.Neg())
			yearGain.PostingPairs = append(yearGain.PostingPairs, postingPairs...)

			capitalGain.Year[year] = yearGain

		}
	}

	return capitalGain
}
