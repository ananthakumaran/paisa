import { afterEach, beforeEach, describe, expect, test } from "bun:test";
import {
  STORAGE_KEY,
  isChineseLocale,
  persistLocale,
  pickOppositeLocale,
  readPersistedLocale
} from "./toggle";

describe("language toggle helpers", () => {
  beforeEach(() => {
    if (typeof localStorage !== "undefined") {
      localStorage.removeItem(STORAGE_KEY);
    }
  });

  afterEach(() => {
    if (typeof localStorage !== "undefined") {
      localStorage.removeItem(STORAGE_KEY);
    }
  });

  test("isChineseLocale recognises zh / zh-CN / ZH-cn but not en", () => {
    expect(isChineseLocale("zh")).toBe(true);
    expect(isChineseLocale("zh-CN")).toBe(true);
    expect(isChineseLocale("ZH-cn")).toBe(true);
    expect(isChineseLocale("en")).toBe(false);
    expect(isChineseLocale("en-US")).toBe(false);
    expect(isChineseLocale(null)).toBe(false);
    expect(isChineseLocale(undefined)).toBe(false);
    expect(isChineseLocale("")).toBe(false);
  });

  test("pickOppositeLocale flips zh-CN to en and anything else to zh-CN", () => {
    expect(pickOppositeLocale("zh-CN")).toBe("en");
    expect(pickOppositeLocale("zh")).toBe("en");
    expect(pickOppositeLocale("en")).toBe("zh-CN");
    expect(pickOppositeLocale("en-US")).toBe("zh-CN");
    // Defensive: unset locale should also offer Chinese as the next step
    // (mirrors the spec — default UI shows `中` when current is non-zh).
    expect(pickOppositeLocale(null)).toBe("zh-CN");
    expect(pickOppositeLocale(undefined)).toBe("zh-CN");
  });

  test("persistLocale writes to localStorage under STORAGE_KEY", () => {
    persistLocale("zh-CN");
    expect(localStorage.getItem(STORAGE_KEY)).toBe("zh-CN");

    persistLocale("en");
    expect(localStorage.getItem(STORAGE_KEY)).toBe("en");
  });

  test("readPersistedLocale returns undefined when nothing is stored", () => {
    expect(readPersistedLocale()).toBeUndefined();
  });

  test("readPersistedLocale round-trips through persistLocale", () => {
    persistLocale("zh-CN");
    expect(readPersistedLocale()).toBe("zh-CN");

    persistLocale("en");
    expect(readPersistedLocale()).toBe("en");
  });
});
