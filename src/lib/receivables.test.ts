import { describe, expect, test } from "bun:test";
import dayjs from "dayjs";
import { isOverdue, sortByOutstandingDesc, type Receivable } from "./receivables";

function makeReceivable(overrides: Partial<Receivable> = {}): Receivable {
  return {
    account: "Assets:Loans:Alice",
    borrower: "Alice",
    outstanding: 1000,
    lend_date: null,
    due_date: null,
    interest_rate: 0,
    note: "",
    kind: "receivable",
    ...overrides
  };
}

describe("isOverdue", () => {
  test("returns false when due_date is null", () => {
    const r = makeReceivable({ due_date: null });
    expect(isOverdue(r, dayjs("2025-01-01"))).toBe(false);
  });

  test("returns true when due_date is strictly before now", () => {
    const r = makeReceivable({ due_date: dayjs("2024-06-01") });
    expect(isOverdue(r, dayjs("2025-01-01"))).toBe(true);
  });

  test("returns false when due_date equals today", () => {
    const r = makeReceivable({ due_date: dayjs("2025-01-01") });
    expect(isOverdue(r, dayjs("2025-01-01"))).toBe(false);
  });

  test("returns false when due_date is in the future", () => {
    const r = makeReceivable({ due_date: dayjs("2025-12-31") });
    expect(isOverdue(r, dayjs("2025-01-01"))).toBe(false);
  });
});

describe("sortByOutstandingDesc", () => {
  test("orders by outstanding desc, then account asc", () => {
    const rs = [
      makeReceivable({ account: "Assets:Loans:A", outstanding: 100 }),
      makeReceivable({ account: "Assets:Loans:Big", outstanding: 100_000 }),
      makeReceivable({ account: "Assets:Loans:Same1", outstanding: 5_000 }),
      makeReceivable({ account: "Assets:Loans:Same2", outstanding: 5_000 })
    ];
    const sorted = sortByOutstandingDesc(rs);
    expect(sorted.map((r) => r.account)).toEqual([
      "Assets:Loans:Big",
      "Assets:Loans:Same1",
      "Assets:Loans:Same2",
      "Assets:Loans:A"
    ]);
  });
});
