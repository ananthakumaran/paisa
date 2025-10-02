import sha256 from "crypto-js/sha256";
import dayjs from "dayjs";
import _ from "lodash";
import * as d3 from "d3";
import { loading } from "../store";
import type { JSONSchema7 } from "json-schema";
import { get } from "svelte/store";
import { obscure } from "../persisted_store";
import { error } from "@sveltejs/kit";
import { goto } from "$app/navigation";
import chroma from "chroma-js";
import { iconGlyph } from "./icon";

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
  status: string;
  tag_recurring: string;
  transaction_begin_line: number;
  transaction_end_line: number;
  file_name: string;
  note: string;
  transaction_note: string;

  market_amount: number;
  balance: number;
}

export interface IncomeStatement {
  startingBalance: number;
  endingBalance: number;
  date: dayjs.Dayjs;
  income: Record<string, number>;
  interest: Record<string, number>;
  equity: Record<string, number>;
  pnl: Record<string, number>;
  liabilities: Record<string, number>;
  tax: Record<string, number>;
  expenses: Record<string, number>;
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

export interface TransactionSchedule {
  actual: dayjs.Dayjs;
  scheduled: dayjs.Dayjs;
  transaction: Transaction;
  key: string;
  amount: number;
}

export interface TransactionSequence {
  transactions: Transaction[];
  period: string;
  key: string;
  interval: number;

  // computed
  schedules: TransactionSchedule[];
  pastSchedules: TransactionSchedule[];
  futureSchedules: TransactionSchedule[];
  schedulesByMonth: Record<string, TransactionSchedule[]>;
}

export interface Transaction {
  id: string;
  date: dayjs.Dayjs;
  payee: string;
  beginLine: number;
  endLine: number;
  fileName: string;
  note: string;
  postings: Posting[];
}

export interface BalancedPosting {
  from: Posting;
  to: Posting;
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
  balanceUnits: number;
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
  market_amount: number;

  // computed
  percent: number;
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

export interface RetirementGoalProgress {
  savingsTotal: number;
  investmentTotal: number;
  gainTotal: number;
  savingsTimeline: Point[];
  swr: number;
  yearlyExpense: number;
  xirr: number;
  name: string;
  type: string;
  icon: string;
  postings: Posting[];
  balances: Record<string, AssetBreakdown>;
}

export interface SavingsGoalProgress {
  investmentTotal: number;
  gainTotal: number;
  savingsTotal: number;
  savingsTimeline: Point[];
  target: number;
  targetDate: string;
  rate: number;
  xirr: number;
  postings: Posting[];
  name: string;
  type: string;
  icon: string;
  paymentPerPeriod: number;
  balances: Record<string, AssetBreakdown>;
}

export interface Legend {
  shape: "line" | "square" | "texture";
  color: string;
  label: string;
  texture?: any;
  onClick?: (legend: Legend) => void;

  selected?: boolean;
}

interface File {
  type: "file";
  name: string;
  content: string;
  versions: string[];
}

// eslint-disable-next-line @typescript-eslint/no-empty-interface
export interface LedgerFile extends File {}

export interface Directory {
  type: "directory";
  name: string;
  children: Array<Directory | LedgerFile | SheetFile>;
}

// eslint-disable-next-line @typescript-eslint/no-empty-interface
export interface SheetFile extends File {}

export function buildDirectoryTree<T extends File>(files: T[]) {
  const root: Directory = {
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
      current = found as Directory;
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

export interface SheetFileError {
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

export interface ImportTemplate {
  id: string;
  name: string;
  content: string;
  template_type: string;
}

export interface Log {
  time: dayjs.Dayjs;
  level: string;
  msg: string;
}

export interface CreditCardBill {
  openingBalance: number;
  closingBalance: number;
  debits: number;
  credits: number;
  statementStartDate: dayjs.Dayjs;
  statementEndDate: dayjs.Dayjs;
  dueDate: dayjs.Dayjs;
  paidDate: dayjs.Dayjs;
  postings: Posting[];
  transactions: Transaction[];
}

export interface CreditCardSummary {
  account: string;
  network: string;
  number: string;
  balance: number;
  bills: CreditCardBill[];
  creditLimit: number;
  expirationDate: dayjs.Dayjs;
  yearlySpends: { [year: string]: { [month: string]: number } };
}

export interface GoalSummary {
  type: string;
  name: string;
  id: string;
  icon: string;
  current: number;
  target: number;
  targetDate: string;
  priority: number;
}

export interface SheetLineResult {
  line: number;
  result: string;
  error: boolean;
  underline?: boolean;
  bold?: boolean;
  align?: "left" | "right";
}

const tokenKey = "token";

type RequestOptions = RequestInit & {
  background?: boolean;
};

export function ajax(
  route: "/api/config"
): Promise<{ config: UserConfig; schema: JSONSchema7; now: dayjs.Dayjs; accounts: string[] }>;
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
export function ajax(
  route: "/api/assets/balance"
): Promise<{ asset_breakdowns: Record<string, AssetBreakdown> }>;
export function ajax(route: "/api/liabilities/repayment"): Promise<{ repayments: Posting[] }>;
export function ajax(
  route: "/api/liabilities/balance"
): Promise<{ liability_breakdowns: LiabilityBreakdown[] }>;
export function ajax(route: "/api/price"): Promise<{ prices: Record<string, Price[]> }>;
export function ajax(route: "/api/transaction"): Promise<{ transactions: Transaction[] }>;
export function ajax(
  route: "/api/transaction/balanced"
): Promise<{ balancedPostings: BalancedPosting[] }>;
export function ajax(route: "/api/networth"): Promise<{
  networthTimeline: Networth[];
  xirr: number;
}>;
export function ajax(route: "/api/gain"): Promise<{
  gain_breakdown: Gain[];
}>;
export function ajax(route: "/api/dashboard"): Promise<{
  checkingBalances: { asset_breakdowns: Record<string, AssetBreakdown> };
  expenses: { [key: string]: Posting[] };
  cashFlows: CashFlow[];
  transactionSequences: TransactionSequence[];
  networth: { networth: Networth; xirr: number };
  transactions: Transaction[];
  budget: {
    budgetsByMonth: { [key: string]: Budget };
  };
  goalSummaries: GoalSummary[];
}>;

export function ajax(
  route: "/api/gain/:name",
  options?: RequestOptions,
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
  graph: { [key: string]: Graph };
}>;

export function ajax(route: "/api/budget"): Promise<{
  budgetsByMonth: { [key: string]: Budget };
  checkingBalance: number;
  availableForBudgeting: number;
}>;

export function ajax(route: "/api/cash_flow"): Promise<{ cash_flows: CashFlow[] }>;
export function ajax(route: "/api/income_statement"): Promise<{
  yearly: Record<string, IncomeStatement>;
  monthly: Record<string, Record<string, IncomeStatement>>;
}>;

export function ajax(
  route: "/api/recurring"
): Promise<{ transaction_sequences: TransactionSequence[] }>;

export function ajax(route: "/api/liabilities/interest"): Promise<{
  interest_timeline_breakdown: Interest[];
}>;

export function ajax(route: "/api/credit_cards"): Promise<{ creditCards: CreditCardSummary[] }>;

export function ajax(
  route: "/api/credit_cards/:account",
  options?: RequestOptions,
  params?: Record<string, string>
): Promise<{ creditCard: CreditCardSummary; found: boolean }>;

export function ajax(route: "/api/goals"): Promise<{ goals: GoalSummary[] }>;
export function ajax(
  route: "/api/goals/retirement/:name",
  options?: RequestOptions,
  params?: Record<string, string>
): Promise<RetirementGoalProgress>;
export function ajax(
  route: "/api/goals/savings/:name",
  options?: RequestOptions,
  params?: Record<string, string>
): Promise<SavingsGoalProgress>;

export function ajax(route: "/api/account/tf_idf"): Promise<AccountTfIdf>;
export function ajax(
  route: "/api/templates",
  options?: RequestOptions
): Promise<{ templates: ImportTemplate[] }>;
export function ajax(
  route: "/api/templates/upsert",
  options?: RequestOptions
): Promise<{ saved: boolean; message?: string; template: ImportTemplate }>;
export function ajax(
  route: "/api/templates/delete",
  options?: RequestOptions
): Promise<{ success: boolean; message?: string }>;

export function ajax(
  route: "/api/editor/files",
  options?: RequestOptions
): Promise<{
  files: LedgerFile[];
  accounts: string[];
  commodities: string[];
  payees: string[];
}>;

export function ajax(
  route: "/api/editor/validate",
  options?: RequestOptions
): Promise<{ errors: LedgerFileError[]; output: string }>;

export function ajax(
  route: "/api/editor/save",
  options?: RequestOptions
): Promise<{ errors: LedgerFileError[]; saved: boolean; file: LedgerFile; message: string }>;

export function ajax(
  route: "/api/editor/file",
  options?: RequestOptions
): Promise<{ file: LedgerFile }>;

export function ajax(
  route: "/api/editor/file/delete_backups",
  options?: RequestOptions
): Promise<{ file: LedgerFile }>;

export function ajax(route: "/api/sheets/files"): Promise<{
  files: SheetFile[];
  postings: Posting[];
}>;

export function ajax(
  route: "/api/sheets/save",
  options?: RequestOptions
): Promise<{ saved: boolean; file: SheetFile; message: string }>;

export function ajax(
  route: "/api/sheets/file",
  options?: RequestOptions
): Promise<{ file: SheetFile }>;

export function ajax(
  route: "/api/sheets/file/delete_backups",
  options?: RequestOptions
): Promise<{ file: SheetFile }>;

export function ajax(
  route: "/api/price/delete",
  options?: RequestOptions
): Promise<{ success: boolean; message: string }>;

export function ajax(
  route: "/api/sync",
  options?: RequestOptions
): Promise<{ success: boolean; message: string }>;
export function ajax(
  route: "/api/price/providers",
  options?: RequestOptions
): Promise<{ providers: PriceProvider[] }>;

export function ajax(
  route: "/api/price/providers/delete/:provider",
  options?: RequestOptions,
  params?: Record<string, string>
): Promise<{
  gain_timeline_breakdown: AccountGain;
  portfolio_allocation: PortfolioAllocation;
  asset_breakdown: AssetBreakdown;
}>;

export function ajax(
  route: "/api/price/autocomplete",
  options?: RequestOptions
): Promise<{ completions: AutoCompleteItem[] }>;
export function ajax(route: "/api/init", options?: RequestOptions): Promise<any>;

export function ajax(
  route: "/api/config",
  options?: RequestOptions
): Promise<{ success: boolean; error?: string }>;

export function ajax(route: "/api/ping"): Promise<{ success: boolean; error?: string }>;

export async function ajax(
  route: string,
  options?: RequestOptions,
  params?: Record<string, string>
) {
  const background = options?.background;
  if (!background) {
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

  const token = localStorage.getItem(tokenKey);
  if (!_.isEmpty(token)) {
    options.headers["X-Auth"] = token;
  }

  const response = await fetch(route, options);
  const body = await response.text();
  if (!background) {
    loading.set(false);
  }

  if (response.status == 401 && route != "/api/ping") {
    logout();
    await goto("/login");
    error(401, "Unauthorized");
  }

  return JSON.parse(body, (key, value) => {
    if (
      _.isString(value) &&
      /Date|date|time|now/.test(key) &&
      /^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(.[0-9]+)?(Z|[+-][0-9]{2}:[0-9]{2})$/.test(
        value
      )
    ) {
      return dayjs(value);
    }
    return value;
  });
}

export async function login(username: string, password: string) {
  localStorage.setItem(tokenKey, `${username}:${sha256(password)}`);
  return await ajax("/api/ping");
}

export function isLoggedIn() {
  return !_.isEmpty(localStorage.getItem(tokenKey));
}

export function logout() {
  localStorage.removeItem(tokenKey);
}

function normalize(value: number) {
  if (get(obscure)) {
    value = 0;
  }

  // minus 0
  if (1 / value === -Infinity) {
    value = 0;
  }

  if (!Number.isFinite(value)) {
    value = 0;
  }

  return value;
}

export function configUpdated() {
  dayjs.locale("en");
  dayjs.updateLocale("en", {
    weekStart: USER_CONFIG.week_starting_day
  });
}

export function setNow(value: dayjs.Dayjs) {
  if (value) {
    globalThis.__now = value;
  }
}

export function now(): dayjs.Dayjs {
  if (globalThis.__now) {
    return globalThis.__now;
  }
  return dayjs();
}

function unicodeMinus(value: string) {
  return value.replace(/^-/, "\u2212");
}

export function formatCurrency(value: number, precision: number = null) {
  value = normalize(value);

  if (precision == null) {
    precision = USER_CONFIG.display_precision;
  }

  return unicodeMinus(
    value.toLocaleString(USER_CONFIG.locale, {
      minimumFractionDigits: precision,
      maximumFractionDigits: precision
    })
  );
}

export function formatCurrencyCrude(value: number) {
  return formatCurrencyCrudeWithPrecision(value, -1);
}

export function formatCurrencyCrudeWithPrecision(value: number, precision: number) {
  value = normalize(value);

  const options: Intl.NumberFormatOptions = {
    notation: "compact"
  };

  if (precision < 0) {
    options.maximumFractionDigits = 2;
  } else {
    options.maximumFractionDigits = precision;
    options.minimumFractionDigits = precision;
  }

  return unicodeMinus(value.toLocaleString(USER_CONFIG.locale, options));
}

export function formatFloat(value: number, precision = 2) {
  value = normalize(value);

  return unicodeMinus(
    value.toLocaleString(USER_CONFIG.locale, {
      minimumFractionDigits: precision,
      maximumFractionDigits: precision
    })
  );
}

export function formatFloatUptoPrecision(value: number, precision = 2) {
  value = normalize(value);

  return unicodeMinus(
    value.toLocaleString(USER_CONFIG.locale, {
      maximumFractionDigits: precision
    })
  );
}

export function formatPercentage(value: number, precision = 0) {
  value = normalize(value);

  return unicodeMinus(
    value.toLocaleString(USER_CONFIG.locale, {
      style: "percent",
      minimumFractionDigits: precision
    })
  );
}

export function formatFixedWidthFloat(value: number, width: number, precision = 2) {
  value = normalize(value);

  const formatted = unicodeMinus(
    value.toLocaleString(USER_CONFIG.locale, {
      minimumFractionDigits: precision,
      maximumFractionDigits: precision
    })
  );

  if (formatted.length < width) {
    return formatted.padStart(width, " ");
  }
  return formatted;
}

export function forEachMonth(
  start: dayjs.Dayjs,
  end: dayjs.Dayjs,
  cb: (current: dayjs.Dayjs) => any
) {
  let current = start.startOf("month");
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
  let current = beginningOfFinancialYear(start);
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

function beginningOfFinancialYear(date: dayjs.Dayjs) {
  date = date.startOf("month");
  if (date.month() + 1 < USER_CONFIG.financial_year_starting_month) {
    return date
      .add(-1, "year")
      .add(USER_CONFIG.financial_year_starting_month - date.month() - 1, "month");
  } else {
    return date.add(-(date.month() + 1 - USER_CONFIG.financial_year_starting_month), "month");
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

export function firstNames(account: string, n: number) {
  return _.take(account.split(":"), n).join(":");
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
  points?: number
) {
  const range = scale.range();
  const width = Math.abs(range[1] - range[0]);
  const s = scale as any;
  points = points || (s.ticks ? s.ticks().length : s.domain().length);
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

export function darkenOrLighten(backgroundColor: string, intensity = 2) {
  const color = d3.rgb(backgroundColor);
  // http://www.w3.org/TR/AERT#color-contrast
  const brightness = (color.r * 299 + color.g * 587 + color.b) / 1000;
  if (brightness > 125) {
    return chroma(backgroundColor).darken(intensity).hex();
  }
  return chroma(backgroundColor).brighten(intensity).hex();
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
  return window.innerWidth < 769;
}

export function rem(value: number) {
  if (isMobile()) {
    return value * 0.857;
  } else {
    return value;
  }
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

export function monthDays(month: string) {
  const monthStart = dayjs(month, "YYYY-MM");
  const monthEnd = monthStart.endOf("month");
  const weekStart = monthStart.startOf("week");
  const weekEnd = monthEnd.endOf("week");

  const days: dayjs.Dayjs[] = [];
  let d = weekStart;
  while (d.isSameOrBefore(weekEnd)) {
    days.push(d);
    d = d.add(1, "day");
  }
  return { days, monthStart, monthEnd };
}

export function prefixMinutesSeconds(cronExpression: string) {
  return cronExpression
    .split("|")
    .map((cron) => "0 0 " + cron)
    .join("|");
}

export function svgTruncate(width: number) {
  return function () {
    const self = d3.select(this);
    let textLength = self.node().getComputedTextLength(),
      text = self.text();
    while (textLength > width && text.length > 0) {
      text = text.slice(0, -1);
      self.text(text + "...");
      textLength = self.node().getComputedTextLength();
    }
  };
}

export function sumPostings(postings: Posting[]) {
  return postings.reduce(
    (sum, p) => (p.account.startsWith("Income:CapitalGains") ? sum + -p.amount : sum + p.amount),
    0
  );
}

export function transactionTotal(transaction: Transaction) {
  return _.sumBy(transaction.postings, (t) => _.max([0, t.amount]));
}

export function formatTextAsHtml(text: string) {
  return `<p>${_.trim(text).replaceAll("\n", "<br />")}</p>`;
}

export function groupSumBy(postings: Posting[], groupBy: _.ValueIteratee<Posting>) {
  return _.chain(postings)
    .groupBy(groupBy)
    .mapValues((ps) => _.sumBy(ps, (p) => p.amount))
    .value();
}

export function asTransaction(p: Posting): Transaction {
  return {
    id: p.id,
    date: p.date,
    payee: p.payee,
    beginLine: p.transaction_begin_line,
    endLine: p.transaction_end_line,
    fileName: p.file_name,
    note: p.transaction_note,
    postings: [p]
  };
}

export function svgUrl(identifier: string) {
  return `url(${new URL("#" + identifier, window.location.toString())})`;
}

export function dueDateIcon(dueDate: dayjs.Dayjs, clearedDate: dayjs.Dayjs) {
  let icon = "fa-circle-check";
  let glyph = iconGlyph("fa6-solid:circle-check");
  let color = "has-text-success";
  let svgColor = "svg-text-success";

  if (!clearedDate) {
    if (dueDate.isBefore(now(), "day")) {
      color = "has-text-danger";
      icon = "fa-exclamation-triangle";
      glyph = iconGlyph("fa6-solid:triangle-exclamation");
      svgColor = "svg-text-danger";
    } else {
      color = "has-text-grey";
      svgColor = "svg-text-grey";
    }
  } else {
    if (clearedDate.isSameOrBefore(dueDate, "day")) {
      color = "has-text-success";
      svgColor = "svg-text-success";
    } else {
      color = "has-text-warning-dark";
      svgColor = "svg-text-warning-dark";
    }
  }

  return { icon, color, svgColor, glyph };
}

export function buildTree<I>(items: I[], accountAccessor: (item: I) => string): I[] {
  const result: I[] = [];

  const sorted = _.sortBy(items, accountAccessor);

  for (const item of sorted) {
    const account = accountAccessor(item);
    const parts = account.split(":");
    let current = result;
    for (let i = 0; i < parts.length; i++) {
      const part = parts[i];
      let found: any = current.find((c) => accountAccessor(c).split(":")[i] === part);
      if (!found) {
        found = { ...item };
        current.push(found);
      }

      if (i !== parts.length - 1) {
        found._children = found._children || [];
        current = found._children;
      }
    }
  }

  return result;
}
