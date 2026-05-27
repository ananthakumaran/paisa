export const trailingSlash = "never";

import type { LayoutLoad } from "./$types";
import { ajax, configUpdated, setNow } from "$lib/utils";
import { initI18n, waitLocale } from "$lib/i18n";

export const load = (async () => {
  const { config, now } = await ajax("/api/config");
  if (now) {
    setNow(now);
  }
  globalThis.USER_CONFIG = config;
  configUpdated();
  // Boot svelte-i18n with the configured locale (falls back to navigator
  // language, then `en`). Await `waitLocale()` so the very first paint has
  // its translation bundle ready and we never flash raw message keys.
  initI18n(config?.locale);
  await waitLocale();
  return {};
}) satisfies LayoutLoad;
