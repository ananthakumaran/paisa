import { describe, it as test } from "@std/testing/bdd";
import { expect } from "@std/expect";

import { applyChanges } from "./bulk_edit.ts";
import type { LedgerFile } from "./utils.ts";
import _ from "lodash";

describe("bulk_editor", () => {
  const before = Deno.readTextFileSync("fixture/main.ledger");
  const transactions = JSON.parse(
    Deno.readTextFileSync("fixture/main.transactions.json"),
  );
  Array.from(Deno.readDirSync("fixture/bulk_edit")).forEach(({ name: dir }) => {
    test(dir, () => {
      const files = Array.from(
        Deno.readDirSync(`fixture/bulk_edit/${dir}`),
        ({ name }) => name,
      );
      for (const file of files) {
        const [name, extension] = file.split(".");
        if (extension === "ledger") {
          const args = JSON.parse(
            Deno.readTextFileSync(`fixture/bulk_edit/${dir}/${name}.json`),
          );
          const after = Deno.readTextFileSync(
            `fixture/bulk_edit/${dir}/${name}.ledger`,
          );
          const ledgerFile: LedgerFile = {
            type: "file",
            name: "main.ledger",
            content: before,
            versions: [],
          };
          const {
            newFiles: [newLedgerFile],
          } = applyChanges([ledgerFile], transactions, dir, args);
          expect(_.trim(newLedgerFile.content)).toBe(_.trim(after.toString()));
        }
      }
    });
  });
});
