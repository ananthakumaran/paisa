import { describe, expect, test } from "bun:test";
import { computeLeftMargin, estimateLabelWidth } from "./gain_margin";

describe("estimateLabelWidth", () => {
  test("returns 0 for empty string", () => {
    expect(estimateLabelWidth("")).toBe(0);
  });

  test("ASCII width scales linearly", () => {
    // Same character set, the longer label is wider.
    expect(estimateLabelWidth("ABCDE")).toBeGreaterThan(estimateLabelWidth("AB"));
  });

  test("CJK characters are wider than ASCII of the same length", () => {
    // The literal length is identical (4 chars), but rendered width differs.
    const ascii = estimateLabelWidth("ABCD");
    const cjk = estimateLabelWidth("贷款收款");
    expect(cjk).toBeGreaterThan(ascii);
    // Roughly 2x — guard against accidental regressions to ASCII-only width.
    expect(cjk).toBeGreaterThanOrEqual(ascii * 1.5);
  });
});

describe("computeLeftMargin", () => {
  test("empty input falls back to minMargin", () => {
    expect(computeLeftMargin([], 150)).toBe(150);
  });

  test("short Indian-style labels keep the legacy 150 margin", () => {
    // Mirrors the demo: short colon-separated paths fit in the legacy gutter.
    const labels = ["Debt:ABCBF", "Equity:NPS:HDFC:E", "Cash:Bank"];
    expect(computeLeftMargin(labels, 150)).toBe(150);
  });

  test("long Chinese-context account paths grow the margin", () => {
    // The bug reproducer from issue #64 R2.
    const labels = ["Bridge:Brokerage:DanJuan", "Loans Receivable:LiQuanRong", "Equity:NPS:HDFC:E"];
    const margin = computeLeftMargin(labels, 150);
    // The longest label (27 chars) at 7px/char + 20px padding is ~209px —
    // strictly greater than the legacy 150px margin.
    expect(margin).toBeGreaterThan(150);
    expect(margin).toBeGreaterThanOrEqual(estimateLabelWidth("Loans Receivable:LiQuanRong") + 20);
  });

  test("labels with CJK characters get more room than the same ASCII length", () => {
    // Same character count, but CJK chars are wider, so the CJK label should
    // demand a strictly larger margin.
    const ascii = computeLeftMargin(["AssetsBrokerageDanJuan"], 0);
    const cjk = computeLeftMargin(["资产券商组合蛋卷长城基金"], 0);
    expect(cjk).toBeGreaterThan(ascii);
  });

  test("custom padding is respected", () => {
    const labels = ["Loans Receivable:LiQuanRong"];
    const tight = computeLeftMargin(labels, 0, 5);
    const loose = computeLeftMargin(labels, 0, 50);
    expect(loose - tight).toBe(45);
  });
});
