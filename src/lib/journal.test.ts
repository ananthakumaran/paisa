import { describe, it as test } from "@std/testing/bdd";
import { expect } from "@std/expect";
import { format } from "./journal.ts";
import fs from "node:fs";

function readFixture(name: string) {
  return fs.readFileSync(`fixture/${name}`).toString();
}

describe("journal", () => {
  test("format", () => {
    expect(format(readFixture("unformatted.ledger"))).toBe(
      readFixture("formatted.ledger"),
    );
  });
});
