import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests/browser",
  fullyParallel: false,
  workers: 1,
  timeout: 45_000,
  expect: {
    timeout: 10_000,
    toHaveScreenshot: { animations: "disabled", maxDiffPixelRatio: 0.01 },
  },
  reporter: [["list"], ["html", {
    outputFolder: "playwright-report",
    open: "never",
  }]],
  use: {
    baseURL: "http://127.0.0.1:5173",
    timezoneId: "UTC",
    locale: "en-IN",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  projects: [{
    name: "chromium",
    use: {
      ...devices["Desktop Chrome"],
      viewport: { width: 1440, height: 900 },
    },
  }],
  outputDir: "test-results",
});
