import { writable } from "svelte/store";
import dayjs from "dayjs";

export const month = writable(dayjs().format("YYYY-MM"));

export const loading = writable(false);
