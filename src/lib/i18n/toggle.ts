// Pure logic for the language toggle UI. Extracted so the toggling /
// persistence rules can be unit tested without spinning up a full Svelte
// component renderer (we don't ship @testing-library/svelte).
//
// The component (`LanguageSwitcher.svelte`) does:
//
//   const next = pickOppositeLocale($locale);
//   persistLocale(next);
//   locale.set(next);
//   location.reload();
//
// Reload is intentional — some screens read translations once in `onMount`
// via `get(t)('key')`, so an in-place locale swap would leave stale labels
// on the page. See CLAUDE.md / spec for #59.

export const STORAGE_KEY = "paisa.locale";

/**
 * Returns true when `tag` is any Chinese locale (zh, zh-CN, zh-TW, …).
 * Case-insensitive so callers don't have to normalise.
 */
export function isChineseLocale(tag: string | null | undefined): boolean {
  if (!tag) return false;
  return tag.toLowerCase().startsWith("zh");
}

/**
 * Given the current locale, return the locale to switch INTO when the user
 * clicks the toggle. Today we only flip between Chinese and English; any
 * non-Chinese locale (`en`, `en-US`, `de`, `null`, …) toggles to `zh-CN`.
 */
export function pickOppositeLocale(current: string | null | undefined): string {
  return isChineseLocale(current) ? "en" : "zh-CN";
}

/**
 * Save the user's explicit locale choice to localStorage. Failures (private
 * mode, quota exceeded, SSR with no `localStorage`) are swallowed — the in
 * memory store update still happens, the choice just won't survive a reload.
 */
export function persistLocale(tag: string): void {
  try {
    if (typeof localStorage !== "undefined") {
      localStorage.setItem(STORAGE_KEY, tag);
    }
  } catch (_) {
    // ignore — best-effort persistence
  }
}

/**
 * Read the user's previously persisted locale, if any. Returns `undefined`
 * when no value is stored or `localStorage` is unavailable (SSR / disabled).
 */
export function readPersistedLocale(): string | undefined {
  try {
    if (typeof localStorage !== "undefined") {
      return localStorage.getItem(STORAGE_KEY) ?? undefined;
    }
  } catch (_) {
    // ignore — treat as "no persisted value"
  }
  return undefined;
}
