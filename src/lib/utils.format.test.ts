import { describe, expect, test, beforeAll } from "bun:test";
import dayjs from "dayjs";
// Imports come from ./format directly because ./utils transitively pulls
// in `$app/navigation`, which only resolves through Vite/SvelteKit. The
// app-level surface (which the rest of the codebase imports from $lib/utils)
// re-exports the same symbols.
import { formatCurrency, formatCurrencyCrude, formatCurrencyCrudeWithPrecision } from "./format";

beforeAll(() => {
  // Mimic what +layout.ts does after fetching /api/config.
  // The test runs under Bun (native ICU), so Intl.NumberFormat("zh-CN", ...) is real.
  globalThis.USER_CONFIG = {
    default_currency: "¥",
    locale: "zh-CN",
    display_precision: 0,
    financial_year_starting_month: 1,
    week_starting_day: 1,
    readonly: false,
    journal_path: "",
    db_path: "",
    amount_alignment_column: 0,
    goals: {},
    accounts: []
  } as any;
});

describe("formatCurrency", () => {
  test("prefixes the default currency symbol and groups digits", () => {
    const out = formatCurrency(7344476);
    expect(out).toContain("¥");
    expect(out).toContain("7,344,476");
  });

  test("honours an explicit currency override", () => {
    const out = formatCurrency(1234, "$");
    expect(out).toContain("$");
    expect(out).toContain("1,234");
  });

  test("renders negatives with a unicode minus", () => {
    const out = formatCurrency(-1000);
    expect(out).toContain("−");
  });

  test("renders zero as ¥0", () => {
    expect(formatCurrency(0)).toBe("¥0");
  });
});

describe("formatCurrencyCrude", () => {
  test("abbreviates millions as 万 with the currency prefix", () => {
    expect(formatCurrencyCrude(7344476)).toBe("¥734.45 万");
  });

  test("abbreviates hundreds of millions as 亿", () => {
    expect(formatCurrencyCrude(120_000_000)).toBe("¥1.2 亿");
  });

  test("12 亿 boundary stays in 亿 with whole number", () => {
    expect(formatCurrencyCrude(1_200_000_000)).toBe("¥12 亿");
  });

  test("small numbers are emitted without abbreviation", () => {
    expect(formatCurrencyCrude(1234)).toBe("¥1,234");
  });

  test("zero stays as ¥0", () => {
    expect(formatCurrencyCrude(0)).toBe("¥0");
  });

  test("honours an explicit currency override", () => {
    expect(formatCurrencyCrude(7344476, "$")).toBe("$734.45 万");
  });

  test("is safe to pass to d3.axis tickFormat (numeric 2nd arg is ignored)", () => {
    // d3 calls tickFormat as (value, index). The second arg must not be
    // mistaken for a currency override.
    const out = (formatCurrencyCrude as any)(7344476, 0);
    expect(out).toBe("¥734.45 万");
  });

  test("renders negatives with unicode minus and prefix", () => {
    const out = formatCurrencyCrude(-7344476);
    expect(out.startsWith("¥−") || out.startsWith("−¥")).toBe(true);
  });
});

describe("formatCurrencyCrudeWithPrecision", () => {
  test("respects an explicit precision when abbreviating", () => {
    expect(formatCurrencyCrudeWithPrecision(7344476, 1)).toBe("¥734.4 万");
  });
});

describe("dayjs zh-CN locale", () => {
  test("ddd renders the zh-CN abbreviated weekday", () => {
    // 2026-05-26 is a Tuesday. zh-cn dayjs renders this as "周二".
    expect(dayjs("2026-05-26").format("ddd")).toBe("周二");
  });

  test("week starts on Monday", () => {
    // 2026-05-25 is a Monday. startOf('week') should land on the same day.
    const monday = dayjs("2026-05-25");
    expect(monday.startOf("week").format("YYYY-MM-DD")).toBe("2026-05-25");
  });
});
