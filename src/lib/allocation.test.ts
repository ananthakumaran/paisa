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
      account: "Allocation:mutual_fund:Assets/Saving/CMB/Fund",
      original_account: "Assets:Saving:CMB:Fund",
      kind: "mutual_fund"
    };
    expect(allocationNodeLabel("Allocation:mutual_fund:Assets/Saving/CMB/Fund", agg)).toBe(
      "Assets:Saving:CMB:Fund"
    );
  });

  test("leaf node without original_account falls back to lastName of synthetic id", () => {
    // No aggregate.original_account supplied — the sanitized leaf has '/'
    // separators so we expect the entire trailing segment back.
    expect(allocationNodeLabel("Allocation:bank_current:Assets/Saving/CMB")).toBe(
      "Assets/Saving/CMB"
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
