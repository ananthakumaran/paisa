import { describe, expect, test } from "bun:test";
import { AccountKind, ACCOUNT_KIND_VALUES, getKindLabel, isValidAccountKind } from "./account_kind";

describe("account_kind", () => {
  test("enum mirrors Go values", () => {
    expect(ACCOUNT_KIND_VALUES).toEqual([
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
      "unknown"
    ]);
  });

  test("AccountKind constants", () => {
    expect(AccountKind.BankCurrent).toBe("bank_current");
    expect(AccountKind.BankSavings).toBe("bank_savings");
    expect(AccountKind.CashEquivalent).toBe("cash_equivalent");
    expect(AccountKind.MutualFund).toBe("mutual_fund");
    expect(AccountKind.Stock).toBe("stock");
    expect(AccountKind.Bond).toBe("bond");
    expect(AccountKind.StructuredDeposit).toBe("structured_deposit");
    expect(AccountKind.RealEstate).toBe("real_estate");
    expect(AccountKind.Vehicle).toBe("vehicle");
    expect(AccountKind.TaxDeferredFund).toBe("tax_deferred_fund");
    expect(AccountKind.TaxDeferredCash).toBe("tax_deferred_cash");
    expect(AccountKind.HousingFund).toBe("housing_fund");
    expect(AccountKind.Receivable).toBe("receivable");
    expect(AccountKind.Crypto).toBe("crypto");
    expect(AccountKind.ForeignCurrency).toBe("foreign_currency");
    expect(AccountKind.Unknown).toBe("unknown");
  });

  test("isValidAccountKind", () => {
    for (const k of ACCOUNT_KIND_VALUES) {
      expect(isValidAccountKind(k)).toBe(true);
    }
    expect(isValidAccountKind("")).toBe(false);
    expect(isValidAccountKind("BANK_CURRENT")).toBe(false);
    expect(isValidAccountKind("savings")).toBe(false);
    expect(isValidAccountKind("foo")).toBe(false);
  });

  test("getKindLabel zh-CN by default", () => {
    expect(getKindLabel("bank_current")).toBe("活期");
    expect(getKindLabel("bank_savings")).toBe("定期/通知存款");
    expect(getKindLabel("cash_equivalent")).toBe("货币基金");
    expect(getKindLabel("mutual_fund")).toBe("开放式基金");
    expect(getKindLabel("stock")).toBe("股票");
    expect(getKindLabel("bond")).toBe("债券/国债逆回购");
    expect(getKindLabel("structured_deposit")).toBe("结构性存款/银行理财");
    expect(getKindLabel("real_estate")).toBe("不动产");
    expect(getKindLabel("vehicle")).toBe("车辆");
    expect(getKindLabel("tax_deferred_fund")).toBe("个人养老金基金");
    expect(getKindLabel("tax_deferred_cash")).toBe("个人养老金现金");
    expect(getKindLabel("housing_fund")).toBe("公积金");
    expect(getKindLabel("receivable")).toBe("应收款");
    expect(getKindLabel("crypto")).toBe("加密货币");
    expect(getKindLabel("foreign_currency")).toBe("外币");
    expect(getKindLabel("unknown")).toBe("未分类");
  });

  test("getKindLabel en locale", () => {
    expect(getKindLabel("bank_current", "en")).toBe("Bank (Current)");
    expect(getKindLabel("bank_savings", "en")).toBe("Bank (Savings)");
    expect(getKindLabel("cash_equivalent", "en")).toBe("Cash Equivalent");
    expect(getKindLabel("mutual_fund", "en")).toBe("Mutual Fund");
    expect(getKindLabel("stock", "en")).toBe("Stock");
    expect(getKindLabel("bond", "en")).toBe("Bond");
    expect(getKindLabel("structured_deposit", "en")).toBe("Structured Deposit");
    expect(getKindLabel("real_estate", "en")).toBe("Real Estate");
    expect(getKindLabel("vehicle", "en")).toBe("Vehicle");
    expect(getKindLabel("tax_deferred_fund", "en")).toBe("Tax-Deferred Fund");
    expect(getKindLabel("tax_deferred_cash", "en")).toBe("Tax-Deferred Cash");
    expect(getKindLabel("housing_fund", "en")).toBe("Housing Fund");
    expect(getKindLabel("receivable", "en")).toBe("Receivable");
    expect(getKindLabel("crypto", "en")).toBe("Crypto");
    expect(getKindLabel("foreign_currency", "en")).toBe("Foreign Currency");
    expect(getKindLabel("unknown", "en")).toBe("Unknown");
  });

  test("getKindLabel unknown kind returns the raw value", () => {
    expect(getKindLabel("not_a_kind" as never)).toBe("not_a_kind");
  });
});
