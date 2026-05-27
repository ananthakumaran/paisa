import * as d3 from "d3";
import _ from "lodash";
import {
  forEachMonth,
  formatCurrency,
  formatCurrencyCrude,
  type Posting,
  restName,
  skipTicks,
  tooltip,
  rem,
  now,
  type Legend
} from "./utils";
import { generateColorScheme } from "./colors";
import { iconify } from "./icon";
import type dayjs from "dayjs";

export interface AmortizationMonth {
  index: number;
  payment: number;
  principal: number;
  interest: number;
  balance: number;
}

export interface Amortization {
  name: string;
  kind: string;
  schedule: "equal_payment" | "equal_principal";
  principal: number;
  apr: number;
  term_months: number;
  start_date?: string;
  monthly_rate: number;
  monthly_payment: number;
  total_payment: number;
  total_principal: number;
  total_interest: number;
  months: AmortizationMonth[];
}

export interface RepaymentResponse {
  repayments: Posting[];
  amortizations: Amortization[];
}

export function renderMonthlyRepaymentTimeline(postings: Posting[]): Legend[] {
  const id = "#d3-repayment-timeline";
  const timeFormat = "MMM-YYYY";
  const MAX_BAR_WIDTH = rem(40);
  const svg = d3.select(id),
    margin = { top: rem(20), right: rem(30), bottom: rem(60), left: rem(40) },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const groups = _.chain(postings)
    .map((p) => restName(p.account))
    .uniq()
    .sort()
    .value();

  const defaultValues = _.zipObject(
    groups,
    _.map(groups, () => 0)
  );

  const start = _.min(_.map(postings, (p) => p.date)),
    end = now().startOf("month");
  const ts = _.groupBy(postings, (p) => p.date.format(timeFormat));

  interface Point {
    month: string;
    [key: string]: number | string | dayjs.Dayjs;
  }
  const points: Point[] = [];

  forEachMonth(start, end, (month) => {
    const postings = ts[month.format(timeFormat)] || [];
    const values = _.chain(postings)
      .groupBy((t) => restName(t.account))
      .map((postings, key) => [key, _.sum(_.map(postings, (p) => p.amount))])
      .fromPairs()
      .value();

    points.push(
      _.merge(
        {
          month: month.format(timeFormat),
          postings: postings
        },
        defaultValues,
        values
      )
    );
  });

  const x = d3.scaleBand().range([0, width]).paddingInner(0.1).paddingOuter(0);
  const y = d3.scaleLinear().range([height, 0]);

  x.domain(points.map((p) => p.month));
  y.domain([0, d3.max(points, (p: Point) => _.sum(_.map(groups, (k) => p[k])))]);

  const z = generateColorScheme(groups);

  g.append("g")
    .attr("class", "axis x")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x)
        .ticks(5)
        .tickFormat(skipTicks(30, x, (d) => d.toString()))
    )
    .selectAll("text")
    .attr("y", 10)
    .attr("x", -8)
    .attr("dy", ".35em")
    .attr("transform", "rotate(-45)")
    .style("text-anchor", "end");

  g.append("g")
    .attr("class", "axis y")
    .call(d3.axisLeft(y).tickSize(-width).tickFormat(formatCurrencyCrude));

  g.append("g")
    .selectAll("g")
    .data(
      d3.stack().offset(d3.stackOffsetDiverging).keys(groups)(points as { [key: string]: number }[])
    )
    .enter()
    .append("g")
    .attr("fill", function (d) {
      return z(d.key.split("-")[0]);
    })
    .selectAll("rect")
    .data(function (d) {
      return d;
    })
    .enter()
    .append("rect")
    .attr("data-tippy-content", (d) => {
      const postings: Posting[] = (d.data as any).postings;
      const total = _.sumBy(postings, (p) => p.amount);
      return tooltip(
        _.sortBy(
          postings.map((p) => [
            _.drop(p.account.split(":")).join(":"),
            [formatCurrency(p.amount), "has-text-weight-bold has-text-right"]
          ]),
          (r) => r[0]
        ),
        { total: formatCurrency(total) }
      );
    })
    .attr("x", function (d) {
      return (
        x((d.data as any).month) + (x.bandwidth() - Math.min(x.bandwidth(), MAX_BAR_WIDTH)) / 2
      );
    })
    .attr("y", function (d) {
      return y(d[1]);
    })
    .attr("height", function (d) {
      return y(d[0]) - y(d[1]);
    })
    .attr("width", Math.min(x.bandwidth(), MAX_BAR_WIDTH));

  return _.map(groups, (group) => ({
    label: iconify(group, { group: "Liabilities" }),
    color: z(group),
    shape: "square"
  }));
}

// Render an amortization schedule as a stacked bar chart (principal + interest)
// with a remaining-balance line overlay.
export function renderAmortizationChart(id: string, a: Amortization): Legend[] {
  const svg = d3.select(id);
  // Clear previous render (in case of re-render).
  svg.selectAll("*").remove();

  const margin = { top: rem(20), right: rem(50), bottom: rem(60), left: rem(60) };
  const container = document.getElementById(id.substring(1));
  if (!container || !container.parentElement) return [];
  const width = container.parentElement.clientWidth - margin.left - margin.right;
  const height = +svg.attr("height") - margin.top - margin.bottom;

  const g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const months = a.months;
  const MAX_BAR_WIDTH = rem(20);

  const x = d3
    .scaleBand<number>()
    .domain(months.map((m) => m.index))
    .range([0, width])
    .paddingInner(0.1)
    .paddingOuter(0);

  const yLeftMax = d3.max(months, (m) => m.principal + m.interest) || 0;
  const yLeft = d3.scaleLinear().domain([0, yLeftMax]).nice().range([height, 0]);

  const yRight = d3.scaleLinear().domain([0, a.principal]).nice().range([height, 0]);

  // X axis (sparse ticks).
  g.append("g")
    .attr("class", "axis x")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x)
        .tickValues(months.filter((m) => m.index % 12 === 0 || m.index === 1).map((m) => m.index))
        .tickFormat((d) => d.toString())
    );

  // Y axis (left) — monthly payment composition.
  g.append("g")
    .attr("class", "axis y")
    .call(d3.axisLeft(yLeft).tickSize(-width).tickFormat(formatCurrencyCrude));

  // Y axis (right) — remaining balance.
  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(" + width + ",0)")
    .call(d3.axisRight(yRight).tickFormat(formatCurrencyCrude));

  const colorPrincipal = "#3273dc";
  const colorInterest = "#ff3860";
  const colorBalance = "#23d160";

  // Stacked bars: principal (bottom) + interest (top).
  const series = [
    {
      key: "principal",
      color: colorPrincipal,
      valueFn: (m: AmortizationMonth) => m.principal,
      baseFn: (_m: AmortizationMonth) => 0
    },
    {
      key: "interest",
      color: colorInterest,
      valueFn: (m: AmortizationMonth) => m.interest,
      baseFn: (m: AmortizationMonth) => m.principal
    }
  ];

  for (const s of series) {
    g.append("g")
      .attr("fill", s.color)
      .selectAll("rect")
      .data(months)
      .enter()
      .append("rect")
      .attr("data-tippy-content", (m) =>
        tooltip(
          [
            ["Month", [m.index.toString(), "has-text-weight-bold has-text-right"]],
            ["Principal", [formatCurrency(m.principal), "has-text-weight-bold has-text-right"]],
            ["Interest", [formatCurrency(m.interest), "has-text-weight-bold has-text-right"]],
            ["Payment", [formatCurrency(m.payment), "has-text-weight-bold has-text-right"]],
            ["Balance", [formatCurrency(m.balance), "has-text-weight-bold has-text-right"]]
          ],
          { total: formatCurrency(m.payment) }
        )
      )
      .attr("x", (m) => x(m.index) + (x.bandwidth() - Math.min(x.bandwidth(), MAX_BAR_WIDTH)) / 2)
      .attr("y", (m) => yLeft(s.baseFn(m) + s.valueFn(m)))
      .attr("height", (m) => yLeft(s.baseFn(m)) - yLeft(s.baseFn(m) + s.valueFn(m)))
      .attr("width", Math.min(x.bandwidth(), MAX_BAR_WIDTH));
  }

  // Remaining balance line.
  const line = d3
    .line<AmortizationMonth>()
    .x((m) => x(m.index) + x.bandwidth() / 2)
    .y((m) => yRight(m.balance));

  g.append("path")
    .datum(months)
    .attr("fill", "none")
    .attr("stroke", colorBalance)
    .attr("stroke-width", 2)
    .attr("d", line);

  return [
    { label: "Principal", color: colorPrincipal, shape: "square" },
    { label: "Interest", color: colorInterest, shape: "square" },
    { label: "Remaining Balance", color: colorBalance, shape: "line" }
  ];
}
