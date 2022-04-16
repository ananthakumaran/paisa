import dayjs from "dayjs";
import isSameOrBefore from "dayjs/plugin/isSameOrBefore";
import { sprintf } from "sprintf-js";
import _ from "lodash";
import * as d3 from "d3";

export interface Posting {
  id: string;
  date: string;
  payee: string;
  account: string;
  commodity: string;
  quanity: number;
  amount: number;
  market_amount: number;

  timestamp: dayjs.Dayjs;
}

export interface Networth {
  date: string;
  actual: number;
  gain: number;

  timestamp: dayjs.Dayjs;
}

export interface Breakdown {
  group: string;
  amount: number;
  market_amount: number;
  xirr: number;
}

export interface Aggregate {
  date: string;
  account: string;
  amount: number;
  market_amount: number;

  timestamp: dayjs.Dayjs;
}

export function ajax(
  route: "/api/investment"
): Promise<{ postings: Posting[] }>;
export function ajax(
  route: "/api/ledger"
): Promise<{ postings: Posting[]; breakdowns: Breakdown[] }>;
export function ajax(
  route: "/api/overview"
): Promise<{ networth_timeline: Networth[] }>;
export function ajax(route: "/api/allocation"): Promise<{
  aggregates: { [key: string]: Aggregate };
  aggregates_timeline: { [key: string]: Aggregate }[];
}>;
export async function ajax(route: string) {
  const response = await fetch(route);
  return await response.json();
}

const obscure = false;

export function formatCurrency(value: number) {
  if (obscure) {
    return "00";
  }

  return Math.round(value).toLocaleString("hi");
}

export function formatCurrencyCrude(value: number) {
  if (obscure) {
    return "00";
  }

  let x = 0,
    unit = "";
  if (Math.abs(value) < 100000) {
    (x = value / 1000), (unit = "K");
  } else if (Math.abs(value) <= 10000000) {
    (x = value / 100000), (unit = "L");
  } else {
    (x = value / 10000000), (unit = "C");
  }
  const precision = 2;
  return sprintf(`%.${precision}f %s`, x, unit);
}

export function formatFloat(value) {
  if (obscure) {
    return "00";
  }
  return sprintf("%.2f", value);
}

export function forEachMonth(
  start: dayjs.Dayjs,
  end: dayjs.Dayjs,
  cb: (current: dayjs.Dayjs) => any
) {
  let current = start;
  while (current.isSameOrBefore(end, "month")) {
    cb(current);
    current = current.add(1, "month");
  }
}

export function lastName(account: string) {
  return _.last(account.split(":"));
}

export function secondName(account: string) {
  return account.split(":")[1];
}

export function parentName(account: string) {
  return _.dropRight(account.split(":"), 1).join(":");
}

export function skipTicks(
  minWidth: number,
  width: number,
  points: number,
  cb: (data: d3.AxisDomain, index: number) => string
) {
  return function (data: d3.AxisDomain, index: number) {
    let skip = Math.round((minWidth * points) / width);
    skip = Math.max(1, skip);

    return index % skip === 0 ? cb(data, index) : null;
  };
}

export function setHtml(selector: string, value: string) {
  var node = document.querySelector(".d3-" + selector);
  node.innerHTML = value;
}

export function rainbowScale(keys: string[]) {
  const x = d3
    .scaleLinear()
    .domain([0, _.size(keys) - 1])
    .range([0, 0.9]);
  return d3
    .scaleOrdinal(_.map(keys, (_value, i) => d3.interpolateRainbow(x(i))))
    .domain(keys);
}
