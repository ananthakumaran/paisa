import assert from "node:assert/strict";
import { afterEach, beforeEach, describe, test } from "node:test";

import { parse, render, asRows } from "./spreadsheet";
import fs from "fs";
import helpers from "./template_helpers";
import _ from "lodash";
import Handlebars from "handlebars";
import * as XLSX from "xlsx";
import dayjs from "dayjs";
import customParseFormat from "dayjs/plugin/customParseFormat";
dayjs.extend(customParseFormat);
import isSameOrBefore from "dayjs/plugin/isSameOrBefore";
dayjs.extend(isSameOrBefore);
import utc from "dayjs/plugin/utc";
import timezone from "dayjs/plugin/timezone"; // dependent on utc plugin
dayjs.extend(utc);
dayjs.extend(timezone);
import localeData from "dayjs/plugin/localeData";
dayjs.extend(localeData);
import updateLocale from "dayjs/plugin/updateLocale";
dayjs.extend(updateLocale);

Handlebars.registerHelper(
  _.mapValues(helpers, (helper, name) => {
    return function (...args: any[]) {
      try {
        return helper.apply(this, args);
      } catch (e) {
        console.log("Error in helper", name, args, e);
      }
    };
  })
);

describe("import", () => {
  fs.readdirSync("fixture/import").forEach((dir) => {
    test(dir, async () => {
      const files = fs.readdirSync(`fixture/import/${dir}`);
      for (const file of files) {
        const [name, extension] = file.split(".");
        if (extension === "ledger") {
          const inputFile = _.find(files, (f) => f != file && f.startsWith(name));
          if (inputFile.endsWith(".pdf")) {
            break;
          }
          const input = fs.readFileSync(`fixture/import/${dir}/${inputFile}`);
          const output = fs.readFileSync(`fixture/import/${dir}/${file}`).toString();
          const template = fs
            .readFileSync(`internal/model/template/templates/${dir}.handlebars`)
            .toString();

          const compiled = Handlebars.compile(template);
          const result = await parse(new File([input], inputFile));
          const rows = asRows(result);

          const actual = render(rows, compiled, { trim: true });

          assert.strictEqual(actual, _.trim(output));
        }
      }
    });
  });
});

describe("template helpers", () => {
  test("acronym", () => {
    assert.strictEqual(helpers.acronym("Foo Bar baz"), "FBB");
    assert.strictEqual(helpers.acronym("foo   the bar"), "FB");
    assert.strictEqual(helpers.acronym("Motital S & P 500"), "MSP");
    assert.strictEqual(helpers.acronym("Axis Liquid Growth Direct Plan"), "AL");
  });
});

describe("numeric template helpers", () => {
  let originalConfig: UserConfig;

  beforeEach(() => {
    originalConfig = globalThis.USER_CONFIG;
    globalThis.USER_CONFIG = { ...originalConfig, locale: "en-IN", amount_alignment_column: 52 };
  });

  afterEach(() => {
    globalThis.USER_CONFIG = originalConfig;
  });

  const evaluate = (template: string, value: string | number = "") =>
    Handlebars.compile(template)({ value });

  for (const [locale, value] of [
    ["en-IN", "₹ 1,234.56"],
    ["en-US", "$1,234.56"],
    ["de-DE", "1.234,56 €"],
    ["fr-FR", "1\u202f234,56 €"]
  ]) {
    test(`numeric helpers use ${locale}`, () => {
      USER_CONFIG.locale = locale;
      assert.strictEqual(evaluate("{{amount value}}", value), "1234.56");
      assert.strictEqual(evaluate("{{negate value}}", value), "-1234.56");
      assert.strictEqual(evaluate("{{round value precision=1}}", value), "1234.6");
      assert.strictEqual(evaluate("{{gt value 1234}}", value), "true");
      assert.strictEqual(evaluate("{{gte value 1234.56}}", value), "true");
      assert.strictEqual(evaluate("{{lt value 1235}}", value), "true");
      assert.strictEqual(evaluate("{{lte value 1234.56}}", value), "true");
    });
  }

  test("negate renders tiny values without scientific notation (#284)", () => {
    assert.strictEqual(evaluate("{{negate value}}", "0.00000001"), "-0.00000001");
    assert.strictEqual(evaluate("{{negate value}}", 0.00000001), "-0.00000001");
    assert.strictEqual(evaluate("{{negate value}}", "-0.00000001"), "0.00000001");
    assert.strictEqual(evaluate("{{negate value}}", 0), "0");
  });

  test("decimal commas are not discarded (#421)", () => {
    USER_CONFIG.locale = "de-DE";
    for (const [value, expected] of [
      ["100,01", "-100.01"],
      ["-200,01", "200.01"],
      ["300,01", "-300.01"],
      ["-400,01", "400.01"]
    ]) {
      assert.strictEqual(evaluate("{{negate value}}", value), expected);
    }
  });

  test("scientific notation is accepted in input (#206)", () => {
    assert.strictEqual(evaluate("{{amount value}}", "3.66e-04"), "0.000366");
    assert.strictEqual(evaluate("{{negate value}}", "1E+3"), "-1000");
    assert.strictEqual(evaluate("{{amount value}}", "1e3 EUR"), "1000");
    USER_CONFIG.locale = "de-DE";
    assert.strictEqual(evaluate("{{amount value}}", "3,66e-04 €"), "0.000366");
  });

  test("numeric inputs do not undergo locale parsing", () => {
    USER_CONFIG.locale = "de-DE";
    assert.strictEqual(evaluate("{{amount value}}", 1234.56), "1234.56");
    assert.strictEqual(evaluate("{{negate value}}", 1234.56), "-1234.56");
  });

  test("nested helpers do not parse canonical output as localized input", () => {
    USER_CONFIG.locale = "de-DE";
    assert.strictEqual(evaluate("{{negate (amount value)}}", "1.234,560"), "-1234.56");
    assert.strictEqual(evaluate("{{amount (negate value)}}", "1.234,56"), "-1234.56");
    assert.strictEqual(evaluate("{{negate (round value precision=2)}}", "1.234,567"), "-1234.57");
    assert.strictEqual(evaluate("{{lt (amount value) 2000}}", "1.234,56"), "true");
  });

  test("decimal precision is preserved for arithmetic and comparisons", () => {
    assert.strictEqual(
      evaluate("{{negate value}}", "9007199254740993.123456789"),
      "-9007199254740993.123456789"
    );
    assert.strictEqual(evaluate('{{gt value "9007199254740992"}}', "9007199254740993"), "true");
    assert.strictEqual(
      evaluate('{{lt value "1.000000000000000002"}}', "1.000000000000000001"),
      "true"
    );
  });

  test("round uses decimal arithmetic and retains rounding towards positive infinity on ties", () => {
    assert.strictEqual(evaluate("{{round value precision=2}}", "1.005"), "1.01");
    assert.strictEqual(evaluate('{{round value precision="2"}}', "1.005"), "1.01");
    assert.strictEqual(evaluate("{{round value}}", "-1.5"), "-1");
    assert.strictEqual(evaluate("{{round value precision=-2}}", "1250.56"), "1251");
    assert.strictEqual(evaluate("{{round value precision=8}}", "0.00000001"), "0.00000001");
  });

  test("amount retains fractional trailing zeros used by existing templates", () => {
    assert.strictEqual(evaluate("{{amount value}}", "₹ 1,234.5600"), "1234.5600");
    USER_CONFIG.locale = "de-DE";
    assert.strictEqual(evaluate("{{amount value}}", "1.234,5600 €"), "1234.5600");
  });

  test("invalid input renders empty amounts and false numeric comparisons", () => {
    for (const value of ["", "N/A", "--12", "1.2.3", NaN, Infinity]) {
      assert.strictEqual(evaluate("{{amount value}}", value), "");
      assert.strictEqual(evaluate("{{negate value}}", value), "");
      assert.strictEqual(evaluate("{{round value}}", value), "");
      for (const helper of ["gt", "gte", "lt", "lte"]) {
        assert.strictEqual(evaluate(`{{${helper} value 0}}`, value), "false");
        assert.strictEqual(evaluate(`{{${helper} 0 value}}`, value), "false");
      }
    }
    assert.strictEqual(evaluate('{{amount value default="0"}}'), "0");
    assert.strictEqual(evaluate("{{amount value default=0}}"), "0");
    assert.strictEqual(evaluate('{{amount value default="fallback"}}'), "fallback");
    assert.strictEqual(evaluate('{{amount value default="fallback"}}', "0"), "0");
  });

  test("renders localized CSV input as complete journal entries", async () => {
    USER_CONFIG.locale = "de-DE";
    const result = await parse(
      new File(['"20.10.2026";"-400,01"\n"20.10.2026";"3,66e-04"'], "amounts.csv")
    );
    const template = Handlebars.compile(
      '{{date ROW.A "DD.MM.YYYY"}} Test\n    Assets:Test  {{negate ROW.B}} EUR\n    Expenses:Test'
    );
    assert.strictEqual(
      render(asRows(result), template, { trim: true }),
      [
        "2026/10/20 Test",
        "    Assets:Test" + " ".repeat(31) + "400.01 EUR",
        "    Expenses:Test",
        "",
        "2026/10/20 Test",
        "    Assets:Test" + " ".repeat(28) + "-0.000366 EUR",
        "    Expenses:Test"
      ].join("\n")
    );
  });
});

describe("spreadsheet columns beyond Z (#276)", () => {
  test("maps multi-letter columns without changing row indices", () => {
    const rows = asRows({ data: [Array.from({ length: 703 }, (_, i) => `cell-${i + 1}`)] });
    for (const [column, expected] of [
      ["A", "cell-1"],
      ["Z", "cell-26"],
      ["AA", "cell-27"],
      ["AB", "cell-28"],
      ["AZ", "cell-52"],
      ["BA", "cell-53"],
      ["ZZ", "cell-702"],
      ["AAA", "cell-703"]
    ]) {
      assert.strictEqual(rows[0][column], expected);
    }
    assert.strictEqual(rows[0].index, 0);
    assert.strictEqual(rows[0]["["], undefined);
  });

  test("renders cell values and bare column references after Z", () => {
    const rows = asRows({ data: [Array.from({ length: 28 }, (_, i) => `cell-${i + 1}`)] });
    const template = Handlebars.compile("{{ROW.Z}}|{{ROW.AA}}|{{ROW.AB}}|{{Z}}|{{AA}}|{{AB}}");
    assert.strictEqual(render(rows, template), "cell-26|cell-27|cell-28|Z|AA|AB");
  });

  test("textRange crosses Z and AA", () => {
    const rows = asRows({ data: [Array.from({ length: 28 }, (_, i) => `cell-${i + 1}`)] });
    const template = Handlebars.compile('{{textRange Z AB separator="|"}}');
    assert.strictEqual(render(rows, template), "cell-26|cell-27|cell-28");
  });

  test("column references are available even on shorter rows", () => {
    const rows = asRows({
      data: [["header"], Array.from({ length: 28 }, (_, i) => `cell-${i + 1}`)]
    });
    const template = Handlebars.compile(
      "{{#if (eq ROW.index 0)}}{{findBelow AA}}|{{findBelow AB}}{{/if}}"
    );
    assert.strictEqual(render(rows, template), "cell-27|cell-28");
    assert.strictEqual(
      render(asRows({ data: [["header"]] }), Handlebars.compile("{{A}}|{{Y}}|{{Z}}")),
      "A|Y|Z"
    );
  });

  for (const extension of ["csv", "xlsx"]) {
    test(`imports ${extension} cells after Z into a template`, async () => {
      const cells = Array.from({ length: 28 }, (_, i) => `cell-${i + 1}`);
      let input: string | ArrayBuffer;
      if (extension === "xlsx") {
        const workbook = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(workbook, XLSX.utils.aoa_to_sheet([cells]), "Sheet1");
        input = XLSX.write(workbook, { type: "array", bookType: "xlsx" });
      } else {
        input = cells.join(",");
      }
      const result = await parse(new File([input], `wide.${extension}`));
      assert.strictEqual(
        render(asRows(result), Handlebars.compile("{{ROW.Z}}|{{ROW.AA}}|{{ROW.AB}}")),
        "cell-26|cell-27|cell-28"
      );
    });
  }
});
