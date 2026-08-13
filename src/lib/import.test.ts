import { describe, it as test } from "@std/testing/bdd";
import { expect } from "@std/expect";

import { asRows, parse, render } from "./spreadsheet.ts";
import helpers from "./template_helpers.ts";
import _ from "lodash";
import Handlebars from "handlebars";
import dayjs from "dayjs";
import customParseFormat from "dayjs/plugin/customParseFormat.js";
dayjs.extend(customParseFormat);
import isSameOrBefore from "dayjs/plugin/isSameOrBefore.js";
dayjs.extend(isSameOrBefore);
import utc from "dayjs/plugin/utc.js";
import timezone from "dayjs/plugin/timezone.js"; // dependent on utc plugin
dayjs.extend(utc);
dayjs.extend(timezone);
import localeData from "dayjs/plugin/localeData.js";
dayjs.extend(localeData);
import updateLocale from "dayjs/plugin/updateLocale.js";
dayjs.extend(updateLocale);

Handlebars.registerHelper(
  _.mapValues(helpers, (helper, name) => {
    return function (this: any, ...args: any[]) {
      try {
        return helper.apply(this, args);
      } catch (e) {
        console.log("Error in helper", name, args, e);
      }
    };
  }),
);

describe("import", () => {
  Array.from(Deno.readDirSync("fixture/import")).forEach(({ name: dir }) => {
    test(dir, async () => {
      const files = Array.from(
        Deno.readDirSync(`fixture/import/${dir}`),
        ({ name }) => name,
      );
      for (const file of files) {
        const [name, extension] = file.split(".");
        if (extension === "ledger") {
          const inputFile = _.find(
            files,
            (f) => f != file && f.startsWith(name),
          );
          if (!inputFile || inputFile.endsWith(".pdf")) {
            break;
          }
          const input = Deno.readFileSync(`fixture/import/${dir}/${inputFile}`);
          const output = Deno.readTextFileSync(`fixture/import/${dir}/${file}`);
          const template = Deno.readTextFileSync(
            `internal/model/template/templates/${dir}.handlebars`,
          );

          const compiled = Handlebars.compile(template);
          const result = await parse(new File([input as any], inputFile));
          const rows = asRows(result);

          const actual = render(rows, compiled, { trim: true });

          expect(actual).toBe(_.trim(output));
        }
      }
    });
  });
});

describe("template helpers", () => {
  test("acronym", () => {
    expect(helpers.acronym("Foo Bar baz")).toBe("FBB");
    expect(helpers.acronym("foo   the bar")).toBe("FB");
    expect(helpers.acronym("Motital S & P 500")).toBe("MSP");
    expect(helpers.acronym("Axis Liquid Growth Direct Plan")).toBe("AL");
  });
});
