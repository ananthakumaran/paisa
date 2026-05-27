export const prerender = false;
export const ssr = false;
export const trailingSlash = "never";

import "../common.scss";
import "../light.scss";
import "../dark.scss";

import type { LayoutLoad } from "./$types";
import { initI18n, waitLocale } from "$lib/i18n";

// Boot svelte-i18n early so error pages (`+error.svelte`) and routes outside
// the `(app)` group (e.g. `/login`) render translated strings on first paint.
// Without `USER_CONFIG.locale` yet, we fall back to `navigator.language` /
// `en`; the `(app)/+layout.ts` re-inits with the configured locale once
// `/api/config` resolves (the call is idempotent).
export const load = (async () => {
  initI18n();
  await waitLocale();
  return {};
}) satisfies LayoutLoad;
