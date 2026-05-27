// Locale / currency / number / date formatting helpers.
//
// Lives in its own module so it can be imported by unit tests without
// pulling in SvelteKit-only bindings (`$app/navigation`, etc.) from utils.ts.
// Consumers should keep importing from `$lib/utils` — that module re-exports
// every public symbol from here.

import dayjs from "dayjs";
import "dayjs/locale/zh-cn";
import updateLocale from "dayjs/plugin/updateLocale";
import { get } from "svelte/store";
import { obscure } from "../persisted_store";

// One-time locale wiring. Safe to run at module load: paisa disables SSR
// in src/routes/(app)/+layout.ts, so this only executes in the browser /
// happy-dom test environment.
dayjs.extend(updateLocale);
dayjs.locale("zh-cn");
dayjs.updateLocale("zh-cn", { weekStart: 1 });

const DEFAULT_LOCALE = "zh-CN";
const DEFAULT_CURRENCY_SYMBOL = "¥";

function userLocale(): string {
  if (typeof globalThis.USER_CONFIG !== "undefined" && globalThis.USER_CONFIG?.locale) {
    return globalThis.USER_CONFIG.locale;
  }
  return DEFAULT_LOCALE;
}

function userCurrencySymbol(): string {
  if (typeof globalThis.USER_CONFIG !== "undefined" && globalThis.USER_CONFIG?.default_currency) {
    return globalThis.USER_CONFIG.default_currency;
  }
  return DEFAULT_CURRENCY_SYMBOL;
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

function unicodeMinus(value: string): string {
  return value.replace(/^-/, "−");
}

// Resolve a user-supplied "currency override" arg. d3.axis.tickFormat passes
// (value, index) to its formatter, so we only honour string overrides — a
// numeric tick index must be ignored.
function resolveCurrency(currency: unknown): string {
  if (typeof currency === "string" && currency.length > 0) {
    return currency;
  }
  return userCurrencySymbol();
}

/**
 * Format a number as a currency-prefixed string, e.g. `¥7,344,476`.
 * The prefix defaults to the configured `default_currency` symbol; pass a
 * string override (e.g. `$`, `HK$`) to use a different symbol. Number
 * grouping follows the configured `locale` (zh-CN by default).
 */
export function formatCurrency(value: number, currency?: string): string;
export function formatCurrency(value: number, precision: number): string;
export function formatCurrency(value: number, arg?: string | number): string {
  let precision = userPrecision();
  let currency: string | undefined;

  if (typeof arg === "number") {
    // Back-compat: legacy callers pass `formatCurrency(value, precision)`.
    precision = arg;
  } else if (typeof arg === "string") {
    currency = arg;
  }

  const prefix = resolveCurrency(currency);
  const v = normalize(value);

  const number = unicodeMinus(
    v.toLocaleString(userLocale(), {
      minimumFractionDigits: precision,
      maximumFractionDigits: precision
    })
  );

  if (number.startsWith("−")) {
    return "−" + prefix + number.slice(1);
  }
  return prefix + number;
}

/**
 * Format a number for d3 axis ticks: large values abbreviate as `万` /
 * `亿`, smaller values render grouped. Always prefixed with the resolved
 * currency symbol. Designed to be safe as a direct argument to
 * `d3.axis().tickFormat(...)` — d3's `(d, index)` invocation does not
 * trigger the override path.
 */
// The optional second parameter accepts `string | number` so that this
// helper is type-compatible with d3.axis.tickFormat((value, index) => string).
// `resolveCurrency` ignores non-string values, so a numeric tick index
// passed by d3 has no effect.
export function formatCurrencyCrude(value: number, currency?: string | number): string {
  return formatCurrencyCrudeWithPrecision(value, -1, currency);
}

export function formatCurrencyCrudeWithPrecision(
  value: number,
  precision: number,
  currency?: string | number
): string {
  const prefix = resolveCurrency(currency);
  const v = normalize(value);
  const abs = Math.abs(v);

  let scaled = v;
  let suffix = "";

  if (abs >= 1e8) {
    scaled = v / 1e8;
    suffix = " 亿";
  } else if (abs >= 1e4) {
    scaled = v / 1e4;
    suffix = " 万";
  }

  const options: Intl.NumberFormatOptions = {};
  if (suffix === "") {
    // raw value — keep configured precision behaviour
    if (precision < 0) {
      options.maximumFractionDigits = userPrecision();
      options.minimumFractionDigits = userPrecision();
    } else {
      options.maximumFractionDigits = precision;
      options.minimumFractionDigits = precision;
    }
  } else if (precision < 0) {
    options.maximumFractionDigits = 2;
  } else {
    options.maximumFractionDigits = precision;
    options.minimumFractionDigits = precision;
  }

  const numberPart = unicodeMinus(scaled.toLocaleString(userLocale(), options));
  if (numberPart.startsWith("−")) {
    return "−" + prefix + numberPart.slice(1) + suffix;
  }
  return prefix + numberPart + suffix;
}

export function formatFloat(value: number, precision = 2): string {
  const v = normalize(value);

  return unicodeMinus(
    v.toLocaleString(userLocale(), {
      minimumFractionDigits: precision,
      maximumFractionDigits: precision
    })
  );
}

export function formatFloatUptoPrecision(value: number, precision = 2): string {
  const v = normalize(value);

  return unicodeMinus(
    v.toLocaleString(userLocale(), {
      maximumFractionDigits: precision
    })
  );
}

export function formatPercentage(value: number, precision = 0): string {
  const v = normalize(value);

  return unicodeMinus(
    v.toLocaleString(userLocale(), {
      style: "percent",
      minimumFractionDigits: precision
    })
  );
}

export function formatFixedWidthFloat(value: number, width: number, precision = 2): string {
  const v = normalize(value);

  const formatted = unicodeMinus(
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

/**
 * Re-apply locale-sensitive runtime state after USER_CONFIG changes.
 * Called from src/routes/(app)/+layout.ts on app load and from the
 * settings page when the user saves a new config.
 */
export function configUpdated(): void {
  // The Intl.NumberFormat side is locale-driven and reads USER_CONFIG
  // lazily on every call, so it just works. The only thing we have to
  // re-apply is the dayjs week-start preference, which is per-locale.
  dayjs.locale("zh-cn");
  const weekStart =
    typeof globalThis.USER_CONFIG !== "undefined" &&
    typeof globalThis.USER_CONFIG?.week_starting_day === "number"
      ? globalThis.USER_CONFIG.week_starting_day
      : 1;
  dayjs.updateLocale("zh-cn", { weekStart });
}
