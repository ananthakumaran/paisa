import { afterEach, beforeEach, describe, expect, test } from "bun:test";
import dayjs from "dayjs";
// Imports come from ./format directly because ./utils transitively pulls
// in `$app/navigation`, which only resolves through Vite/SvelteKit. The
// app-level surface (which the rest of the codebase imports from $lib/utils)
// re-exports the same symbols.
import {
  configUpdated,
  formatCurrency,
  formatCurrencyCrude,
  formatCurrencyCrudeWithPrecision
} from "./format";

interface ConfigOverrides {
  default_currency?: string;
  locale?: string;
  display_precision?: number;
  week_starting_day?: number;
}

function setConfig(overrides: ConfigOverrides = {}) {
  globalThis.USER_CONFIG = {
    default_currency: "CNY",
    locale: "zh-CN",
    display_precision: 0,
    week_starting_day: 1,
    readonly: false,
    journal_path: "",
    db_path: "",
    amount_alignment_column: 0,
    goals: {},
    accounts: [],
    ...overrides
  } as any;
}

beforeEach(() => {
  setConfig();
});

afterEach(() => {
  // Restore zh-cn defaults so unrelated tests don't see stale state.
  setConfig();
  configUpdated();
});

describe("formatCurrency (default zh-CN + CNY)", () => {
  test("ISO code CNY resolves to ¥ via Intl", () => {
    const out = formatCurrency(7344476, "CNY");
    expect(out).toBe("¥7,344,476.00");
  });

  test("default currency falls back to USER_CONFIG.default_currency", () => {
    const out = formatCurrency(7344476);
    expect(out).toBe("¥7,344,476.00");
  });

  test("USD override under zh-CN renders US$ / $", () => {
    const out = formatCurrency(11.23, "USD");
    // zh-CN's "symbol" display for USD is "US$".
    expect(out).toBe("US$11.23");
  });

  test("HKD override under zh-CN renders HK$", () => {
    const out = formatCurrency(1663.8, "HKD");
    expect(out).toBe("HK$1,663.80");
  });

  test("negative renders unicode minus before the prefix", () => {
    const out = formatCurrency(-1000, "CNY");
    expect(out).toBe("−¥1,000.00");
  });

  test("legacy (value, precision) overload skips the currency symbol", () => {
    // The price page relies on this shape.
    expect(formatCurrency(1234.5678, 2)).toBe("1,234.57");
    expect(formatCurrency(0, 4)).toBe("0.0000");
  });

  test("non-ISO commodity (e.g. AAPL) falls back to literal prefix", () => {
    const out = formatCurrency(100, "AAPL");
    expect(out.startsWith("AAPL")).toBe(true);
    expect(out).toContain("100");
  });
});

describe("formatCurrency in en-IN with INR", () => {
  test("preserves lakh grouping", () => {
    setConfig({ locale: "en-IN", default_currency: "INR" });
    const out = formatCurrency(7344476, "INR");
    expect(out).toBe("₹73,44,476.00");
  });
});

describe("formatCurrencyCrude (zh-CN locale)", () => {
  test("abbreviates millions as 万 with the resolved symbol", () => {
    expect(formatCurrencyCrude(7344476, "CNY")).toBe("¥734.45 万");
  });

  test("abbreviates hundreds of millions as 亿", () => {
    expect(formatCurrencyCrude(120_000_000, "CNY")).toBe("¥1.2 亿");
  });

  test("12亿 boundary keeps integer formatting", () => {
    expect(formatCurrencyCrude(1_200_000_000, "CNY")).toBe("¥12 亿");
  });

  test("small numbers skip abbreviation", () => {
    expect(formatCurrencyCrude(1234, "CNY")).toBe("¥1,234");
  });

  test("zero renders as ¥0", () => {
    expect(formatCurrencyCrude(0, "CNY")).toBe("¥0");
  });

  test("USD override under zh-CN still uses 万 abbreviation", () => {
    expect(formatCurrencyCrude(7344476, "USD")).toBe("US$734.45 万");
  });

  test("is safe as a d3.axis tickFormat — numeric 2nd arg is ignored", () => {
    const out = (formatCurrencyCrude as any)(7344476, 0);
    expect(out).toBe("¥734.45 万");
  });

  test("renders negatives with unicode minus and prefix", () => {
    const out = formatCurrencyCrude(-7344476, "CNY");
    expect(out).toBe("−¥734.45 万");
  });

  test("non-ISO commodity falls back to literal prefix + abbreviation", () => {
    const out = formatCurrencyCrude(120_000_000, "AAPL");
    expect(out.startsWith("AAPL")).toBe(true);
    expect(out.endsWith(" 亿")).toBe(true);
  });
});

describe("formatCurrencyCrude (non-zh locales fall back to compact)", () => {
  test("en-IN with INR uses Indian lakh/crore notation, NOT 万/亿", () => {
    setConfig({ locale: "en-IN", default_currency: "INR" });
    const out = formatCurrencyCrude(7344476, "INR");
    // Should be lakh notation like "₹73.44L", never contain a 万 or 亿.
    expect(out).not.toContain("万");
    expect(out).not.toContain("亿");
    expect(out).toContain("L"); // lakh suffix
    expect(out).toContain("₹");
  });

  test("en-US with USD uses short scale", () => {
    setConfig({ locale: "en-US", default_currency: "USD" });
    const out = formatCurrencyCrude(7344476, "USD");
    expect(out).not.toContain("万");
    expect(out).toContain("$");
    expect(out).toContain("M");
  });

  test("default-currency fallback in non-zh locale still works", () => {
    setConfig({ locale: "en-IN", default_currency: "INR" });
    const out = formatCurrencyCrude(120_000_000);
    expect(out).toContain("₹");
    expect(out).not.toContain("万");
  });
});

describe("formatCurrencyCrudeWithPrecision", () => {
  test("respects explicit precision when abbreviating (zh-CN)", () => {
    expect(formatCurrencyCrudeWithPrecision(7344476, 1, "CNY")).toBe("¥734.4 万");
  });

  test("explicit precision honoured in non-zh locale compact", () => {
    setConfig({ locale: "en-IN", default_currency: "INR" });
    const out = formatCurrencyCrudeWithPrecision(7344476, 1, "INR");
    expect(out).toContain("₹");
    // 1-decimal precision: "₹73.4L"
    expect(out).toContain("L");
    expect(out).not.toContain("万");
  });
});

describe("dayjs locale wiring", () => {
  test("configUpdated('zh-CN') puts dayjs in zh-cn with Monday week-start", () => {
    setConfig({ locale: "zh-CN", week_starting_day: 1 });
    configUpdated();
    // 2026-05-26 is a Tuesday; zh-cn abbreviation is "周二".
    expect(dayjs("2026-05-26").format("ddd")).toBe("周二");
    // 2026-05-25 is Monday; startOf('week') should land on the same day.
    expect(dayjs("2026-05-25").startOf("week").format("YYYY-MM-DD")).toBe("2026-05-25");
  });

  test("configUpdated('en-IN') stays on dayjs's default `en` locale", () => {
    setConfig({ locale: "en-IN", default_currency: "INR", week_starting_day: 0 });
    configUpdated();
    expect(dayjs("2026-05-26").format("ddd")).toBe("Tue");
  });
});
