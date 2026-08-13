import { describe, it as test } from "@std/testing/bdd";
import { queryLanguage } from "./query.ts";
import { fileTests } from "@lezer/generator/dist/test";

import { dirname, fromFileUrl, join } from "@std/path";
const caseDir = dirname(fromFileUrl(import.meta.url));

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
      test(name, () => run(queryLanguage.parser));
    }
  });
}
