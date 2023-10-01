import _ from "lodash";
import {
  prefixMinutesSeconds,
  type Transaction,
  type TransactionSchedule,
  type TransactionSequence
} from "./utils";
import dayjs from "dayjs";
import { parse, type CronExprs } from "@datasert/cronjs-parser";
import { getFutureMatches } from "@datasert/cronjs-matcher";

const now = dayjs();
const end = dayjs().add(36, "month");

function zip(schedules: dayjs.Dayjs[], transactions: Transaction[], key: string, amount: number) {
  let si = 0;
  let ti = 0;
  const transactionSchedules: TransactionSchedule[] = [];

  while (si < schedules.length) {
    const s1 = schedules[si];
    const s2 = schedules[si + 1];
    const t1 = transactions[ti];
    const t2 = transactions[ti + 1];

    if (!t1) {
      transactionSchedules.push({
        key,
        amount,
        scheduled: s1,
        actual: null,
        transaction: null
      });
      si++;
      continue;
    }

    const t1s1diff = Math.abs(t1.date.diff(s1, "day"));
    let t1s2diff = Number.MAX_VALUE;
    if (s2) {
      t1s2diff = Math.abs(t1.date.diff(s2, "day"));
    }

    let t2s1diff = Number.MAX_VALUE;
    if (t2) {
      t2s1diff = Math.abs(t2.date.diff(s1, "day"));
    }

    if (t1s1diff > t2s1diff) {
      transactionSchedules.push({
        key,
        amount,
        scheduled: t1.date,
        actual: t1.date,
        transaction: t1
      });
      ti++;
    } else if (t1s1diff > t1s2diff) {
      transactionSchedules.push({
        key,
        amount,
        scheduled: s1,
        actual: null,
        transaction: null
      });
      si++;
    } else {
      transactionSchedules.push({
        key,
        amount,
        scheduled: s1,
        actual: t1.date,
        transaction: t1
      });
      si++;
      ti++;
    }
  }

  return transactionSchedules;
}

function enrich(ts: TransactionSequence) {
  const transactions = ts.transactions.slice().reverse();
  const amount = totalRecurring(ts);
  const start = transactions[0].date;
  let periodAvailable = false;
  let cron: CronExprs;
  try {
    if (ts.period != "") {
      cron = parse(prefixMinutesSeconds(ts.period), { hasSeconds: false });
      periodAvailable = true;
    } else {
      periodAvailable = false;
    }
  } catch (e) {
    periodAvailable = false;
  }

  if (periodAvailable) {
    const schedules = getFutureMatches(cron, {
      startAt: start.toISOString(),
      endAt: end.toISOString(),
      matchCount: 1000,
      timezone: dayjs.tz.guess()
    });

    ts.schedules = zip(
      _.map(schedules, (s) => dayjs(s)),
      transactions,
      ts.key,
      amount
    );
  } else {
    const schedules: dayjs.Dayjs[] = _.map(transactions, (t) => t.date);
    let next = _.last(schedules);
    do {
      next = nextDate(ts, next);
      schedules.push(next);
    } while (schedules.length < 1000 && end.isAfter(next));
    ts.schedules = zip(schedules, transactions, ts.key, amount);
  }

  const [past, future] = _.partition(ts.schedules, (s) => s.scheduled?.isBefore(now));
  ts.pastSchedules = past;
  ts.futureSchedules = future;
  ts.schedulesByMonth = _.groupBy(ts.schedules, (s) => s.scheduled?.format("YYYY-MM") || "NA");
  ts.interval = _.first(future).scheduled.diff(_.last(past).scheduled, "day");
  return ts;
}

export function nextUnpaidSchedule(ts: TransactionSequence) {
  const last = _.last(ts.pastSchedules);
  if (last && !last.actual) {
    return last;
  }
  return _.find(ts.futureSchedules, (s) => !s.actual);
}

export function scheduleIcon(schedule: TransactionSchedule) {
  let icon = "fa-circle-check";
  let glyph = "";
  let color = "has-text-success";
  let svgColor = "svg-text-success";

  if (!schedule.actual) {
    if (schedule.scheduled.isBefore(now, "day")) {
      color = "has-text-danger";
      icon = "fa-exclamation-triangle";
      glyph = "";
      svgColor = "svg-text-danger";
    } else {
      color = "has-text-grey";
      svgColor = "svg-text-grey";
    }
  } else {
    if (schedule.actual.isSameOrBefore(schedule.scheduled, "day")) {
      color = "has-text-success";
      svgColor = "svg-text-success";
    } else {
      color = "has-text-warning-dark";
      svgColor = "svg-text-warning-dark";
    }
  }

  return { icon, color, svgColor, glyph };
}

export function intervalText(ts: TransactionSequence) {
  if (ts.interval >= 7 && ts.interval <= 8) {
    return "weekly";
  }

  if (ts.interval >= 14 && ts.interval <= 16) {
    return "bi-weekly";
  }

  if (ts.interval >= 28 && ts.interval <= 33) {
    return "monthly";
  }

  if (ts.interval >= 87 && ts.interval <= 100) {
    return "quarterly";
  }

  if (ts.interval >= 175 && ts.interval <= 190) {
    return "half-yearly";
  }

  if (ts.interval >= 350 && ts.interval <= 395) {
    return "yearly";
  }

  return `every ${ts.interval} days`;
}

function nextDate(ts: TransactionSequence, date: dayjs.Dayjs) {
  if (ts.interval >= 28 && ts.interval <= 33) {
    return date.add(1, "month");
  }

  if (ts.interval >= 360 && ts.interval <= 370) {
    return date.add(1, "year");
  }

  return date.add(ts.interval, "day");
}

export function totalRecurring(ts: TransactionSequence) {
  const lastTransaction = ts.transactions[0];
  return _.sumBy(lastTransaction.postings, (t) => _.max([0, t.amount]));
}

export function enrichTrantionSequence(transactionSequences: TransactionSequence[]) {
  return _.map(transactionSequences, (ts) => enrich(ts));
}

export function sortTrantionSequence(transactionSequences: TransactionSequence[]) {
  return _.chain(transactionSequences)
    .sortBy((ts) => {
      return Math.abs(nextUnpaidSchedule(ts).scheduled.diff(now));
    })
    .value();
}
