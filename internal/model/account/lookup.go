package account

import "strings"

// Account is a minimal record used by GetKind to resolve an account's kind.
// It mirrors the fields of internal/config.Account that are relevant to kind
// resolution; callers convert their richer config struct down to this shape
// to avoid an import cycle (internal/config already depends on this package).
type Account struct {
	Name string
	Kind AccountKind
}

// GetKind resolves the AccountKind for a given ledger account name.
//
// Resolution order:
//  1. Exact match in the supplied `accounts` list with a non-empty `Kind`.
//     This is the user-configured override coming from `paisa.yaml`.
//  2. Path-prefix fallback rules (see below).
//  3. KindUnknown if nothing matches.
//
// Path-prefix fallback rules (applied only when no explicit kind exists):
//
//	Assets:Saving:*        -> bank_current      (default; users override per-account)
//	Assets:Property:House* -> real_estate
//	Assets:Property:Car*   -> vehicle
//	Assets:Property:Stock* -> stock
//	Assets:Brokerage:*     -> mutual_fund       (ambiguous default; user should override)
//	Assets:Bridge:*        -> bank_current      (transit account)
//	Assets:PPA:*:Fund      -> tax_deferred_fund
//	Assets:PPA:*           -> tax_deferred_cash (anything else under PPA)
//	Assets:Loans*          -> receivable
//	Assets:YangLiu:*       -> receivable        (个人持有的应收款,看起来是其他人名 prefix)
//	Liabilities:*          -> KindUnknown       (kind is for Assets only)
//	(everything else)      -> KindUnknown
//
// Matching is case-sensitive and operates on the ledger account name exactly
// as it appears in the journal (e.g. "Assets:Saving:CMB").
func GetKind(name string, accounts []Account) AccountKind {
	// 1. Explicit override
	for _, a := range accounts {
		if a.Name == name && a.Kind != "" {
			return a.Kind
		}
	}

	// 2. Fallback by path prefix
	switch {
	case strings.HasPrefix(name, "Liabilities:") || name == "Liabilities":
		return KindUnknown

	case strings.HasPrefix(name, "Assets:Saving:") || name == "Assets:Saving":
		return BankCurrent

	case strings.HasPrefix(name, "Assets:Property:House"):
		return RealEstate
	case strings.HasPrefix(name, "Assets:Property:Car"):
		return Vehicle
	case strings.HasPrefix(name, "Assets:Property:Stock"):
		return Stock

	case strings.HasPrefix(name, "Assets:Brokerage:") || name == "Assets:Brokerage":
		return MutualFund

	case strings.HasPrefix(name, "Assets:Bridge:") || name == "Assets:Bridge":
		return BankCurrent

	case strings.HasPrefix(name, "Assets:PPA:") || name == "Assets:PPA":
		if strings.HasSuffix(name, ":Fund") || strings.HasSuffix(name, "Fund") {
			return TaxDeferredFund
		}
		return TaxDeferredCash

	case strings.HasPrefix(name, "Assets:Loans"):
		return Receivable

	case strings.HasPrefix(name, "Assets:YangLiu:") || name == "Assets:YangLiu":
		return Receivable
	}

	return KindUnknown
}
