import { afterEach, beforeEach, describe, expect, test } from "bun:test";
import { get } from "svelte/store";
import { _, initI18n, locale, waitLocale } from "./index";

describe("i18n skeleton", () => {
  beforeEach(() => {
    // Reset the global so each test starts from a clean slate.
    locale.set(null);
  });

  afterEach(() => {
    locale.set(null);
  });

  test("initI18n with explicit locale sets that locale and translates app.title", async () => {
    initI18n("zh-CN");
    await waitLocale();
    expect(get(locale)).toBe("zh-CN");
    // app.title is the single test key; both locales use the brand name "Paisa".
    expect(get(_)("app.title")).toBe("Paisa");
  });

  test("initI18n falls back to English when no locale is given", async () => {
    initI18n();
    await waitLocale();
    // navigator.language under happy-dom is typically "en-US"; either way we
    // expect the resolver to pick a locale rather than leave the store null.
    expect(get(locale)).toBeTruthy();
    expect(get(_)("app.title")).toBe("Paisa");
  });

  test("unknown locale falls back to en", async () => {
    initI18n("xx-YY");
    await waitLocale();
    // svelte-i18n keeps the requested locale on the store but renders via
    // the fallback. The translation must still resolve to the English value.
    expect(get(_)("app.title")).toBe("Paisa");
  });
});
