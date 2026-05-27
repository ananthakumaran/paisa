export const trailingSlash = "never";

import type { LayoutLoad } from "./$types";
import { ajax, configUpdated, setNow } from "$lib/utils";
import { initI18n, waitLocale } from "$lib/i18n";
import { readPersistedLocale } from "$lib/i18n/toggle";

export const load = (async () => {
  const { config, now } = await ajax("/api/config");
  if (now) {
    setNow(now);
  }
  globalThis.USER_CONFIG = config;
  configUpdated();
  // Resolution order:
  //   1. localStorage (explicit user toggle) — always wins.
  //   2. `USER_CONFIG.locale` from the server.
  //   3. `navigator.language`.
  //   4. `en` fallback (`initI18n` defaults to this).
  // Await `waitLocale()` so the very first paint has its translation bundle
  // ready and we never flash raw message keys.
  let target: string | undefined = readPersistedLocale();
  if (!target) {
    target = config?.locale;
  }
  if (!target && typeof navigator !== "undefined" && navigator.language) {
    target = navigator.language;
  }
  initI18n(target);
  await waitLocale();
  return {};
}) satisfies LayoutLoad;
