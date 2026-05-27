// Locale / currency / number / date formatting helpers.
//
// Lives in its own module so it can be imported by unit tests without
// pulling in SvelteKit-only bindings (`$app/navigation`, etc.) from utils.ts.
// Consumers should keep importing from `$lib/utils` — that module re-exports
// every public symbol from here.

import dayjs from "dayjs";
import updateLocale from "dayjs/plugin/updateLocale";
import { get } from "svelte/store";
import { obscure } from "../persisted_store";

// One-time locale wiring. Safe to run at module load: paisa disables SSR
// in src/routes/(app)/+layout.ts, so this only executes in the browser /
// happy-dom test environment.
//
// We inline the zh-cn locale definition (rather than `import "dayjs/locale/zh-cn"`)
// because that locale file ships as a UMD bundle whose IIFE registration
// call gets tree-shaken by Vite in production builds, leaving the locale
// unregistered. Defining + registering the object directly ties the
// side effect to a live expression so the bundler keeps it.
const zhCnLocale = {
  name: "zh-cn",
  weekdays: "星期日_星期一_星期二_星期三_星期四_星期五_星期六".split("_"),
  weekdaysShort: "周日_周一_周二_周三_周四_周五_周六".split("_"),
  weekdaysMin: "日_一_二_三_四_五_六".split("_"),
  months: "一月_二月_三月_四月_五月_六月_七月_八月_九月_十月_十一月_十二月".split("_"),
  monthsShort: "1月_2月_3月_4月_5月_6月_7月_8月_9月_10月_11月_12月".split("_"),
  weekStart: 1,
  yearStart: 4,
  formats: {
    LT: "HH:mm",
    LTS: "HH:mm:ss",
    L: "YYYY/MM/DD",
    LL: "YYYY年M月D日",
    LLL: "YYYY年M月D日Ah点mm分",
    LLLL: "YYYY年M月D日ddddAh点mm分",
    l: "YYYY/M/D",
    ll: "YYYY年M月D日",
    lll: "YYYY年M月D日 HH:mm",
    llll: "YYYY年M月D日dddd HH:mm"
  },
  relativeTime: {
    future: "%s内",
    past: "%s前",
    s: "几秒",
    m: "1 分钟",
    mm: "%d 分钟",
    h: "1 小时",
    hh: "%d 小时",
    d: "1 天",
    dd: "%d 天",
    M: "1 个月",
    MM: "%d 个月",
    y: "1 年",
    yy: "%d 年"
  }
};
dayjs.locale(zhCnLocale, null, true);
dayjs.extend(updateLocale);

const DEFAULT_LOCALE = "en";
// Only used when USER_CONFIG hasn't loaded yet (initial render, tests).
const DEFAULT_CURRENCY = "INR";

function userLocale(): string {
  if (typeof globalThis.USER_CONFIG !== "undefined" && globalThis.USER_CONFIG?.locale) {
    return globalThis.USER_CONFIG.locale;
  }
  return DEFAULT_LOCALE;
}

function userDefaultCurrency(): string {
  if (typeof globalThis.USER_CONFIG !== "undefined" && globalThis.USER_CONFIG?.default_currency) {
    return globalThis.USER_CONFIG.default_currency;
  }
  return DEFAULT_CURRENCY;
}

function userPrecision(): number {
  if (typeof globalThis.USER_CONFIG !== "undefined") {
    const p = globalThis.USER_CONFIG?.display_precision;
    if (typeof p === "number") return p;
  }
  return 0;
}

function normalize(value: number): number {
  if (typeof obscure !== "undefined") {
    try {
      if (get(obscure)) {
        value = 0;
      }
    } catch {
      // tests without the store mounted — ignore
    }
  }

  // minus 0
  if (1 / value === -Infinity) {
    value = 0;
  }

  if (!Number.isFinite(value)) {
    value = 0;
  }

  return value;
}

function unicodeMinusReplace(value: string): string {
  // Replace any leading "-" or a "-" right after the currency symbol/prefix.
  return value.replace(/^-/, "−").replace(/^([^0-9-−]+)-/, "$1−");
}

// ISO 4217 codes are exactly three uppercase letters. Anything else
// (commodity tickers like "AAPL", literal symbols like "¥") is rejected
// so we fall back to plain-number formatting with the literal as a prefix.
function isLikelyIsoCurrency(code: string): boolean {
  return /^[A-Z]{3}$/.test(code);
}

// Resolve a currency override. d3.axis.tickFormat passes (value, index)
// to its formatter, so we only honour string overrides — a numeric tick
// index must not be mistaken for a currency.
function resolveCurrency(currency: unknown): string {
  if (typeof currency === "string" && currency.length > 0) {
    return currency;
  }
  return userDefaultCurrency();
}

interface CurrencyParts {
  symbol: string;
  number: string;
  isNegative: boolean;
}

// Format `value` using Intl currency style and split the result into a
// (symbol, number, isNegative) triple. The symbol carries Intl's
// locale-aware resolution (CNY→¥, USD→$, HKD→HK$, INR→₹, …).
function formatCurrencyParts(
  value: number,
  currency: string,
  options: Intl.NumberFormatOptions
): CurrencyParts {
  const formatter = new Intl.NumberFormat(userLocale(), {
    style: "currency",
    currency,
    currencyDisplay: "symbol",
    ...options
  });

  const parts = formatter.formatToParts(value);
  let symbol = "";
  let number = "";
  let isNegative = false;
  for (const p of parts) {
    if (p.type === "currency") {
      symbol += p.value;
    } else if (p.type === "minusSign") {
      isNegative = true;
    } else if (p.type === "literal") {
      // Discard spaces between the currency and the number — we control
      // glue ourselves so output stays consistent across locales.
    } else {
      number += p.value;
    }
  }
  return { symbol, number, isNegative };
}

// Format `value` using literal `prefix` + plain locale-grouped number.
// Used when `currency` is not a recognisable ISO code (e.g. "AAPL",
// "¥" literal in the user's config).
function formatWithLiteralPrefix(
  value: number,
  prefix: string,
  options: Intl.NumberFormatOptions
): string {
  const num = unicodeMinusReplace(value.toLocaleString(userLocale(), options));
  if (num.startsWith("−")) {
    return "−" + prefix + num.slice(1);
  }
  return prefix + num;
}

/**
 * Format `value` as a currency string with the locale-appropriate symbol
 * prefix.
 *
 * - `formatCurrency(value)` uses `USER_CONFIG.default_currency`.
 * - `formatCurrency(value, "USD")` (or any ISO 4217 code) renders with
 *   that currency: e.g. `$11.23` / `US$11.23` / `HK$1,663.80`.
 * - `formatCurrency(value, 2)` is the legacy overload — emits a plain
 *   locale-grouped number with the given fixed precision and NO currency
 *   prefix. The price page depends on this.
 */
export function formatCurrency(value: number, currency?: string): string;
export function formatCurrency(value: number, precision: number): string;
export function formatCurrency(value: number, arg?: string | number): string {
  const v = normalize(value);

  if (typeof arg === "number") {
    // Legacy: caller wants a plain locale-grouped number at this precision.
    return unicodeMinusReplace(
      v.toLocaleString(userLocale(), {
        minimumFractionDigits: arg,
        maximumFractionDigits: arg
      })
    );
  }

  const currency = resolveCurrency(arg);
  const precision = userPrecision();

  // Non-ISO commodity / literal symbol — render with literal prefix.
  if (!isLikelyIsoCurrency(currency)) {
    return formatWithLiteralPrefix(v, currency, {
      minimumFractionDigits: precision,
      maximumFractionDigits: precision
    });
  }

  try {
    const { symbol, number, isNegative } = formatCurrencyParts(v, currency, {});
    return (isNegative ? "−" : "") + symbol + number;
  } catch {
    // Unknown ISO code (extremely rare given the regex guard above) — fall back.
    return formatWithLiteralPrefix(v, currency, {
      minimumFractionDigits: precision,
      maximumFractionDigits: precision
    });
  }
}

/**
 * Format a number for d3 axis ticks and other "crude" displays.
 *
 * - For zh-* locales, large values abbreviate as `万` / `亿` (e.g.
 *   `¥734.45 万`, `¥1.2 亿`) so axis labels stay short.
 * - For all other locales, fall back to Intl's compact notation so en-IN
 *   renders lakh/crore (`₹73.44L`), en-US renders short scale (`$7.34M`),
 *   etc. — this keeps existing fixtures / muscle memory intact.
 *
 * The optional second parameter accepts `string | number` so the helper
 * is type-compatible with `d3.axis.tickFormat((value, index) => string)`.
 * A numeric tick index passed by d3 is ignored.
 */
export function formatCurrencyCrude(value: number, currency?: string | number): string {
  return formatCurrencyCrudeWithPrecision(value, -1, currency);
}

export function formatCurrencyCrudeWithPrecision(
  value: number,
  precision: number,
  currency?: string | number
): string {
  const v = normalize(value);
  const currencyCode = resolveCurrency(currency);
  const locale = userLocale();
  const useChineseAbbreviation = /^zh\b/i.test(locale);

  if (useChineseAbbreviation) {
    return chineseAbbreviation(v, precision, currencyCode);
  }

  // Non-zh: lean on Intl compact notation.
  return compactCurrency(v, precision, currencyCode);
}

function chineseAbbreviation(value: number, precision: number, currency: string): string {
  const abs = Math.abs(value);
  let scaled = value;
  let suffix = "";
  if (abs >= 1e8) {
    scaled = value / 1e8;
    suffix = " 亿";
  } else if (abs >= 1e4) {
    scaled = value / 1e4;
    suffix = " 万";
  }

  const numberOptions: Intl.NumberFormatOptions = {};
  if (suffix === "") {
    // Below 万 — keep configured precision for raw numbers.
    if (precision < 0) {
      numberOptions.maximumFractionDigits = userPrecision();
      numberOptions.minimumFractionDigits = userPrecision();
    } else {
      numberOptions.maximumFractionDigits = precision;
      numberOptions.minimumFractionDigits = precision;
    }
  } else if (precision < 0) {
    // Crude tick formatter: keep at most 2 fractional digits, no trailing
    // zeros (so `1_200_000_000` → `¥12 亿`, not `¥12.00 亿`). Intl's currency
    // style defaults to 2 fraction digits min, so we override both bounds.
    numberOptions.maximumFractionDigits = 2;
    numberOptions.minimumFractionDigits = 0;
  } else {
    numberOptions.maximumFractionDigits = precision;
    numberOptions.minimumFractionDigits = precision;
  }

  if (!isLikelyIsoCurrency(currency)) {
    return formatWithLiteralPrefix(scaled, currency, numberOptions) + suffix;
  }

  try {
    const { symbol, number, isNegative } = formatCurrencyParts(scaled, currency, numberOptions);
    return (isNegative ? "−" : "") + symbol + number + suffix;
  } catch {
    return formatWithLiteralPrefix(scaled, currency, numberOptions) + suffix;
  }
}

function compactCurrency(value: number, precision: number, currency: string): string {
  const opts: Intl.NumberFormatOptions = { notation: "compact" };
  if (precision < 0) {
    opts.maximumFractionDigits = 2;
    opts.minimumFractionDigits = 0;
  } else {
    opts.maximumFractionDigits = precision;
    opts.minimumFractionDigits = precision;
  }

  if (!isLikelyIsoCurrency(currency)) {
    return formatWithLiteralPrefix(value, currency, opts);
  }

  try {
    const { symbol, number, isNegative } = formatCurrencyParts(value, currency, opts);
    return (isNegative ? "−" : "") + symbol + number;
  } catch {
    return formatWithLiteralPrefix(value, currency, opts);
  }
}

export function formatFloat(value: number, precision = 2): string {
  const v = normalize(value);

  return unicodeMinusReplace(
    v.toLocaleString(userLocale(), {
      minimumFractionDigits: precision,
      maximumFractionDigits: precision
    })
  );
}

export function formatFloatUptoPrecision(value: number, precision = 2): string {
  const v = normalize(value);

  return unicodeMinusReplace(
    v.toLocaleString(userLocale(), {
      maximumFractionDigits: precision
    })
  );
}

export function formatPercentage(value: number, precision = 0): string {
  const v = normalize(value);

  return unicodeMinusReplace(
    v.toLocaleString(userLocale(), {
      style: "percent",
      minimumFractionDigits: precision
    })
  );
}

export function formatFixedWidthFloat(value: number, width: number, precision = 2): string {
  const v = normalize(value);

  const formatted = unicodeMinusReplace(
    v.toLocaleString(userLocale(), {
      minimumFractionDigits: precision,
      maximumFractionDigits: precision
    })
  );

  if (formatted.length < width) {
    return formatted.padStart(width, " ");
  }
  return formatted;
}

// Map a configured `locale` (e.g. "zh-CN", "en-IN", "es-EU") to a dayjs
// locale name (`zh-cn`, `en`, `es`, …). dayjs uses lower-case BCP-47-ish
// names; we only ship `zh-cn` (registered above) and the built-in `en`.
// Unknown locales fall through to `en` which is dayjs's default.
function dayjsLocaleFor(configLocale: string): string {
  if (/^zh\b/i.test(configLocale)) return "zh-cn";
  return configLocale.toLowerCase().split("-")[0] || "en";
}

/**
 * Re-apply locale-sensitive runtime state after USER_CONFIG changes.
 * Called from src/routes/(app)/+layout.ts on app load and from the
 * settings page when the user saves a new config.
 *
 * Only switches dayjs to `zh-cn` when the configured locale is in the zh
 * family. For other locales we stick to dayjs's default `en`, which
 * preserves existing en-IN / es / etc. user behaviour.
 */
export function configUpdated(): void {
  const locale = userLocale();
  const dl = dayjsLocaleFor(locale);
  dayjs.locale(dl);
  const weekStart =
    typeof globalThis.USER_CONFIG !== "undefined" &&
    typeof globalThis.USER_CONFIG?.week_starting_day === "number"
      ? globalThis.USER_CONFIG.week_starting_day
      : dl === "zh-cn"
        ? 1
        : 0;
  dayjs.updateLocale(dl, { weekStart });
}
