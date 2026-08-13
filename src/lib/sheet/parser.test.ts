import { describe, it as test } from "@std/testing/bdd";
import { sheetLanguage } from "./language.ts";
import { fileTests } from "@lezer/generator/dist/test";

import { dirname, fromFileUrl, join } from "@std/path";
const caseDir = dirname(fromFileUrl(import.meta.url));

const parser = sheetLanguage.parser.configure({
  strict: false,
  dialect: "comment",
});

for (const { name: file } of Deno.readDirSync(caseDir)) {
  if (!/\.txt$/.test(file)) continue;

  const name = /^[^.]*/.exec(file)[0];
  describe(name, () => {
    for (
      const { name, run } of fileTests(
        Deno.readTextFileSync(join(caseDir, file)),
        file,
      )
    ) {
      test(name, () => run(parser));
    }
  });
}
