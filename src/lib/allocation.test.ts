import { describe, expect, test } from "bun:test";
import { allocationNodeLabel, timelineGroupLabel } from "./allocation_label";

describe("allocationNodeLabel", () => {
  test("root node passes through verbatim", () => {
    expect(allocationNodeLabel("Allocation")).toBe("Allocation");
  });

  test("kind bucket renders the localized label", () => {
    // mutual_fund → 基金/开放式基金 (zh-CN)
    expect(allocationNodeLabel("Allocation:mutual_fund")).toBe("开放式基金");
    expect(allocationNodeLabel("Allocation:bank_current")).toBe("活期");
    expect(allocationNodeLabel("Allocation:real_estate")).toBe("不动产");
    expect(allocationNodeLabel("Allocation:unknown")).toBe("未分类");
  });

  test("unknown kind bucket falls back to raw kind code", () => {
    expect(allocationNodeLabel("Allocation:not_a_real_kind")).toBe("not_a_real_kind");
  });

  test("leaf node uses original_account when supplied", () => {
    const agg = {
      account: "Allocation:mutual_fund:Assets__Saving__CMB__Fund",
      original_account: "Assets:Saving:CMB:Fund",
      kind: "mutual_fund"
    };
    expect(allocationNodeLabel("Allocation:mutual_fund:Assets__Saving__CMB__Fund", agg)).toBe(
      "Assets:Saving:CMB:Fund"
    );
  });

  test("leaf node without original_account restores ':' from sanitized id", () => {
    // No aggregate.original_account supplied — the sanitizer collapses ':'
    // to '__'; the fallback path splits that back out so the user still
    // sees a real-looking ledger account name.
    expect(allocationNodeLabel("Allocation:bank_current:Assets__Saving__CMB")).toBe(
      "Assets:Saving:CMB"
    );
  });
});

describe("timelineGroupLabel", () => {
  test("maps known kind codes to localized labels", () => {
    expect(timelineGroupLabel("mutual_fund")).toBe("开放式基金");
    expect(timelineGroupLabel("real_estate")).toBe("不动产");
    expect(timelineGroupLabel("unknown")).toBe("未分类");
  });

  test("unknown codes pass through unchanged", () => {
    expect(timelineGroupLabel("future_kind")).toBe("future_kind");
  });
});
