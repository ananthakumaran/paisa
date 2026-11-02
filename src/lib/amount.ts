import { BigNumber } from "bignumber.js";

// Keep parsed values distinct from localized strings when helpers are nested.
// Handlebars renders this as ordinary decimal text, never scientific notation.
export class Amount {
  constructor(
    readonly value: BigNumber,
    private readonly text: string = value.toFixed()
  ) {}

  toString() {
    return this.text;
  }
}

interface NumberSymbols {
  decimal: string;
  group?: string;
  minus: string;
}

const symbolsByLocale = new Map<string, NumberSymbols>();

function numberSymbols(locale: string): NumberSymbols {
  let symbols = symbolsByLocale.get(locale);
  if (!symbols) {
    const parts = new Intl.NumberFormat(locale).formatToParts(-12345.6);
    symbols = {
      decimal: parts.find((part) => part.type === "decimal").value,
      group: parts.find((part) => part.type === "group")?.value,
      minus: parts.find((part) => part.type === "minusSign").value
    };
    symbolsByLocale.set(locale, symbols);
  }
  return symbols;
}

export function parseAmount(value: unknown, locale: string): Amount | undefined {
  if (value instanceof Amount) {
    return value;
  }

  if (typeof value === "number") {
    return Number.isFinite(value) ? new Amount(new BigNumber(value)) : undefined;
  }

  if (typeof value !== "string") {
    return;
  }

  const symbols = numberSymbols(locale);
  let text = value
    .trim()
    .replaceAll(symbols.minus, "-")
    .replaceAll("−", "-")
    .replace(/\((.+)\)/, "-$1");

  if (symbols.group) {
    text = text.replaceAll(symbols.group, "");
  }
  text = text.replaceAll(symbols.decimal, ".");

  // Be permissive about currency labels, whitespace and irregular grouping,
  // but only retain e/E when it is actually part of a numeric exponent.
  text = text
    .replace(/[eE]/g, (letter, index, source) =>
      /[0-9]/.test(source[index - 1] || "") && /^[+-]?[0-9]/.test(source.slice(index + 1))
        ? "e"
        : ""
    )
    .replace(/[^0-9.e+-]/g, "");

  const decimal = new BigNumber(text);
  if (!decimal.isFinite()) {
    return;
  }

  // Preserve fractional trailing zeros in amount(), as existing templates do.
  // Exponential input is expanded to plain decimal notation.
  return new Amount(decimal, text.includes("e") ? decimal.toFixed() : text);
}
