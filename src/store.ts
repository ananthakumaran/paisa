import { persisted } from "svelte-local-storage-store";
import { writable, derived, get } from "svelte/store";
import * as d3 from "d3";

import type { AccountTfIdf } from "$lib/utils";
import dayjs from "dayjs";

export const month = writable(dayjs().format("YYYY-MM"));
export const year = writable<string>("");
export const dateRangeOption = writable<number>(3);

export const dateMin = writable(dayjs("1980", "YYYY"));
export const dateMax = writable(dayjs());

export const dateRange = derived(
  [dateMin, dateMax, dateRangeOption],
  ([$dateMin, $dateMax, $dateRangeOption]) => {
    if ($dateRangeOption === -1) {
      return { from: $dateMin, to: $dateMax };
    } else {
      return {
        from: $dateMax.subtract($dateRangeOption, "year"),
        to: $dateMax
      };
    }
  }
);

export const cashflowType = persisted("cashflowType", "hierarchy");

export const cashflowExpenseDepthAllowed = writable({ min: 1, max: 1 });
export const cashflowExpenseDepth = persisted("cashflowExpenseDepth", 0);
export const cashflowIncomeDepthAllowed = writable({ min: 1, max: 1 });
export const cashflowIncomeDepth = persisted("cashflowIncomeDepth", 0);

export function setCashflowDepthAllowed(expense: number, income: number) {
  cashflowExpenseDepthAllowed.set({ min: 1, max: expense });
  if (get(cashflowExpenseDepth) == 0 || get(cashflowExpenseDepth) > expense) {
    cashflowExpenseDepth.set(expense);
  }

  cashflowIncomeDepthAllowed.set({ min: 1, max: income });
  if (get(cashflowIncomeDepth) == 0 || get(cashflowIncomeDepth) > income) {
    cashflowIncomeDepth.set(income);
  }
}

export const theme = writable("light");

export const loading = writable(false);

let timeoutId: NodeJS.Timeout;
export const delayedLoading = derived([loading], ([$l], set) => {
  if (timeoutId) {
    clearTimeout(timeoutId);
  }

  if (!$l) {
    set($l);
  } else {
    timeoutId = setTimeout(() => {
      return set($l);
    }, 200);
  }
});

export const willClearTippy = writable(0);

export const accountTfIdf = writable<AccountTfIdf>(null);

export function setAllowedDateRange(dates: dayjs.Dayjs[]) {
  const [start, end] = d3.extent(dates);
  if (start) {
    dateMin.set(start);
    dateMax.set(end);
  }
}

export const willRefresh = writable(0);
export function refresh() {
  willRefresh.update((n) => n + 1);
}
