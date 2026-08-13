import "@testing-library/jest-dom/vitest";

globalThis.USER_CONFIG = {
  accounts: [],
  default_currency: "INR",
  display_precision: 2,
  locale: "en-IN",
  financial_year_starting_month: 4,
  week_starting_day: 1,
} as typeof USER_CONFIG;

Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener() {},
    removeListener() {},
    addEventListener() {},
    removeEventListener() {},
    dispatchEvent: () => false,
  }),
});
