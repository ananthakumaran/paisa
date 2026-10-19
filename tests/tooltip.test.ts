import assert from "node:assert/strict";
import { test } from "node:test";
import { tooltip } from "../src/lib/tooltip";

test("truncated tooltip keeps the latest rows and always shows the total", () => {
  Object.defineProperty(globalThis, "window", {
    configurable: true,
    value: { innerHeight: 260 }
  });

  const rows = Array.from({ length: 8 }, (_, index) => [
    `Account ${index + 1}`,
    [`$${index + 1}`, "has-text-right"]
  ]);

  const html = tooltip(rows, { total: "$36" });

  assert.match(
    html,
    /<td class='has-text-grey has-text-centered' colspan='2'>6 earlier rows omitted<\/td>/
  );
  assert.match(html, /<td>Account 7<\/td>/);
  assert.match(html, /<td>Account 8<\/td>/);
  assert.match(html, /<td class='has-text-weight-bold'>Total<\/td>/);
  assert.match(html, /<td class='has-text-weight-bold has-text-right'>\$36<\/td>/);
});
