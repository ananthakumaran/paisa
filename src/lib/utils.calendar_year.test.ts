import { describe, expect, test } from "bun:test";
import dayjs from "dayjs";
// Imports come from ./calendar_year directly because ./utils transitively pulls
// in `$app/navigation`, which only resolves through Vite/SvelteKit.
import { calendarYear, forEachCalendarYear } from "./calendar_year";

// These tests pin the calendar-year (Jan-Dec) aggregation contract that
// replaced Indian fiscal year (Apr-Mar) display logic. The format must
// be a plain 4-digit year like "2024" — never a range like "2024-25".
//
// See issue #5.

describe("calendarYear", () => {
  test("returns the year of a January date as a 4-digit string", () => {
    expect(calendarYear(dayjs("2024-01-15"))).toBe("2024");
  });

  test("returns the year of a mid-year date", () => {
    expect(calendarYear(dayjs("2024-06-15"))).toBe("2024");
  });

  test("returns the year of a December date", () => {
    expect(calendarYear(dayjs("2024-12-31"))).toBe("2024");
  });

  test("never returns a range like 2024-25", () => {
    const out = calendarYear(dayjs("2024-06-15"));
    expect(out).not.toContain("-");
    expect(out).not.toContain(" ");
  });
});

describe("forEachCalendarYear", () => {
  test("iterates one entry per calendar year inclusive", () => {
    const visited: string[] = [];
    const years = forEachCalendarYear(dayjs("2021-06-15"), dayjs("2024-02-01"), (d) => {
      visited.push(d.format("YYYY-MM-DD"));
    });
    // Years should be 2021, 2022, 2023, 2024 — all anchored to Jan 1.
    expect(visited).toEqual(["2021-01-01", "2022-01-01", "2023-01-01", "2024-01-01"]);
    expect(years.map((d) => d.format("YYYY-MM-DD"))).toEqual([
      "2021-01-01",
      "2022-01-01",
      "2023-01-01",
      "2024-01-01"
    ]);
  });

  test("single-year range returns one entry", () => {
    const visited: string[] = [];
    forEachCalendarYear(dayjs("2024-03-01"), dayjs("2024-09-01"), (d) =>
      visited.push(d.format("YYYY-MM-DD"))
    );
    expect(visited).toEqual(["2024-01-01"]);
  });
});
