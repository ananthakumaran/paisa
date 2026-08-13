import { describe, it as test } from "@std/testing/bdd";
import { expect } from "@std/expect";
import { format } from "./journal.ts";

function readFixture(name: string) {
  return Deno.readTextFileSync(`fixture/${name}`);
}

describe("journal", () => {
  test("format", () => {
    expect(format(readFixture("unformatted.ledger"))).toBe(
      readFixture("formatted.ledger"),
    );
  });
});
