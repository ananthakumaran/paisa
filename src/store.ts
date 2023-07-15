import { writable } from "svelte/store";

import type { AccountTfIdf } from "$lib/utils";
import dayjs from "dayjs";

export const month = writable(dayjs().format("YYYY-MM"));
export const year = writable<string>("");

export const dateMin = writable(dayjs("1980", "YYYY"));
export const dateMax = writable(dayjs());

export const loading = writable(false);

export const willClearTippy = writable(0);

export const accountTfIdf = writable<AccountTfIdf>(null);
