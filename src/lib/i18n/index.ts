// Centralised svelte-i18n setup for paisa. Locales are registered once at
// module load; the runtime locale is selected later by `initI18n()` so it
// can pick up `USER_CONFIG.locale` after `/api/config` resolves.
//
// New keys go in `locales/<tag>.json` using the `page.area.key` convention
// (see `docs/contributing/i18n.md`). Keep the bundle for each locale in
// sync — missing keys fall back to `en` automatically.
import { init, register, waitLocale } from "svelte-i18n";

register("en", () => import("./locales/en.json"));
register("zh-CN", () => import("./locales/zh-CN.json"));

const FALLBACK_LOCALE = "en";

function detectLocale(): string {
  if (typeof navigator !== "undefined" && navigator.language) {
    return navigator.language;
  }
  return FALLBACK_LOCALE;
}

/**
 * Initialise svelte-i18n with the user's configured locale.
 *
 * `configLocale` is the value from `USER_CONFIG.locale` (populated by
 * `/api/config`). When unset we fall back to the browser's `navigator.language`,
 * then finally `en`. Unknown locales render via the English fallback bundle.
 *
 * Idempotent — safe to call again from the config page after the user changes
 * their locale.
 */
export function initI18n(configLocale?: string): void {
  init({
    fallbackLocale: FALLBACK_LOCALE,
    initialLocale: configLocale || detectLocale()
  });
}

export { _, locale, waitLocale } from "svelte-i18n";
