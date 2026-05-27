import dayjs from "dayjs";

// Calendar-year (Jan–Dec) helpers. Replace the previous fiscal-year
// helpers after issue #5. Kept in a standalone module so the bun test
// runner can import them without pulling in the SvelteKit runtime
// (`$app/navigation`) which `./utils` transitively loads.

/**
 * Return the calendar year of a date as a plain 4-digit string, e.g.
 * "2024". Never returns a range like "2024-25".
 */
export function calendarYear(date: dayjs.Dayjs): string {
  return date.year().toString();
}

/**
 * Iterate one anchor per calendar year (Jan 1) covering the inclusive
 * range [start, end]. The callback receives a `dayjs` instance pinned
 * to Jan 1 of each year.
 */
export function forEachCalendarYear(
  start: dayjs.Dayjs,
  end: dayjs.Dayjs,
  cb?: (current: dayjs.Dayjs) => any
): dayjs.Dayjs[] {
  const years: dayjs.Dayjs[] = [];
  let current = start.startOf("year");
  const last = end.startOf("year");
  while (current.isSame(last, "year") || current.isBefore(last, "year")) {
    if (cb) {
      cb(current);
    }
    years.push(current);
    current = current.add(1, "year");
  }
  return years;
}
