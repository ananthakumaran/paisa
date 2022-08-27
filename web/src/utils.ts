import dayjs from "dayjs";
import { sprintf } from "sprintf-js";
import _ from "lodash";
import * as d3 from "d3";

export interface Posting {
  id: string;
  date: string;
  payee: string;
  account: string;
  commodity: string;
  quantity: number;
  amount: number;
  market_amount: number;

  timestamp: dayjs.Dayjs;
}

export interface Overview {
  date: string;
  investment_amount: number;
  withdrawal_amount: number;
  gain_amount: number;

  timestamp: dayjs.Dayjs;
}

export interface Gain {
  account: string;
  overview_timeline: Overview[];
  xirr: number;
}

export interface Breakdown {
  group: string;
  investment_amount: number;
  withdrawal_amount: number;
  balance_units: number;
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

export interface AllocationTarget {
  name: string;
  target: number;
  current: number;
  aggregates: { [key: string]: Aggregate };
}

export interface Income {
  date: string;
  postings: Posting[];

  timestamp: dayjs.Dayjs;
}

export interface Tax {
  start_date: string;
  end_date: string;
  postings: Posting[];
}

export interface YearlyCard {
  start_date: string;
  end_date: string;
  postings: Posting[];
  net_tax: number;
  gross_salary_income: number;
  gross_other_income: number;
  net_income: number;
  net_investment: number;
  net_expense: number;

  start_date_timestamp: dayjs.Dayjs;
  end_date_timestamp: dayjs.Dayjs;
}

export function ajax(
  route: "/api/investment"
): Promise<{ assets: Posting[]; yearly_cards: YearlyCard[] }>;
export function ajax(
  route: "/api/ledger"
): Promise<{ postings: Posting[]; breakdowns: Breakdown[] }>;
export function ajax(route: "/api/overview"): Promise<{
  overview_timeline: Overview[];
  xirr: number;
}>;
export function ajax(route: "/api/gain"): Promise<{
  gain_timeline_breakdown: Gain[];
}>;
export function ajax(route: "/api/allocation"): Promise<{
  aggregates: { [key: string]: Aggregate };
  aggregates_timeline: { [key: string]: Aggregate }[];
  allocation_targets: AllocationTarget[];
}>;
export function ajax(route: "/api/income"): Promise<{
  income_timeline: Income[];
  tax_timeline: Tax[];
}>;
export function ajax(route: "/api/expense"): Promise<{
  expenses: Posting[];
  current_month: {
    expenses: Posting[];
    incomes: Posting[];
    investments: Posting[];
    taxes: Posting[];
  };
}>;
export async function ajax(route: string) {
  const response = await fetch(route);
  return await response.json();
}

const obscure = false;

export function formatCurrency(value: number, precision = 0) {
  if (obscure) {
    return "00";
  }

  // minus 0
  if (1 / value === -Infinity) {
    value = 0;
  }

  return value.toLocaleString("hi", {
    minimumFractionDigits: precision,
    maximumFractionDigits: precision
  });
}

export function formatCurrencyCrude(value: number) {
  if (obscure) {
    return "00";
  }

  let x = 0,
    unit = "";
  if (Math.abs(value) < 100000) {
    (x = value / 1000), (unit = "K");
  } else if (Math.abs(value) < 10000000) {
    (x = value / 100000), (unit = "L");
  } else {
    (x = value / 10000000), (unit = "C");
  }
  const precision = 2;
  return sprintf(`%.${precision}f %s`, x, unit);
}

export function formatFloat(value, precision = 2) {
  if (obscure) {
    return "00";
  }
  return sprintf(`%.${precision}f`, value);
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

export function forEachYear(
  start: dayjs.Dayjs,
  end: dayjs.Dayjs,
  cb: (current: dayjs.Dayjs) => any
) {
  let current = start;
  while (current.isSameOrBefore(end, "year")) {
    cb(current);
    current = current.add(1, "year");
  }
}

export function lastName(account: string) {
  return _.last(account.split(":"));
}

export function secondName(account: string) {
  return account.split(":")[1];
}

export function restName(account: string) {
  return _.drop(account.split(":")).join(":");
}

export function parentName(account: string) {
  return _.dropRight(account.split(":"), 1).join(":");
}

export function depth(account: string) {
  return account.split(":").length;
}

export function skipTicks<Domain>(
  minWidth: number,
  scale: d3.AxisScale<Domain>,
  cb: (data: d3.AxisDomain, index: number) => string,
  ticks?: number
) {
  const range = scale.range();
  const width = Math.abs(range[1] - range[0]);
  const points = ticks ? ticks : (scale as any).ticks().length;
  return function (data: d3.AxisDomain, index: number) {
    let skip = Math.round((minWidth * points) / width);
    skip = Math.max(1, skip);

    return index % skip === 0 ? cb(data, index) : null;
  };
}

export function setHtml(selector: string, value: string) {
  const node = document.querySelector(".d3-" + selector);
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

export function textColor(backgroundColor: string) {
  const color = d3.rgb(backgroundColor);
  // http://www.w3.org/TR/AERT#color-contrast
  const brightness = (color.r * 299 + color.g * 587 + color.b) / 1000;
  if (brightness > 125) {
    return "black";
  }
  return "white";
}

export function tooltip(rows: Array<Array<string | [string, string]>>) {
  const trs = rows
    .map((r) => {
      const cells = r
        .map((c) => {
          if (typeof c == "string") {
            return `<td>${c}</td>`;
          } else {
            return `<td class='${c[1]}'>${c[0]}</td>`;
          }
        })
        .join("\n");

      return `<tr>${cells}</tr>`;
    })
    .join("\n");
  return `<table class='table is-narrow is-size-7'><tbody>${trs}</tbody></table>`;
}
