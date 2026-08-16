import assert from "node:assert/strict";
import { describe, test } from "node:test";
import { format } from "./journal";
import fs from "fs";

function readFixture(name: string) {
  return fs.readFileSync(`fixture/${name}`).toString();
}

describe("journal", () => {
  test("format", () => {
    assert.strictEqual(format(readFixture("unformatted.ledger")), readFixture("formatted.ledger"));
  });
});
