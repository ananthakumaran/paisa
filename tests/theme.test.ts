import assert from "node:assert/strict";
import { beforeEach, describe, test } from "node:test";
import { get } from "svelte/store";

class SystemTheme extends EventTarget {
  matches = false;

  setDark(dark: boolean) {
    this.matches = dark;
    this.dispatchEvent(new Event("change"));
  }
}

const media = new SystemTheme();
const browser = Object.assign(new EventTarget(), { matchMedia: () => media });
const root = { dataset: { theme: "" }, style: { colorScheme: "" } };
const storage = new Map<string, string>([["theme-preference", "dark"]]);
let storageUnavailable = false;

Object.defineProperties(globalThis, {
  window: { configurable: true, value: browser },
  document: { configurable: true, value: { documentElement: root } },
  localStorage: {
    configurable: true,
    value: {
      getItem(key: string) {
        if (storageUnavailable) throw new Error("Storage unavailable");
        return storage.get(key) ?? null;
      },
      setItem(key: string, value: string) {
        if (storageUnavailable) throw new Error("Storage unavailable");
        storage.set(key, value);
      }
    }
  }
});

const { getColorPreference, initializeTheme, setColorPreference, themePreference } = await import(
  "../src/lib/theme"
);

function storageChanged(key: string | null = "theme-preference") {
  browser.dispatchEvent(Object.assign(new Event("storage"), { key }));
}

test("preserves a previously saved explicit preference", () => {
  assert.equal(get(themePreference), "dark");
  assert.equal(getColorPreference(), "dark");
});

describe("theme preferences", () => {
  beforeEach(() => {
    storageUnavailable = false;
    storage.clear();
    media.matches = false;
    setColorPreference("system");
    storage.clear();
  });

  test("defaults to system and follows live changes without saving a resolved color", () => {
    const resolved: string[] = [];
    const dispose = initializeTheme((value) => resolved.push(value));
    try {
      assert.equal(get(themePreference), "system");
      assert.equal(getColorPreference(), "light");
      assert.equal(root.dataset.theme, "light");
      assert.equal(root.style.colorScheme, "light");
      media.setDark(true);
      assert.equal(getColorPreference(), "dark");
      assert.equal(root.dataset.theme, "dark");
      assert.equal(root.style.colorScheme, "dark");
      media.setDark(false);
      assert.deepEqual(resolved, ["light", "dark", "light"]);
      assert.equal(storage.get("theme-preference"), undefined);
      assert.equal(get(themePreference), "system");
    } finally {
      dispose();
    }
  });

  test("explicit modes ignore system changes, and returning to system resolves immediately", () => {
    const dispose = initializeTheme(() => undefined);
    try {
      setColorPreference("dark");
      media.setDark(false);
      assert.equal(root.dataset.theme, "dark");
      setColorPreference("light");
      media.setDark(true);
      assert.equal(root.dataset.theme, "light");
      assert.equal(storage.get("theme-preference"), "light");
      setColorPreference("system");
      assert.equal(root.dataset.theme, "dark");
      assert.equal(storage.get("theme-preference"), "system");
      media.setDark(false);
      assert.equal(root.dataset.theme, "light");
      assert.equal(storage.get("theme-preference"), "system");
    } finally {
      dispose();
    }
  });

  test("synchronizes saved preferences and treats missing or invalid values as system", () => {
    const dispose = initializeTheme(() => undefined);
    try {
      storage.set("theme-preference", "dark");
      storageChanged();
      assert.equal(get(themePreference), "dark");
      assert.equal(root.dataset.theme, "dark");
      storage.set("theme-preference", "invalid");
      storageChanged();
      assert.equal(get(themePreference), "system");
      assert.equal(root.dataset.theme, "light");
      storage.set("theme-preference", "light");
      storageChanged();
      media.setDark(true);
      assert.equal(root.dataset.theme, "light");
      storage.clear();
      storageChanged(null);
      assert.equal(get(themePreference), "system");
      assert.equal(root.dataset.theme, "dark");
    } finally {
      dispose();
    }
  });

  test("keeps working when local storage is unavailable", () => {
    const dispose = initializeTheme(() => undefined);
    try {
      storageUnavailable = true;
      storageChanged();
      assert.equal(get(themePreference), "system");
      setColorPreference("dark");
      assert.equal(root.dataset.theme, "dark");
      setColorPreference("system");
      media.setDark(true);
      assert.equal(root.dataset.theme, "dark");
    } finally {
      dispose();
    }
  });

  test("removes subscriptions and browser listeners on cleanup", () => {
    const resolved: string[] = [];
    const dispose = initializeTheme((value) => resolved.push(value));
    dispose();
    media.setDark(true);
    storage.set("theme-preference", "light");
    storageChanged();
    assert.equal(get(themePreference), "system");
    setColorPreference("dark");
    assert.deepEqual(resolved, ["light"]);
  });
});
