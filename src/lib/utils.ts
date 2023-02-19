import chroma from "chroma-js";
import dayjs from "dayjs";
import { sprintf } from "sprintf-js";
import _ from "lodash";
import * as d3 from "d3";
import { loading } from "../store";

export interface Posting {
  id: string;
  date: dayjs.Dayjs;
  payee: string;
  account: string;
  commodity: string;
  quantity: number;
  amount: number;
  market_amount: number;

  timestamp: dayjs.Dayjs;
}

export interface Transaction {
  id: string;
  date: dayjs.Dayjs;
  payee: string;
  postings: Posting[];
}

export interface Price {
  id: string;
  date: string;
  commodity_type: string;
  commodity_id: string;
  commodity_name: string;
  value: number;

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

export interface InterestOverview {
  date: dayjs.Dayjs;
  drawn_amount: number;
  repaid_amount: number;
  interest_amount: number;
}

export interface Interest {
  account: string;
  overview_timeline: InterestOverview[];
  apr: number;
}

export interface AssetBreakdown {
  group: string;
  investment_amount: number;
  withdrawal_amount: number;
  balance_units: number;
  market_amount: number;
  xirr: number;
}

export interface LiabilityBreakdown {
  group: string;
  drawn_amount: number;
  repaid_amount: number;
  interest_amount: number;
  balance_amount: number;
  apr: number;
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

export interface PostingPair {
  purchase: Posting;
  sell: Posting;
  gain: number;
  taxable_gain: number;
  short_term_tax: number;
  long_term_tax: number;
}

export interface FYCapitalGain {
  gain: number;
  taxable_gain: number;
  short_term_tax: number;
  long_term_tax: number;
  units: number;
  purchase_price: number;
  sell_price: number;
  posting_pairs: PostingPair[];
}
export interface HarvestBreakdown {
  units: number;
  purchase_date: string;
  purchase_price: number;
  purchase_unit_price: number;
  current_price: number;
  unrealized_gain: number;
  taxable_unrealized_gain: number;
}

export interface Harvestable {
  account: string;
  tax_category: string;
  total_units: number;
  harvestable_units: number;
  unrealized_gain: number;
  taxable_unrealized_gain: number;
  current_unit_price: number;
  grandfather_unit_price: number;
  current_unit_date: string;
  harvest_breakdown: HarvestBreakdown[];
}

export interface CapitalGain {
  account: string;
  tax_category: string;
  fy: { [key: string]: FYCapitalGain };
}

export interface Issue {
  level: string;
  summary: string;
  description: string;
  details: string;
}

export interface ScheduleALSection {
  code: string;
  section: string;
  details: string;
}

export interface ScheduleALEntry {
  section: ScheduleALSection;
  amount: number;
}

export interface Point {
  date: dayjs.Dayjs;
  value: number;
}

export interface Forecast {
  date: dayjs.Dayjs;
  value: number;
  error: number;
}

export interface RetirementProgress {
  savings_total: number;
  savings_timeline: Point[];
  swr: number;
  yearly_expense: number;
}

export function ajax(route: "/api/retirement/progress"): Promise<RetirementProgress>;
export function ajax(route: "/api/harvest"): Promise<{ harvestables: Record<string, Harvestable> }>;
export function ajax(
  route: "/api/capital_gains"
): Promise<{ capital_gains: Record<string, CapitalGain> }>;
export function ajax(route: "/api/schedule_al"): Promise<{
  schedule_al_entries: ScheduleALEntry[];
  date: string;
}>;
export function ajax(route: "/api/diagnosis"): Promise<{ issues: Issue[] }>;
export function ajax(
  route: "/api/investment"
): Promise<{ assets: Posting[]; yearly_cards: YearlyCard[] }>;
export function ajax(route: "/api/ledger"): Promise<{ postings: Posting[] }>;
export function ajax(route: "/api/assets/balance"): Promise<{ asset_breakdowns: AssetBreakdown[] }>;
export function ajax(
  route: "/api/liabilities/balance"
): Promise<{ liability_breakdowns: LiabilityBreakdown[] }>;
export function ajax(route: "/api/price"): Promise<{ prices: Price[] }>;
export function ajax(route: "/api/transaction"): Promise<{ transactions: Transaction[] }>;
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
  month_wise: {
    expenses: { [key: string]: Posting[] };
    incomes: { [key: string]: Posting[] };
    investments: { [key: string]: Posting[] };
    taxes: { [key: string]: Posting[] };
  };
}>;

export function ajax(route: "/api/liabilities/interest"): Promise<{
  interest_timeline_breakdown: Interest[];
}>;

export async function ajax(route: "/api/sync", options?: RequestInit): Promise<any>;

export async function ajax(route: string, options?: RequestInit) {
  loading.set(true);
  const response = await fetch(route, options);
  const body = await response.text();
  loading.set(false);
  return JSON.parse(body, (key, value) => {
    if (
      _.isString(value) &&
      /date/.test(key) &&
      /^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$/.test(value)
    ) {
      return dayjs(value);
    }
    return value;
  });
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
  let precision = 2;
  if (x == Math.round(x)) {
    precision = 0;
  } else if (x * 10 == Math.round(x * 10)) {
    precision = 1;
  }
  return sprintf(`%.${precision}f %s`, x, unit);
}

export function formatFloat(value: number, precision = 2) {
  if (obscure) {
    return "00";
  }
  return sprintf(`%.${precision}f`, value);
}

export function formatPercentage(value: number, precision = 0) {
  if (obscure) {
    return "00";
  }

  if (!Number.isFinite(value)) {
    value = 0;
  }

  // minus 0
  if (1 / value === -Infinity) {
    value = 0;
  }

  return value.toLocaleString("hi", {
    style: "percent",
    minimumFractionDigits: precision
  });
}

export function formatFixedWidthFloat(value: number, width: number, precision = 2) {
  if (obscure) {
    value = 0;
  }
  return sprintf(`%${width}.${precision}f`, value);
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

export function firstName(account: string) {
  return _.first(account.split(":"));
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
  cb: (data: d3.AxisDomain, index: number) => string
) {
  const range = scale.range();
  const width = Math.abs(range[1] - range[0]);
  const s = scale as any;
  const points = s.ticks ? s.ticks().length : s.domain().length;
  return function (data: d3.AxisDomain, index: number) {
    let skip = Math.round((minWidth * points) / width);
    skip = Math.max(1, skip);

    return index % skip === 0 ? cb(data, index) : null;
  };
}

export function generateColorScheme(domain: string[]) {
  let colors: string[];

  const n = domain.length;
  if (n <= 12) {
    colors = {
      1: ["#7570b3"],
      2: ["#7fc97f", "#fdc086"],
      3: ["#66c2a5", "#fc8d62", "#8da0cb"],
      4: ["#66c2a5", "#fc8d62", "#8da0cb", "#e78ac3"],
      5: ["#66c2a5", "#fc8d62", "#8da0cb", "#e78ac3", "#a6d854"],
      6: ["#66c2a5", "#fc8d62", "#8da0cb", "#e78ac3", "#a6d854", "#ffd92f"],
      7: ["#8dd3c7", "#ffed6f", "#bebada", "#fb8072", "#80b1d3", "#fdb462", "#b3de69"],
      8: ["#8dd3c7", "#ffed6f", "#bebada", "#fb8072", "#80b1d3", "#fdb462", "#b3de69", "#fccde5"],
      9: [
        "#8dd3c7",
        "#ffed6f",
        "#bebada",
        "#fb8072",
        "#80b1d3",
        "#fdb462",
        "#b3de69",
        "#fccde5",
        "#d9d9d9"
      ],
      10: [
        "#8dd3c7",
        "#ffed6f",
        "#bebada",
        "#fb8072",
        "#80b1d3",
        "#fdb462",
        "#b3de69",
        "#fccde5",
        "#d9d9d9",
        "#bc80bd"
      ],
      11: [
        "#8dd3c7",
        "#ffed6f",
        "#bebada",
        "#fb8072",
        "#80b1d3",
        "#fdb462",
        "#b3de69",
        "#fccde5",
        "#d9d9d9",
        "#bc80bd",
        "#ccebc5"
      ],
      12: [
        "#8dd3c7",
        "#ffffb3",
        "#bebada",
        "#fb8072",
        "#80b1d3",
        "#fdb462",
        "#b3de69",
        "#fccde5",
        "#d9d9d9",
        "#bc80bd",
        "#ccebc5",
        "#ffed6f"
      ]
    }[n];
  } else {
    const z = d3
      .scaleSequential()
      .domain([0, n - 1])
      .interpolator(d3.interpolateSinebow);
    colors = _.map(_.range(0, n), (n) => chroma(z(n)).desaturate(1.5).hex());
  }

  return d3.scaleOrdinal<string>().domain(domain).range(colors);
}

export function setHtml(selector: string, value: string, color?: string) {
  const node: HTMLElement = document.querySelector(".d3-" + selector);
  if (color) {
    node.style.backgroundColor = color;
    node.style.padding = "5px";
    node.style.color = "white";
  }
  node.innerHTML = value;
}

export function rainbowScale(keys: string[]) {
  const x = d3
    .scaleLinear()
    .domain([0, _.size(keys) - 1])
    .range([0, 0.9]);
  return d3.scaleOrdinal(_.map(keys, (_value, i) => d3.interpolateRainbow(x(i)))).domain(keys);
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

export function tooltip(
  rows: Array<Array<string | string[]>>,
  options: {
    total?: string;
  } = {}
) {
  if (options.total && rows.length > 0) {
    const totalRow: Array<string | string[]> = [
      ["Total", "has-text-weight-bold"],
      [options.total, "has-text-weight-bold has-text-right"]
    ];

    for (let i = 2; i < rows[0].length; i++) {
      totalRow.unshift("");
    }

    rows.push(totalRow);
  }

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
  return `<table class='table is-narrow is-size-7 popup-table'><tbody>${trs}</tbody></table>`;
}

export function isMobile() {
  return window.innerWidth < 1024;
}
