import assert from "node:assert/strict";
import { describe, test } from "node:test";

import { parse, render, asRows } from "./spreadsheet";
import fs from "fs";
import helpers from "./template_helpers";
import _ from "lodash";
import Handlebars from "handlebars";
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
