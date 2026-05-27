export const prerender = false;
export const ssr = false;
export const trailingSlash = "never";

import "../common.scss";
import "../light.scss";
import "../dark.scss";

import type { LayoutLoad } from "./$types";
import { initI18n, waitLocale } from "$lib/i18n";
import { readPersistedLocale } from "$lib/i18n/toggle";

// Boot svelte-i18n early so error pages (`+error.svelte`) and routes outside
// the `(app)` group (e.g. `/login`) render translated strings on first paint.
//
// Locale resolution order:
//   1. localStorage (`paisa.locale`) — the user's explicit toggle choice
//      ALWAYS wins, no matter what the server config says.
//   2. `navigator.language` — best guess from the browser.
//   3. `en` (handled by `initI18n` internally).
//
// The `(app)/+layout.ts` re-inits with `USER_CONFIG.locale` once
// `/api/config` resolves, but the persisted choice still takes priority
// there too — the user toggle is the source of truth.
export const load = (async () => {
  let initial: string | undefined = readPersistedLocale();
  if (!initial && typeof navigator !== "undefined" && navigator.language) {
    initial = navigator.language;
  }
  initI18n(initial);
  await waitLocale();
  return {};
}) satisfies LayoutLoad;
