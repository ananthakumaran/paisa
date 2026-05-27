// AccountKind enum mirroring internal/model/account/kind.go.
//
// Keep this file in sync with:
//   - internal/model/account/kind.go (Go source of truth)
//   - internal/config/schema.json (`accounts[].kind` enum)
//
// Display labels are provided in zh-CN (default) and en for now. Add more
// locales as needed; the keys must cover every member of ACCOUNT_KIND_VALUES.

export const AccountKind = {
  BankCurrent: "bank_current",
  BankSavings: "bank_savings",
  CashEquivalent: "cash_equivalent",
  MutualFund: "mutual_fund",
  Stock: "stock",
  Bond: "bond",
  StructuredDeposit: "structured_deposit",
  RealEstate: "real_estate",
  Vehicle: "vehicle",
  TaxDeferredFund: "tax_deferred_fund",
  TaxDeferredCash: "tax_deferred_cash",
  HousingFund: "housing_fund",
  Receivable: "receivable",
  Crypto: "crypto",
  ForeignCurrency: "foreign_currency",
  Unknown: "unknown"
} as const;

export type AccountKindValue = (typeof AccountKind)[keyof typeof AccountKind];

// Declaration order matches kind.go's const block and AllAccountKinds().
export const ACCOUNT_KIND_VALUES: readonly AccountKindValue[] = [
  AccountKind.BankCurrent,
  AccountKind.BankSavings,
  AccountKind.CashEquivalent,
  AccountKind.MutualFund,
  AccountKind.Stock,
  AccountKind.Bond,
  AccountKind.StructuredDeposit,
  AccountKind.RealEstate,
  AccountKind.Vehicle,
  AccountKind.TaxDeferredFund,
  AccountKind.TaxDeferredCash,
  AccountKind.HousingFund,
  AccountKind.Receivable,
  AccountKind.Crypto,
  AccountKind.ForeignCurrency,
  AccountKind.Unknown
] as const;

export function isValidAccountKind(v: string): v is AccountKindValue {
  return (ACCOUNT_KIND_VALUES as readonly string[]).includes(v);
}

type LabelMap = Record<AccountKindValue, string>;

const LABELS_ZH_CN: LabelMap = {
  bank_current: "活期",
  bank_savings: "定期/通知存款",
  cash_equivalent: "货币基金",
  mutual_fund: "开放式基金",
  stock: "股票",
  bond: "债券/国债逆回购",
  structured_deposit: "结构性存款/银行理财",
  real_estate: "不动产",
  vehicle: "车辆",
  tax_deferred_fund: "个人养老金基金",
  tax_deferred_cash: "个人养老金现金",
  housing_fund: "公积金",
  receivable: "应收款",
  crypto: "加密货币",
  foreign_currency: "外币",
  unknown: "未分类"
};

const LABELS_EN: LabelMap = {
  bank_current: "Bank (Current)",
  bank_savings: "Bank (Savings)",
  cash_equivalent: "Cash Equivalent",
  mutual_fund: "Mutual Fund",
  stock: "Stock",
  bond: "Bond",
  structured_deposit: "Structured Deposit",
  real_estate: "Real Estate",
  vehicle: "Vehicle",
  tax_deferred_fund: "Tax-Deferred Fund",
  tax_deferred_cash: "Tax-Deferred Cash",
  housing_fund: "Housing Fund",
  receivable: "Receivable",
  crypto: "Crypto",
  foreign_currency: "Foreign Currency",
  unknown: "Unknown"
};

export type SupportedLocale = "zh-CN" | "en";

/**
 * Returns the localized display label for an AccountKind.
 *
 * Defaults to zh-CN to match the primary user audience. Passing a locale we
 * don't have a table for falls back to zh-CN. Passing a non-enum kind returns
 * the raw value so the UI can still render something.
 */
export function getKindLabel(kind: AccountKindValue, locale: SupportedLocale = "zh-CN"): string {
  if (!isValidAccountKind(kind as string)) {
    return kind as string;
  }
  const table = locale === "en" ? LABELS_EN : LABELS_ZH_CN;
  return table[kind];
}
