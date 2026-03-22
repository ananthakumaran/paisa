export const trailingSlash = "never";

import type { LayoutLoad } from "./$types";
import { ajax, configUpdated, setNow } from "$lib/utils";

export const load = (async () => {
  const { config, now } = await ajax("/api/config", { background: true });
  const enc = await ajax("/api/encryption/status", { background: true });
  if (now) {
    setNow(now);
  }
  globalThis.USER_CONFIG = config;
  configUpdated();
  return {
    encryptionLocked: enc.needs_unlock === true
  };
}) satisfies LayoutLoad;
