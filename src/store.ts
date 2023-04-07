import { writable } from "svelte/store";
import dayjs from "dayjs";
import { financialYear } from "$lib/utils";

export const month = writable(dayjs().format("YYYY-MM"));
export const year = writable(financialYear(dayjs()));

export const dateMin = writable(dayjs("1980", "YYYY"));
export const dateMax = writable(dayjs());

export const loading = writable(false);
