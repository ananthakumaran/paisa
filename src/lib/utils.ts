import dayjs from "dayjs";
import { sprintf } from "sprintf-js";
import _ from "lodash";
import * as d3 from "d3";
import { loading } from "../store";
import type { JSONSchema7 } from "json-schema";

export interface AutoCompleteItem {
  label: string;
  id: string;
}

export interface AutoCompleteField {
  id: string;
  label: string;
  help: string;
  inputType: string;
}

export interface PriceProvider {
  code: string;
  fields: AutoCompleteField[];
  label: string;
  description: string;
}

export interface Posting {
  id: string;
  date: dayjs.Dayjs;
  payee: string;
  account: string;
  commodity: string;
  quantity: number;
  amount: number;
  market_amount: number;
  status: string;
  tag_recurring: string;
  transaction_begin_line: number;
  transaction_end_line: number;
  file_name: string;
}

export interface CashFlow {
  date: dayjs.Dayjs;
  income: number;
  liabilities: number;
  expenses: number;
  investment: number;
  tax: number;
  checking: number;
  balance: number;
}

export interface TransactionSequenceKey {
  tagRecurring: string;
}

export interface TransactionSequence {
  transactions: Transaction[];
  key: TransactionSequenceKey;
  interval: number;
}

export interface Transaction {
  id: string;
  date: dayjs.Dayjs;
  payee: string;
  beginLine: number;
  endLine: number;
  fileName: string;
  postings: Posting[];
}

export interface Price {
  id: string;
  date: dayjs.Dayjs;
  commodity_type: string;
  commodity_id: string;
  commodity_name: string;
  value: number;
}

export interface Networth {
  date: dayjs.Dayjs;
  investmentAmount: number;
  withdrawalAmount: number;
  gainAmount: number;
  balanceAmount: number;
  netInvestmentAmount: number;
}

export interface Gain {
  account: string;
  networth: Networth;
  xirr: number;
  postings: Posting[];
}

export interface AccountGain {
  account: string;
  networthTimeline: Networth[];
  xirr: number;
  postings: Posting[];
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

export interface AccountTfIdf {
  tf_idf: Record<string, Record<string, number>>;
  index: {
    docs: Record<string, Record<string, number>>;
    tokens: Record<string, Record<string, number>>;
  };
}

export interface AssetBreakdown {
  group: string;
  investmentAmount: number;
  withdrawalAmount: number;
  balanceUnits: number;
  marketAmount: number;
  xirr: number;
  gainAmount: number;
  absoluteReturn: number;
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
  date: dayjs.Dayjs;
  account: string;
  amount: number;
  market_amount: number;
}

export interface CommodityBreakdown {
  commodity_name: string;
  security_name: string;
  security_id: string;
  security_type: string;
  amount: number;
  percentage: number;
}

export interface PortfolioAllocation {
  name_and_security_type: PortfolioAggregate[];
  security_type: PortfolioAggregate[];
  rating: PortfolioAggregate[];
  industry: PortfolioAggregate[];
  commodities: string[];
}

export interface PortfolioAggregate {
  id: string;
  group: string;
  sub_group: string;
  amount: number;
  percentage: number;
  breakdowns: CommodityBreakdown[];
}

export interface AllocationTarget {
  name: string;
  target: number;
  current: number;
  aggregates: { [key: string]: Aggregate };
}

export interface Income {
  date: dayjs.Dayjs;
  postings: Posting[];
}

export interface Tax {
  start_date: string;
  end_date: string;
  postings: Posting[];
}

export interface InvestmentYearlyCard {
  start_date: dayjs.Dayjs;
  end_date: dayjs.Dayjs;
  postings: Posting[];
  net_tax: number;
  gross_salary_income: number;
  gross_other_income: number;
  net_income: number;
  net_investment: number;
  net_expense: number;
  savings_rate: number;
}

export interface IncomeYearlyCard {
  start_date: dayjs.Dayjs;
  end_date: dayjs.Dayjs;
  postings: Posting[];
  net_tax: number;
  gross_income: number;
  net_income: number;
}

export interface Tax {
  gain: number;
  taxable: number;
  short_term: number;
  long_term: number;
  slab: number;
}

export interface PostingPair {
  purchase: Posting;
  sell: Posting;
  tax: Tax;
}

export interface FYCapitalGain {
  tax: Tax;
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
  tax: Tax;
}

export interface Harvestable {
  account: string;
  tax_category: string;
  total_units: number;
  harvestable_units: number;
  unrealized_gain: number;
  taxable_unrealized_gain: number;
  current_unit_price: number;
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

export interface ScheduleAL {
  entries: ScheduleALEntry[];
  date: dayjs.Dayjs;
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

export interface Budget {
  date: dayjs.Dayjs;
  accounts: AccountBudget[];
  endOfMonthBalance: number;
  availableThisMonth: number;
  forecast: number;
}

export interface AccountBudget {
  account: string;
  date: dayjs.Dayjs;
  actual: number;
  budgeted: number;
  forecast: number;
  available: number;
  rollover: number;
  expenses: Posting[];
}

export interface RetirementProgress {
  savings_total: number;
  savings_timeline: Point[];
  swr: number;
  yearly_expense: number;
  xirr: number;
}

export interface LedgerFile {
  type: "file";
  name: string;
  content: string;
  versions: string[];
}

export interface LedgerDirectory {
  type: "directory";
  name: string;
  children: Array<LedgerDirectory | LedgerFile>;
}

export function buildLedgerTree(files: LedgerFile[]) {
  const root: LedgerDirectory = {
    type: "directory",
    name: "",
    children: []
  };

  for (const file of _.sortBy(files, (f) => f.name)) {
    const parts = file.name.split("/");
    let current = root;
    for (const part of _.dropRight(parts, 1)) {
      let found = current.children.find((c) => c.name === part);
      if (!found) {
        found = {
          type: "directory",
          name: part,
          children: []
        };
        current.children.push(found);
      }
      current = found as LedgerDirectory;
    }
    current.children.push(file);
  }

  return root.children;
}

export interface LedgerFileError {
  line_from: number;
  line_to: number;
  error: string;
  message: string;
}

export interface Node {
  id: number;
  name: string;
}

export interface Link {
  source: number;
  target: number;
  value: number;
}

export interface Graph {
  nodes: Node[];
  links: Link[];
}

export interface Template {
  id: number;
  name: string;
  content: string;
  template_type: string;
}

export interface Log {
  time: dayjs.Dayjs;
  level: string;
  msg: string;
}

const BACKGROUND = [
  "/api/editor/validate",
  "/api/editor/save",
  "/api/editor/file",
  "/api/editor/files",
  "/api/editor/file/delete_backups",
  "/api/templates",
  "/api/templates/upsert",
  "/api/templates/delete",
  "/api/price/autocomplete",
  "/api/price/providers/delete/:provider",
  "/api/price/providers"
];

export function ajax(route: "/api/config"): Promise<{ config: UserConfig; schema: JSONSchema7 }>;
export function ajax(route: "/api/retirement/progress"): Promise<RetirementProgress>;
export function ajax(route: "/api/harvest"): Promise<{ harvestables: Record<string, Harvestable> }>;
export function ajax(
  route: "/api/capital_gains"
): Promise<{ capital_gains: Record<string, CapitalGain> }>;
export function ajax(route: "/api/schedule_al"): Promise<{
  schedule_als: Record<string, ScheduleAL>;
}>;
export function ajax(route: "/api/diagnosis"): Promise<{ issues: Issue[] }>;
export function ajax(route: "/api/logs"): Promise<{ logs: Log[] }>;
export function ajax(
  route: "/api/investment"
): Promise<{ assets: Posting[]; yearly_cards: InvestmentYearlyCard[] }>;
export function ajax(route: "/api/ledger"): Promise<{ postings: Posting[] }>;
export function ajax(route: "/api/assets/balance"): Promise<{ asset_breakdowns: AssetBreakdown[] }>;
export function ajax(route: "/api/liabilities/repayment"): Promise<{ repayments: Posting[] }>;
export function ajax(
  route: "/api/liabilities/balance"
): Promise<{ liability_breakdowns: LiabilityBreakdown[] }>;
export function ajax(route: "/api/price"): Promise<{ prices: Record<string, Price[]> }>;
export function ajax(route: "/api/transaction"): Promise<{ transactions: Transaction[] }>;
export function ajax(route: "/api/networth"): Promise<{
  networthTimeline: Networth[];
  xirr: number;
}>;
export function ajax(route: "/api/gain"): Promise<{
  gain_breakdown: Gain[];
}>;
export function ajax(route: "/api/dashboard"): Promise<{
  expenses: { [key: string]: Posting[] };
  cashFlows: CashFlow[];
  transactionSequences: TransactionSequence[];
  networth: { networth: Networth; xirr: number };
  transactions: Transaction[];
  budget: {
    budgetsByMonth: { [key: string]: Budget };
  };
}>;

export function ajax(
  route: "/api/gain/:name",
  options?: RequestInit,
  params?: Record<string, string>
): Promise<{
  gain_timeline_breakdown: AccountGain;
  portfolio_allocation: PortfolioAllocation;
  asset_breakdown: AssetBreakdown;
}>;

export function ajax(route: "/api/allocation"): Promise<{
  aggregates: { [key: string]: Aggregate };
  aggregates_timeline: { [key: string]: Aggregate }[];
  allocation_targets: AllocationTarget[];
}>;
export function ajax(route: "/api/portfolio_allocation"): Promise<PortfolioAllocation>;
export function ajax(route: "/api/income"): Promise<{
  income_timeline: Income[];
  tax_timeline: Tax[];
  yearly_cards: IncomeYearlyCard[];
}>;
export function ajax(route: "/api/expense"): Promise<{
  expenses: Posting[];
  month_wise: {
    expenses: { [key: string]: Posting[] };
    incomes: { [key: string]: Posting[] };
    investments: { [key: string]: Posting[] };
    taxes: { [key: string]: Posting[] };
  };
  year_wise: {
    expenses: { [key: string]: Posting[] };
    incomes: { [key: string]: Posting[] };
    investments: { [key: string]: Posting[] };
    taxes: { [key: string]: Posting[] };
  };
  graph: { [key: string]: { flat: Graph; hierarchy: Graph } };
}>;

export function ajax(route: "/api/budget"): Promise<{
  budgetsByMonth: { [key: string]: Budget };
  checkingBalance: number;
  availableForBudgeting: number;
}>;

export function ajax(route: "/api/cash_flow"): Promise<{ cash_flows: CashFlow[] }>;

export function ajax(
  route: "/api/recurring"
): Promise<{ transaction_sequences: TransactionSequence[] }>;

export function ajax(route: "/api/liabilities/interest"): Promise<{
  interest_timeline_breakdown: Interest[];
}>;

export function ajax(route: "/api/account/tf_idf"): Promise<AccountTfIdf>;
export function ajax(route: "/api/templates"): Promise<{ templates: Template[] }>;
export function ajax(route: "/api/templates/upsert", options?: RequestInit): Promise<Template>;
export function ajax(route: "/api/templates/delete", options?: RequestInit): Promise<void>;

export function ajax(route: "/api/editor/files"): Promise<{
  files: LedgerFile[];
  accounts: string[];
  commodities: string[];
  payees: string[];
}>;

export function ajax(
  route: "/api/editor/validate",
  options?: RequestInit
): Promise<{ errors: LedgerFileError[]; output: string }>;

export function ajax(
  route: "/api/editor/save",
  options?: RequestInit
): Promise<{ errors: LedgerFileError[]; saved: boolean; file: LedgerFile; message: string }>;

export function ajax(
  route: "/api/editor/file",
  options?: RequestInit
): Promise<{ file: LedgerFile }>;

export function ajax(
  route: "/api/editor/file/delete_backups",
  options?: RequestInit
): Promise<{ file: LedgerFile }>;

export function ajax(route: "/api/sync", options?: RequestInit): Promise<any>;
export function ajax(
  route: "/api/price/providers",
  options?: RequestInit
): Promise<{ providers: PriceProvider[] }>;

export function ajax(
  route: "/api/price/providers/delete/:provider",
  options?: RequestInit,
  params?: Record<string, string>
): Promise<{
  gain_timeline_breakdown: AccountGain;
  portfolio_allocation: PortfolioAllocation;
  asset_breakdown: AssetBreakdown;
}>;

export function ajax(
  route: "/api/price/autocomplete",
  options?: RequestInit
): Promise<{ completions: AutoCompleteItem[] }>;
export function ajax(route: "/api/init", options?: RequestInit): Promise<any>;

export function ajax(
  route: "/api/config",
  options?: RequestInit
): Promise<{ success: boolean; error?: string }>;

export async function ajax(route: string, options?: RequestInit, params?: Record<string, string>) {
  if (!_.includes(BACKGROUND, route)) {
    loading.set(true);
  }

  if (!_.isEmpty(params)) {
    _.each(params, (value, key) => {
      route = route.replace(`:${key}`, value);
    });
  }

  options = options || {};

  options.headers = {
    "Content-Type": "application/json"
  };

  const response = await fetch(route, options);
  const body = await response.text();
  if (!_.includes(BACKGROUND, route)) {
    loading.set(false);
  }
  return JSON.parse(body, (key, value) => {
    if (
      _.isString(value) &&
      /date|time/.test(key) &&
      /^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(.[0-9]+)?(Z|[+-][0-9]{2}:[0-9]{2})$/.test(
        value
      )
    ) {
      return dayjs(value);
    }
    return value;
  });
}

function obscure() {
  return localStorage.getItem("obscure") == "true";
}

export function formatCurrency(value: number, precision = 0) {
  if (obscure()) {
    return "00";
  }

  // minus 0
  if (1 / value === -Infinity) {
    value = 0;
  }

  return value.toLocaleString(USER_CONFIG.locale, {
    minimumFractionDigits: precision,
    maximumFractionDigits: precision
  });
}

export function formatCurrencyCrude(value: number) {
  return formatCurrencyCrudeWithPrecision(value, -1);
}

export function formatCurrencyCrudeWithPrecision(value: number, precision: number) {
  if (obscure()) {
    return "00";
  }

  const options: Intl.NumberFormatOptions = {
    notation: "compact"
  };

  if (precision < 0) {
    options.maximumFractionDigits = 2;
  } else {
    options.maximumFractionDigits = precision;
    options.minimumFractionDigits = precision;
  }

  return value.toLocaleString(USER_CONFIG.locale, options);
}

export function formatFloat(value: number, precision = 2) {
  if (obscure()) {
    return "00";
  }
  return sprintf(`%.${precision}f`, value);
}

export function formatPercentage(value: number, precision = 0) {
  if (obscure()) {
    return "00";
  }

  if (!Number.isFinite(value)) {
    value = 0;
  }

  // minus 0
  if (1 / value === -Infinity) {
    value = 0;
  }

  return value.toLocaleString(USER_CONFIG.locale, {
    style: "percent",
    minimumFractionDigits: precision
  });
}

export function formatFixedWidthFloat(value: number, width: number, precision = 2) {
  if (obscure()) {
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

export function forEachFinancialYear(
  start: dayjs.Dayjs,
  end: dayjs.Dayjs,
  cb?: (current: dayjs.Dayjs) => any
) {
  let current = start;
  if (current.month() < 3) {
    current = current.year(current.year() - 1);
  }
  current = current.month(3).date(1);

  const years: dayjs.Dayjs[] = [];
  while (current.isSameOrBefore(end, "month")) {
    if (cb) {
      cb(current);
    }
    years.push(current);
    current = current.add(1, "year");
  }
  return years;
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
    header?: string;
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

  if (options.header && rows.length > 0) {
    const headerRow: Array<string | string[]> = [
      [options.header, "has-text-weight-bold has-text-centered", rows[0].length.toString()]
    ];
    rows.unshift(headerRow);
  }

  const trs = rows
    .map((r) => {
      const cells = r
        .map((c) => {
          if (typeof c == "string") {
            return `<td>${c}</td>`;
          } else {
            if (c.length == 3) {
              return `<td class='${c[1]}' colspan='${c[2]}'>${c[0]}</td>`;
            }
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

export function financialYear(date: dayjs.Dayjs) {
  if (USER_CONFIG.financial_year_starting_month == 1) {
    return date.year().toString();
  }

  if (date.month() < USER_CONFIG.financial_year_starting_month - 1) {
    return `${date.year() - 1} - ${(date.year() % 100).toLocaleString("en-US", {
      minimumIntegerDigits: 2
    })}`;
  } else {
    return `${date.year()} - ${((date.year() + 1) % 100).toLocaleString("en-US", {
      minimumIntegerDigits: 2
    })}`;
  }
}

export function helpUrl(section: string) {
  return `https://paisa.fyi/reference/${section}`;
}

export function postingUrl(posting: Posting) {
  return `/ledger/editor/${encodeURIComponent(posting.file_name)}#${
    posting.transaction_begin_line
  }`;
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

export function nextDate(ts: TransactionSequence) {
  const lastTransaction = ts.transactions[0];
  if (ts.interval >= 28 && ts.interval <= 33) {
    return lastTransaction.date.add(1, "month");
  }

  if (ts.interval >= 360 && ts.interval <= 370) {
    return lastTransaction.date.add(1, "year");
  }

  return lastTransaction.date.add(ts.interval, "day");
}

export function totalRecurring(ts: TransactionSequence) {
  const lastTransaction = ts.transactions[0];
  return _.sumBy(lastTransaction.postings, (t) => _.max([0, t.amount]));
}

export function sortTrantionSequence(transactionSequences: TransactionSequence[]) {
  return _.chain(transactionSequences)
    .sortBy((ts) => {
      return Math.abs(nextDate(ts).diff(dayjs()));
    })
    .value();
}

const storageKey = "theme-preference";

export function getColorPreference() {
  if (localStorage.getItem(storageKey)) {
    return localStorage.getItem(storageKey);
  } else {
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  }
}

export function darkLightColor(dark: string, light: string) {
  return getColorPreference() == "dark" ? dark : light;
}

export function setColorPreference(theme: string) {
  localStorage.setItem(storageKey, theme);
}

export function isZero(n: number) {
  return n < 0.0001 && n > -0.0001;
}
