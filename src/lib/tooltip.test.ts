import assert from "node:assert/strict";
import { describe, test } from "node:test";
import { fitTooltipRows, type TooltipRow } from "./tooltip";

describe("fitTooltipRows", () => {
  test("keeps every row when the tooltip fits", () => {
    const rows: TooltipRow[] = [["one"], ["two"]];

    assert.deepStrictEqual(fitTooltipRows(rows, { total: "3" }, 200), rows);
  });

  test("keeps the bottom rows and reserves space for the total", () => {
    const rows: TooltipRow[] = Array.from({ length: 10 }, (_, index) => [
      `row ${index + 1}`,
      `${index + 1}`
    ]);

    const visibleRows = fitTooltipRows(rows, { total: "55" }, 180);

    assert.deepStrictEqual(visibleRows, [
      [["… 8 rows omitted", "has-text-grey has-text-centered", "2"]],
      ["row 9", "9"],
      ["row 10", "10"]
    ]);
    assert.strictEqual(rows.length, 10);
  });

  test("reserves space for both a header and a total", () => {
    const rows: TooltipRow[] = Array.from({ length: 5 }, (_, index) => [`row ${index + 1}`]);

    assert.deepStrictEqual(fitTooltipRows(rows, { header: "Month", total: "15" }, 150), [
      [["… 5 rows omitted", "has-text-grey has-text-centered", "1"]]
    ]);
  });
});
