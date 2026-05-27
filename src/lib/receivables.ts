// View-model layer for the /api/receivables endpoint.
//
// Mirrors internal/server/receivables.go::Receivable. Dates come over the
// wire as ISO-8601 strings; `ajax()` in $lib/utils auto-revives any key
// matching /Date|date|time|now/, so by the time the page receives the
// payload `lend_date` and `due_date` are either `null` or a `dayjs.Dayjs`
// instance.

import dayjs from "dayjs";
import _ from "lodash";

export interface Receivable {
  account: string;
  borrower: string;
  outstanding: number;
  lend_date: dayjs.Dayjs | null;
  due_date: dayjs.Dayjs | null;
  interest_rate: number;
  note: string;
  kind: string;
}

export interface ReceivablesResponse {
  receivables: Receivable[];
  total_outstanding: number;
}

/**
 * isOverdue returns true when a receivable has a `due_date` strictly in
 * the past relative to `now`. The function deliberately treats a missing
 * `due_date` as "not overdue" so accounts without an agreed return date
 * never get the red treatment.
 *
 * Exposed for testing.
 */
export function isOverdue(r: Receivable, now: dayjs.Dayjs = dayjs()): boolean {
  if (!r.due_date) {
    return false;
  }
  return r.due_date.isBefore(now, "day");
}

/**
 * sortByOutstandingDesc returns a new array sorted by `outstanding` in
 * descending order. The backend already returns rows in this order but
 * tabulator can re-sort on user input, so the helper is exposed for
 * tests and for the default initialSort config.
 */
export function sortByOutstandingDesc(rs: Receivable[]): Receivable[] {
  return _.orderBy(rs, [(r) => r.outstanding, (r) => r.account], ["desc", "asc"]);
}
