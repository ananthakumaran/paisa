import { expect, test } from "@playwright/test";

test.beforeAll(async ({ request }) => {
  const response = await request.post("/api/sync", { data: { journal: true } });
  expect(response.ok()).toBeTruthy();
});

for (
  const [name, path, ready] of [
    ["dashboard", "/", "text=Net worth"],
    ["transactions", "/ledger/transaction", "p.is-6"],
    ["networth", "/assets/networth", "text=Net worth"],
  ] as const
) {
  test(`@visual ${name}`, async ({ page }) => {
    await page.goto(path);
    await page.emulateMedia({ reducedMotion: "reduce" });
    await page.locator(ready).first().waitFor();
    await page.evaluate("document.fonts.ready");
    await expect(page).toHaveScreenshot(`${name}.png`, { fullPage: true });
  });
}
