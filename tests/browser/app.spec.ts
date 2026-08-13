import { expect, test } from "@playwright/test";

test.beforeAll(async ({ request }) => {
  const response = await request.post("/api/sync", { data: { journal: true } });
  expect(response.ok()).toBeTruthy();
});

test("application starts without fatal browser errors", async ({ page }) => {
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  page.on("console", (message) => {
    if (message.type() === "error") errors.push(message.text());
  });
  await page.goto("/");
  await expect(page.getByRole("navigation", { name: "main navigation" }))
    .toBeVisible();
  expect(errors).toEqual([]);
});

test("dashboard renders synchronized fixture data", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByText("Net worth", { exact: true }).first())
    .toBeVisible();
  await expect(page.getByRole("link", { name: "Transactions" }).first())
    .toBeVisible();
});

test("major pages are routable", async ({ page }) => {
  for (
    const path of [
      "/ledger/transaction",
      "/ledger/editor",
      "/ledger/import",
      "/assets/networth",
      "/cash_flow/monthly",
    ]
  ) {
    await page.goto(path);
    await expect(page.locator("body")).not.toBeEmpty();
    await expect(page.getByRole("navigation", { name: "main navigation" }))
      .toBeVisible();
  }
});

test("transaction search filters fixture transactions", async ({ page }) => {
  await Promise.all([
    page.waitForResponse((response) =>
      response.url().endsWith("/api/transaction")
    ),
    page.goto("/ledger/transaction"),
  ]);
  const count = page.locator("p.is-6").filter({ hasText: "transaction(s)" });
  await expect(count).toBeVisible();
  const before = await count.textContent();
  const search = page.locator(".search-query-editor .cm-content");
  await search.click();
  await page.keyboard.type('payee = "Rent"');
  await expect(count).not.toHaveText(before ?? "");
});

test("journal editor saves and reloads changes", async ({ page }) => {
  await Promise.all([
    page.waitForResponse((response) => response.url().includes("/api/editor")),
    page.goto("/ledger/editor/main.ledger"),
  ]);
  const editor = page.locator(".cm-content").first();
  await expect(editor).toBeVisible();
  await editor.press("Control+End");
  await editor.pressSequentially("\n; browser smoke marker");
  await page.getByText("Save", { exact: true }).click();
  await page.reload();
  await expect(editor).toContainText("browser smoke marker");
});

test("import produces a preview without saving", async ({ page }) => {
  await page.goto("/ledger/import");
  await page.locator('input[type="file"]').setInputFiles(
    "fixture/import/Paytm/statement.csv",
  );
  await expect(page.locator("table")).toBeVisible();
  await expect(page.locator("button.save")).toBeVisible();
});

test("analytics pages render representative data", async ({ page }) => {
  await page.goto("/assets/networth");
  await expect(page.getByText("Net worth", { exact: true })).toBeVisible();
  await page.goto("/cash_flow/monthly");
  await expect(page.locator("svg").first()).toBeVisible();
});

test("an API failure leaves a visible error instead of a blank page", async ({ page }) => {
  await page.route(
    "**/api/dashboard",
    (route) =>
      route.fulfill({
        status: 500,
        contentType: "application/json",
        body: '{"error":"test failure"}',
      }),
  );
  await page.goto("/");
  await expect(page.locator("body")).toContainText(
    /error|failed|report this issue/i,
  );
});
