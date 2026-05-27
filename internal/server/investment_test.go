package server

import (
	"testing"

	"github.com/ananthakumaran/paisa/internal/model/account"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func makeInvestPosting(account string, amount int64) posting.Posting {
	return posting.Posting{
		Account:   account,
		Amount:    decimal.NewFromInt(amount),
		Quantity:  decimal.NewFromInt(amount),
		Commodity: "CNY",
	}
}

// TestFilterOutNonInvestmentKinds_DropsBankCurrent — an account configured
// as `kind: bank_current` must be dropped from the Investment Timeline.
// The second posting (`Assets:Equity:VOO`) has no config and no matching
// path-prefix rule, so it resolves to KindUnknown and is also dropped;
// the assertion below specifically guards the BankCurrent → drop path,
// which is the regression the issue describes.
func TestFilterOutNonInvestmentKinds_DropsBankCurrent(t *testing.T) {
	accounts := []account.Account{
		{Name: "Assets:Saving:CMB", Kind: account.BankCurrent},
	}
	postings := []posting.Posting{
		makeInvestPosting("Assets:Saving:CMB", 1000), // bank_current → dropped
		makeInvestPosting("Assets:Equity:VOO", 5000), // unknown → dropped
	}

	out := filterOutNonInvestmentKinds(postings, accounts)

	for _, p := range out {
		assert.NotEqual(t, "Assets:Saving:CMB", p.Account,
			"bank_current account leaked into investment timeline")
	}
}

// TestFilterOutNonInvestmentKinds_KeepsInvestmentKinds — when an account is
// explicitly classified as mutual_fund / stock / bond / etc., it must
// remain in the investment timeline regardless of its ledger path.
func TestFilterOutNonInvestmentKinds_KeepsInvestmentKinds(t *testing.T) {
	accounts := []account.Account{
		{Name: "Assets:Saving:CMB:Fund", Kind: account.MutualFund},
		{Name: "Assets:Property:Stock:AAPL", Kind: account.Stock},
		{Name: "Assets:Bond:Reverse", Kind: account.Bond},
	}
	postings := []posting.Posting{
		makeInvestPosting("Assets:Saving:CMB:Fund", 2000),
		makeInvestPosting("Assets:Property:Stock:AAPL", 3000),
		makeInvestPosting("Assets:Bond:Reverse", 1000),
	}

	out := filterOutNonInvestmentKinds(postings, accounts)
	assert.Len(t, out, 3, "all three investment-kind postings should remain")
}

// TestFilterOutNonInvestmentKinds_DropsRealEstateAndVehicle — buying a house
// or car shouldn't count as an "investment" flow.
func TestFilterOutNonInvestmentKinds_DropsRealEstateAndVehicle(t *testing.T) {
	accounts := []account.Account{
		{Name: "Assets:Property:House:LCHWK1154", Kind: account.RealEstate},
		{Name: "Assets:Property:Car:Tesla", Kind: account.Vehicle},
	}
	postings := []posting.Posting{
		makeInvestPosting("Assets:Property:House:LCHWK1154", 1_000_000),
		makeInvestPosting("Assets:Property:Car:Tesla", 200_000),
		makeInvestPosting("Assets:Brokerage:Vanguard", 5_000),
	}

	out := filterOutNonInvestmentKinds(postings, accounts)
	for _, p := range out {
		assert.NotContains(t, p.Account, "House", "house leaked")
		assert.NotContains(t, p.Account, "Car", "car leaked")
	}
	// Brokerage (mutual_fund fallback) is kept.
	assert.Len(t, out, 1)
	assert.Equal(t, "Assets:Brokerage:Vanguard", out[0].Account)
}

// TestFilterOutNonInvestmentKinds_FallsBackToPrefix — accounts not present
// in `accounts` config use the path-prefix fallback (M1-D's GetKind).
func TestFilterOutNonInvestmentKinds_FallsBackToPrefix(t *testing.T) {
	// No explicit config — rely entirely on path prefix.
	postings := []posting.Posting{
		makeInvestPosting("Assets:Saving:CMB", 1000),         // bank_current — drop
		makeInvestPosting("Assets:Brokerage:Vanguard", 5000), // mutual_fund — keep
		makeInvestPosting("Assets:Property:Car:Tesla", 200),  // vehicle — drop
	}
	out := filterOutNonInvestmentKinds(postings, nil)

	assert.Len(t, out, 1)
	assert.Equal(t, "Assets:Brokerage:Vanguard", out[0].Account)
}

// TestIsInvestmentKind_Whitelist documents which kinds count as
// "investment" flows for the Investment Timeline.
func TestIsInvestmentKind_Whitelist(t *testing.T) {
	investmentKinds := []account.AccountKind{
		account.MutualFund,
		account.Stock,
		account.Bond,
		account.StructuredDeposit,
		account.TaxDeferredFund,
		account.Crypto,
	}
	for _, k := range investmentKinds {
		assert.True(t, isInvestmentKind(k), "%s should count as investment", k)
	}

	nonInvestmentKinds := []account.AccountKind{
		account.BankCurrent,
		account.BankSavings,
		account.CashEquivalent,
		account.RealEstate,
		account.Vehicle,
		account.HousingFund,
		account.Receivable,
		account.ForeignCurrency,
		account.KindUnknown,
		account.TaxDeferredCash,
	}
	for _, k := range nonInvestmentKinds {
		assert.False(t, isInvestmentKind(k), "%s should NOT count as investment", k)
	}
}
