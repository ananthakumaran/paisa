package account

import "testing"

func TestGetKindFallback(t *testing.T) {
	cases := []struct {
		name string
		want AccountKind
	}{
		// Assets:Saving:* -> bank_current
		{"Assets:Saving:CMB", BankCurrent},
		{"Assets:Saving:CMB:Sub", BankCurrent},

		// Assets:Property:House* -> real_estate
		{"Assets:Property:House", RealEstate},
		{"Assets:Property:House:LCHWK1154", RealEstate},

		// Assets:Property:Car* -> vehicle
		{"Assets:Property:Car", Vehicle},
		{"Assets:Property:Car:Tesla", Vehicle},

		// Assets:Property:Stock* -> stock
		{"Assets:Property:Stock", Stock},
		{"Assets:Property:Stock:AAPL", Stock},

		// Assets:Brokerage:* -> mutual_fund (default ambiguous)
		{"Assets:Brokerage:Vanguard", MutualFund},

		// Assets:Bridge:* -> bank_current (transit)
		{"Assets:Bridge:Pending", BankCurrent},

		// Assets:PPA:* -> tax_deferred_fund (Fund suffix) / tax_deferred_cash
		{"Assets:PPA:CMB:Fund", TaxDeferredFund},
		{"Assets:PPA:CMB:Cash", TaxDeferredCash},
		{"Assets:PPA:CMB", TaxDeferredCash},

		// Assets:Loans* -> receivable
		{"Assets:Loans", Receivable},
		{"Assets:Loans:Alice", Receivable},

		// Assets:YangLiu:* -> receivable
		{"Assets:YangLiu:Foo", Receivable},

		// Liabilities -> unknown (no kind for liabilities)
		{"Liabilities:CreditCard:Chase", KindUnknown},

		// Truly unmatched
		{"Expenses:Food", KindUnknown},
		{"Assets:Unknown:Foo", KindUnknown},
	}

	for _, c := range cases {
		got := GetKind(c.name, nil)
		if got != c.want {
			t.Errorf("GetKind(%q, nil) = %q; want %q", c.name, got, c.want)
		}
	}
}

func TestGetKindOverride(t *testing.T) {
	accounts := []Account{
		{Name: "Assets:Saving:CMB:Fund", Kind: MutualFund},
		{Name: "Assets:Property:House:LCHWK1154", Kind: RealEstate},
		// Empty Kind should be ignored (fall through to prefix rules).
		{Name: "Assets:Saving:CMB", Kind: ""},
	}

	// Explicit override beats fallback.
	if got := GetKind("Assets:Saving:CMB:Fund", accounts); got != MutualFund {
		t.Errorf("override: GetKind(\"Assets:Saving:CMB:Fund\") = %q; want %q", got, MutualFund)
	}

	// Override that matches fallback (real_estate would also match): still respected.
	if got := GetKind("Assets:Property:House:LCHWK1154", accounts); got != RealEstate {
		t.Errorf("override: GetKind(\"Assets:Property:House:LCHWK1154\") = %q; want %q", got, RealEstate)
	}

	// Empty Kind in config -> fall back to prefix rule.
	if got := GetKind("Assets:Saving:CMB", accounts); got != BankCurrent {
		t.Errorf("empty kind in config should fall through: got %q; want %q", got, BankCurrent)
	}

	// Account not in config -> fallback.
	if got := GetKind("Assets:Property:Car:Tesla", accounts); got != Vehicle {
		t.Errorf("uncovered account falls back: got %q; want %q", got, Vehicle)
	}
}
