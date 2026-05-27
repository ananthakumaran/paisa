import { describe, expect, test } from "bun:test";
import { filterRefunds, isRefundPosting } from "./refund";

// These tests pin the M3-G refund discipline on the frontend:
// a posting whose amount is negative inside an Expenses:* account
// is a refund (style A in issue #25), and the page toggle decides
// whether the chart sums them in (net) or drops them (gross).

describe("isRefundPosting", () => {
  test("positive amount is a real expense", () => {
    expect(isRefundPosting({ amount: 100 })).toBe(false);
  });
  test("negative amount is a refund", () => {
    expect(isRefundPosting({ amount: -50 })).toBe(true);
  });
  test("zero amount is not a refund (degenerate, treat as expense)", () => {
    expect(isRefundPosting({ amount: 0 })).toBe(false);
  });
});

describe("filterRefunds", () => {
  const postings = [
    { amount: 30000, account: "Expenses:Transport:Train" },
    { amount: -2990, account: "Expenses:Transport:Train" },
    { amount: 500, account: "Expenses:Food" }
  ];

  test("net view (showGross=false) keeps refunds — bar shows actual spend", () => {
    const out = filterRefunds(postings, false);
    expect(out).toHaveLength(3);
    expect(out.reduce((s, p) => s + p.amount, 0)).toBe(27510);
  });

  test("gross view (showGross=true) drops refunds — bar shows original outflow", () => {
    const out = filterRefunds(postings, true);
    expect(out).toHaveLength(2);
    expect(out.reduce((s, p) => s + p.amount, 0)).toBe(30500);
    expect(out.every((p) => p.amount >= 0)).toBe(true);
  });

  test("toggle is reversible — same input yields stable filtered set", () => {
    const a = filterRefunds(postings, true);
    const b = filterRefunds(postings, true);
    expect(a).toEqual(b);
  });
});
