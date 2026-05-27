// Package account defines the AccountKind enum used to classify user accounts
// (banks, mutual funds, real estate, etc.) for downstream features such as
// allocation aggregation (M1-E) and receivables tracking (M2-G).
//
// The enum values are mirrored in the TypeScript frontend at
// src/lib/account_kind.ts and must stay in lockstep.
package account

// AccountKind is a string-typed enum classifying an asset account.
//
// Keep the constants below in sync with:
//   - src/lib/account_kind.ts (TypeScript mirror)
//   - internal/config/schema.json `accounts[].kind` enum
type AccountKind string

const (
	BankCurrent       AccountKind = "bank_current"       // 活期
	BankSavings       AccountKind = "bank_savings"       // 定期/通知存款
	CashEquivalent    AccountKind = "cash_equivalent"    // 货币基金/朝朝宝/余额宝
	MutualFund        AccountKind = "mutual_fund"        // 开放式基金
	Stock             AccountKind = "stock"              // 股票
	Bond              AccountKind = "bond"               // 债券/国债逆回购
	StructuredDeposit AccountKind = "structured_deposit" // 结构性存款/银行理财
	RealEstate        AccountKind = "real_estate"        // 不动产
	Vehicle           AccountKind = "vehicle"            // 车辆
	TaxDeferredFund   AccountKind = "tax_deferred_fund"  // PPA 个人养老金基金部分
	TaxDeferredCash   AccountKind = "tax_deferred_cash"  // PPA 现金部分
	HousingFund       AccountKind = "housing_fund"       // 公积金
	Receivable        AccountKind = "receivable"         // 应收款
	Crypto            AccountKind = "crypto"             // 加密货币
	ForeignCurrency   AccountKind = "foreign_currency"   // 外币现钞/活期
	KindUnknown       AccountKind = "unknown"            // 未知/未分类
)

// AllAccountKinds returns every defined AccountKind in declaration order.
//
// Used by tests and (later) by API surface helpers; the slice is freshly
// allocated each call so callers may mutate it safely.
func AllAccountKinds() []AccountKind {
	return []AccountKind{
		BankCurrent,
		BankSavings,
		CashEquivalent,
		MutualFund,
		Stock,
		Bond,
		StructuredDeposit,
		RealEstate,
		Vehicle,
		TaxDeferredFund,
		TaxDeferredCash,
		HousingFund,
		Receivable,
		Crypto,
		ForeignCurrency,
		KindUnknown,
	}
}

// IsValid reports whether the receiver is one of the defined AccountKind
// constants. The empty string is not valid; callers that want to treat an
// unset Kind as "fall through to defaults" should check that separately.
func (k AccountKind) IsValid() bool {
	switch k {
	case BankCurrent, BankSavings, CashEquivalent, MutualFund, Stock, Bond,
		StructuredDeposit, RealEstate, Vehicle, TaxDeferredFund, TaxDeferredCash,
		HousingFund, Receivable, Crypto, ForeignCurrency, KindUnknown:
		return true
	}
	return false
}
