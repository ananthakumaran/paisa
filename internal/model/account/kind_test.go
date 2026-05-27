package account

import "testing"

func TestAccountKindString(t *testing.T) {
	cases := []struct {
		kind AccountKind
		want string
	}{
		{BankCurrent, "bank_current"},
		{BankSavings, "bank_savings"},
		{CashEquivalent, "cash_equivalent"},
		{MutualFund, "mutual_fund"},
		{Stock, "stock"},
		{Bond, "bond"},
		{StructuredDeposit, "structured_deposit"},
		{RealEstate, "real_estate"},
		{Vehicle, "vehicle"},
		{TaxDeferredFund, "tax_deferred_fund"},
		{TaxDeferredCash, "tax_deferred_cash"},
		{HousingFund, "housing_fund"},
		{Receivable, "receivable"},
		{Crypto, "crypto"},
		{ForeignCurrency, "foreign_currency"},
		{KindUnknown, "unknown"},
	}

	for _, c := range cases {
		if string(c.kind) != c.want {
			t.Errorf("AccountKind(%q) string = %q; want %q", c.kind, string(c.kind), c.want)
		}
	}
}

func TestAccountKindIsValid(t *testing.T) {
	validValues := []string{
		"bank_current",
		"bank_savings",
		"cash_equivalent",
		"mutual_fund",
		"stock",
		"bond",
		"structured_deposit",
		"real_estate",
		"vehicle",
		"tax_deferred_fund",
		"tax_deferred_cash",
		"housing_fund",
		"receivable",
		"crypto",
		"foreign_currency",
		"unknown",
	}

	for _, v := range validValues {
		if !AccountKind(v).IsValid() {
			t.Errorf("AccountKind(%q).IsValid() = false; want true", v)
		}
	}

	invalid := []string{
		"",
		"BANK_CURRENT",
		"savings",
		"foo",
		"bank-current",
	}
	for _, v := range invalid {
		if AccountKind(v).IsValid() {
			t.Errorf("AccountKind(%q).IsValid() = true; want false", v)
		}
	}
}

func TestAllAccountKinds(t *testing.T) {
	all := AllAccountKinds()
	if len(all) != 16 {
		t.Fatalf("AllAccountKinds length = %d; want 16", len(all))
	}
	// Each must be valid.
	seen := make(map[AccountKind]bool)
	for _, k := range all {
		if !k.IsValid() {
			t.Errorf("AllAccountKinds returned invalid kind %q", k)
		}
		if seen[k] {
			t.Errorf("AllAccountKinds returned duplicate %q", k)
		}
		seen[k] = true
	}
}
