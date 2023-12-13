import type { LayoutLoad } from "./$types";

import { ajax, configUpdated, setNow } from "$lib/utils";

import { goto } from "$app/navigation";

export const load = (async ({}) => {
  try {
    const { config, now } = await ajax("/api/config");
    if (now) {
      setNow(now);
    }
    globalThis.USER_CONFIG = config;
    configUpdated();
    return {};
  } catch (e) {
    globalThis.USER_CONFIG = {} as any;
    await goto("/login");
  }
}) satisfies LayoutLoad;
