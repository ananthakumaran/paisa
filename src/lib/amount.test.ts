import assert from "node:assert/strict";
import { describe, test } from "node:test";
import { BigNumber } from "bignumber.js";
import { Amount, parseAmount } from "./amount";

describe("parseAmount", () => {
  for (const [locale, input, expected] of [
    ["en-US", "$1,234.56", "1234.56"],
    ["en-IN", "₹ 12,34,567.89", "1234567.89"],
    ["en-IN", "₹ 1,23,4.5600", "1234.5600"],
    ["de-DE", "1.234,56 €", "1234.56"],
    ["de-DE", "1.23.4,5600 €", "1234.5600"],
    ["de-DE", "1.234", "1234"],
    ["en-US", "1.234", "1.234"],
    ["fr-FR", "1\u202f234,56 €", "1234.56"],
    ["fr-FR", "1\u00a0234,56 €", "1234.56"],
    ["fr-FR", "1 234,56 €", "1234.56"],
    ["de-CH", "CHF 1’234.56", "1234.56"],
    ["en-US", "(₹ 1,234.56)", "-1234.56"],
    ["de-DE", "(1.234,56 €)", "-1234.56"],
    ["en-US", "−123.45", "-123.45"],
    ["de-DE", "−123,45 €", "-123.45"],
    ["en-US", "  +$123.45  ", "+123.45"],
    ["en-US", "12abc34", "1234"],
    ["en-US", "EUR 123.45", "123.45"],
    ["en-US", "123.45 EUR", "123.45"],
    ["en-US", "USD 3.66E-04", "0.000366"],
    ["en-US", "1e+3 EUR", "1000"],
    ["de-DE", "3,66e-04 €", "0.000366"],
    ["en-US", "0", "0"],
    ["en-US", "0.00", "0.00"],
    ["en-US", "9007199254740993.123456789", "9007199254740993.123456789"]
  ]) {
    test(`${locale}: ${input}`, () => {
      const amount = parseAmount(input, locale);
      assert.ok(amount);
      assert.strictEqual(amount.toString(), expected);
      assert.ok(amount.value.isEqualTo(new BigNumber(expected)));
    });
  }

  test("accepts locale-formatted separators and signs", () => {
    for (const locale of ["en-IN", "en-US", "de-DE", "fr-FR", "sv-SE"]) {
      for (const value of [-1234.56, 0, 1234.56]) {
        const input = new Intl.NumberFormat(locale).format(value);
        assert.strictEqual(parseAmount(input, locale)?.value.toFixed(), value.toString());
      }
    }
  });

  test("does not reinterpret numbers or previously parsed amounts", () => {
    assert.strictEqual(parseAmount(1234.56, "de-DE")?.toString(), "1234.56");
    assert.strictEqual(parseAmount(0.00000001, "de-DE")?.toString(), "0.00000001");
    const amount = parseAmount("1.234,5600", "de-DE");
    assert.strictEqual(parseAmount(amount, "en-US"), amount);
    assert.strictEqual(parseAmount(amount, "de-DE"), amount);
  });

  test("rejects values that still are not numbers after permissive cleanup", () => {
    for (const input of [
      "",
      "  ",
      "N/A",
      "EUR",
      "Infinity",
      "--12",
      "1.2.3",
      "1e2e3",
      NaN,
      Infinity,
      -Infinity,
      undefined,
      null,
      true,
      {},
      []
    ]) {
      assert.strictEqual(parseAmount(input, "en-US"), undefined);
    }
  });

  test("arithmetic results render without exponent notation or negative zero", () => {
    assert.strictEqual(new Amount(new BigNumber("-1e-8")).toString(), "-0.00000001");
    assert.strictEqual(new Amount(new BigNumber("-0")).toString(), "0");
  });
});
