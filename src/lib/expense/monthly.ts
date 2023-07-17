import * as d3 from "d3";
import legend from "d3-svg-legend";
import dayjs, { Dayjs } from "dayjs";
import chroma from "chroma-js";
import _ from "lodash";
import {
  forEachMonth,
  formatFixedWidthFloat,
  formatCurrency,
  formatPercentage,
  formatCurrencyCrude,
  type Posting,
  setHtml,
  skipTicks,
  tooltip,
  generateColorScheme,
  restName,
  firstName
} from "$lib/utils";
import COLORS from "$lib/colors";
import type { Writable } from "svelte/store";
import { iconify } from "$lib/icon";
import { byExpenseGroup, expenseGroup, pieData } from "$lib/expense";

export function renderCalendar(
  month: string,
  expenses: Posting[],
  z: d3.ScaleOrdinal<string, string, never>,
  groups: string[]
) {
  const id = "#d3-current-month-expense-calendar";
  const monthStart = dayjs(month, "YYYY-MM");
  const monthEnd = monthStart.endOf("month");
  const weekStart = monthStart.startOf("week");
  const weekEnd = monthEnd.endOf("week");

  const alpha = d3.scaleLinear().range([0.3, 1]);

  const expensesByDay: Record<string, Posting[]> = {};
  const days: Dayjs[] = [];
  let d = weekStart;
  while (d.isSameOrBefore(weekEnd)) {
    days.push(d);
    expensesByDay[d.format("YYYY-MM-DD")] = _.filter(
      expenses,
      (e) => e.date.isSame(d, "day") && _.includes(groups, expenseGroup(e))
    );

    d = d.add(1, "day");
  }

  const expensesByDayTotal = _.mapValues(expensesByDay, (ps) => _.sumBy(ps, (p) => p.amount));

  alpha.domain(d3.extent(_.values(expensesByDayTotal)));

  const root = d3.select(id);
  const dayDivs = root.select("div.days").selectAll("div").data(days);

  const tooltipContent = (d: Dayjs) => {
    const es = expensesByDay[d.format("YYYY-MM-DD")];
    if (_.isEmpty(es)) {
      return null;
    }
    const total = _.sumBy(es, (p) => p.amount);
    return tooltip(
      es.map((p) => {
        return [
          [iconify(restName(p.account), { group: firstName(p.account) })],
          [p.payee, "is-clipped"],
          [formatCurrency(p.amount), "has-text-weight-bold has-text-right"]
        ];
      }),
      { total: formatCurrency(total), header: es[0].date.format("DD MMM YYYY") }
    );
  };

  const dayDiv = dayDivs
    .join("div")
    .attr("class", "date p-1")
    .style("position", "relative")
    .attr("data-tippy-content", tooltipContent)
    .style("visibility", (d) =>
      d.isBefore(monthStart) || d.isAfter(monthEnd) ? "hidden" : "visible"
    );

  dayDiv
    .selectAll("span.day")
    .data((d) => [d])
    .join("span")
    .attr("class", "day has-text-grey-light")
    .style("position", "absolute")
    .text((d) => d.date().toString());

  dayDiv
    .selectAll("span.total")
    .data((d) => [d])
    .join("span")
    .attr("class", "total is-size-7 has-text-weight-bold")
    .style("position", "absolute")
    .style("bottom", "-5px")
    .style("color", (d) =>
      chroma(COLORS.lossText)
        .alpha(alpha(expensesByDayTotal[d.format("YYYY-MM-DD")]))
        .hex()
    )
    .text((d) => {
      const total = expensesByDayTotal[d.format("YYYY-MM-DD")];
      if (total > 0) {
        return formatCurrencyCrude(total);
      }
      return "";
    });

  const width = 35;
  const height = 50;

  dayDiv
    .selectAll("svg")
    .data((d) => [d])
    .join("svg")
    .attr("width", width)
    .attr("height", height)
    .attr("viewBox", [-width / 2, -height / 2, width, height])
    .attr("style", "max-width: 100%; height: auto; height: intrinsic;")
    .selectAll("path")
    .data((d) => pieData(expensesByDay[d.format("YYYY-MM-DD")]))
    .join("path")
    .attr("fill", function (d) {
      return z(d.data.category);
    })
    .attr("d", (arc) => {
      return d3.arc().innerRadius(13).outerRadius(17)(arc as any);
    });
}

export function renderSelectedMonth(
  renderer: (ps: Posting[]) => void,
  expenses: Posting[],
  incomes: Posting[],
  taxes: Posting[],
  investments: Posting[]
) {
  renderer(expenses);
  setHtml("current-month-income", sumCurrency(incomes, -1), COLORS.gainText);
  const taxRate = sum(taxes) / sum(incomes, -1);
  setHtml("current-month-tax", sumCurrency(taxes), COLORS.lossText);
  setHtml("current-month-tax-rate", formatPercentage(taxRate), COLORS.lossText);
  const expensesRate = sum(expenses) / (sum(incomes, -1) - sum(taxes));
  setHtml("current-month-expenses", sumCurrency(expenses), COLORS.lossText);
  setHtml("current-month-expenses-rate", formatPercentage(expensesRate), COLORS.lossText);
  setHtml("current-month-investment", sumCurrency(investments), COLORS.secondary);
  const savingsRate = sum(investments) / (sum(incomes, -1) - sum(taxes));
  setHtml("current-month-savings-rate", formatPercentage(savingsRate), COLORS.secondary);
}

function sum(postings: Posting[], sign = 1) {
  return sign * _.sumBy(postings, (p) => p.amount);
}

function sumCurrency(postings: Posting[], sign = 1) {
  return formatCurrency(sign * _.sumBy(postings, (p) => p.amount));
}

export function renderMonthlyExpensesTimeline(
  postings: Posting[],
  groupsStore: Writable<string[]>,
  monthStore: Writable<string>
) {
  const id = "#d3-monthly-expense-timeline";
  const timeFormat = "MMM-YYYY";
  const MAX_BAR_WIDTH = 40;
  const svg = d3.select(id),
    margin = { top: 40, right: 30, bottom: 60, left: 40 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const groups = _.chain(postings).map(expenseGroup).uniq().sort().value();

  const defaultValues = _.zipObject(
    groups,
    _.map(groups, () => 0)
  );

  const [start, end] = d3.extent(_.map(postings, (p) => p.date));
  const ms = _.groupBy(postings, (p) => p.date.format(timeFormat));
  const ys = _.chain(postings)
    .groupBy((p) => p.date.format("YYYY"))
    .map((ps, k) => {
      const trend = _.chain(ps)
        .groupBy(expenseGroup)
        .map((ps, g) => {
          let months = 12;
          if (start.format("YYYY") == k) {
            months -= start.month();
          }

          if (end.format("YYYY") == k) {
            months -= 11 - end.month();
          }

          return [g, _.sum(_.map(ps, (p) => p.amount)) / months];
        })
        .fromPairs()
        .value();

      return [k, _.merge({}, defaultValues, trend)];
    })
    .fromPairs()
    .value();

  interface Point {
    month: string;
    timestamp: Dayjs;
    [key: string]: number | string | Dayjs;
  }

  const points: Point[] = [];

  forEachMonth(start, end, (month) => {
    const postings = ms[month.format(timeFormat)] || [];
    const values = _.chain(postings)
      .groupBy(expenseGroup)
      .map((postings, key) => [key, _.sum(_.map(postings, (p) => p.amount))])
      .fromPairs()
      .value();

    points.push(
      _.merge(
        {
          timestamp: month,
          month: month.format(timeFormat),
          postings: postings,
          trend: {}
        },
        defaultValues,
        values
      )
    );
  });

  const x = d3.scaleBand().range([0, width]).paddingInner(0.1).paddingOuter(0);
  const y = d3.scaleLinear().range([height, 0]);

  const z = generateColorScheme(groups);

  const tooltipContent = (allowedGroups: string[]) => {
    return (d: d3.SeriesPoint<Record<string, number>>) => {
      let grandTotal = 0;
      return tooltip(
        _.flatMap(allowedGroups, (key) => {
          const total = (d.data as any)[key];
          if (total > 0) {
            grandTotal += total;
            return [
              [
                iconify(key, { group: "Expenses" }),
                [formatCurrency(total), "has-text-weight-bold has-text-right"]
              ]
            ];
          }
          return [];
        }),
        { total: formatCurrency(grandTotal), header: (d.data.timestamp as any).format("MMM YYYY") }
      );
    };
  };

  const xAxis = g.append("g").attr("class", "axis x");
  const yAxis = g.append("g").attr("class", "axis y");

  const bars = g.append("g");
  const line = g
    .append("path")
    .attr("stroke", COLORS.primary)
    .attr("stroke-width", "2px")
    .attr("stroke-linecap", "round")
    .attr("stroke-dasharray", "5,5");

  const render = (allowedGroups: string[]) => {
    groupsStore.set(allowedGroups);
    const sum = (p: Point) => _.sum(_.map(allowedGroups, (k) => p[k]));
    x.domain(points.map((p) => p.month));
    y.domain([0, d3.max(points, sum)]);

    const t = svg.transition().duration(750);
    xAxis
      .attr("transform", "translate(0," + height + ")")
      .transition(t)
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

    yAxis.transition(t).call(d3.axisLeft(y).tickSize(-width).tickFormat(formatCurrencyCrude));

    line
      .transition(t)
      .attr(
        "d",
        d3
          .line<Point>()
          .curve(d3.curveStepAfter)
          .x((p) => x(p.month))
          .y((p) => {
            const total = _.chain(ys[p.timestamp.format("YYYY")])
              .pick(allowedGroups)
              .values()
              .sum()
              .value();

            return y(total);
          })(points)
      )
      .attr("fill", "none");

    bars
      .selectAll("g")
      .data(
        d3.stack().offset(d3.stackOffsetDiverging).keys(allowedGroups)(
          points as { [key: string]: number }[]
        ),
        (d: any) => d.key
      )
      .join(
        (enter) =>
          enter.append("g").attr("fill", function (d) {
            return z(d.key);
          }),
        (update) => update.transition(t),
        (exit) =>
          exit.selectAll("rect").transition(t).attr("y", y.range()[0]).attr("height", 0).remove()
      )
      .selectAll("rect")
      .data(function (d) {
        return d;
      })
      .join(
        (enter) =>
          enter
            .append("rect")
            .attr("class", "zoomable")
            .on("click", (_event, data) => {
              const timestamp: Dayjs = data.data.timestamp as any;
              monthStore.set(timestamp.format("YYYY-MM"));
            })
            .attr("data-tippy-content", tooltipContent(allowedGroups))
            .attr("x", function (d) {
              return (
                x((d.data as any).month) +
                (x.bandwidth() - Math.min(x.bandwidth(), MAX_BAR_WIDTH)) / 2
              );
            })
            .attr("width", Math.min(x.bandwidth(), MAX_BAR_WIDTH))
            .attr("y", y.range()[0])
            .transition(t)
            .attr("y", function (d) {
              return y(d[1]);
            })
            .attr("height", function (d) {
              return y(d[0]) - y(d[1]);
            }),
        (update) =>
          update
            .attr("data-tippy-content", tooltipContent(allowedGroups))
            .transition(t)
            .attr("y", function (d) {
              return y(d[1]);
            })
            .attr("height", function (d) {
              return y(d[0]) - y(d[1]);
            }),
        (exit) => exit.transition(t).remove()
      );
  };

  let selectedGroups = groups;
  render(selectedGroups);

  svg.append("g").attr("class", "legendOrdinal").attr("transform", "translate(40,0)");

  const legendOrdinal = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(100)
    .labels(({ i, generatedLabels }: { i: number; generatedLabels: string[] }) => {
      return iconify(generatedLabels[i], { group: "Expenses" });
    })
    .on("cellclick", function () {
      const group = this.__data__;
      if (selectedGroups.length == 1 && selectedGroups[0] == group) {
        selectedGroups = groups;
        svg.selectAll(".legendOrdinal .cell .label").classed("selected", false);
      } else {
        selectedGroups = [group];
        svg.selectAll(".legendOrdinal .cell .label").classed("selected", false);
        d3.select(this).selectAll(".label").classed("selected", true);
      }

      render(selectedGroups);
    })
    .scale(z);

  svg.select(".legendOrdinal").call(legendOrdinal as any);
  return { z: z };
}

export function renderCurrentExpensesBreakdown(z: d3.ScaleOrdinal<string, string, never>) {
  const id = "#d3-current-month-breakdown";
  const BAR_HEIGHT = 20;
  const svg = d3.select(id),
    margin = { top: 0, right: 160, bottom: 20, left: 100 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const x = d3.scaleLinear().range([0, width]);
  const y = d3.scaleBand().paddingInner(0.1).paddingOuter(0);

  const xAxis = g.append("g").attr("class", "axis y");
  const yAxis = g.append("g").attr("class", "axis y dark");

  const bar = g.append("g");

  return (postings: Posting[]) => {
    interface Point {
      category: string;
      postings: Posting[];
      total: number;
    }

    const categories = byExpenseGroup(postings);
    const keys = _.chain(categories)
      .sortBy((c) => c.total)
      .map((c) => c.category)
      .value();

    const points = _.values(categories);
    const total = _.sumBy(points, (p) => p.total);

    const height = BAR_HEIGHT * keys.length;
    svg.attr("height", height + margin.top + margin.bottom);

    y.domain(keys);
    x.domain([0, d3.max(points, (p) => p.total)]);
    y.range([height, 0]);

    const t = svg.transition().duration(750);

    xAxis
      .attr("transform", "translate(0," + height + ")")
      .transition(t)
      .call(d3.axisBottom(x).tickSize(-height).tickFormat(skipTicks(60, x, formatCurrencyCrude)));

    yAxis
      .transition(t)
      .call(d3.axisLeft(y).tickFormat((g) => iconify(g, { group: "Expenses", suffix: true })));

    const tooltipContent = (d: Point) => {
      const total = _.sumBy(d.postings, (p) => p.amount);
      return tooltip(
        d.postings.map((p) => {
          return [
            p.date.format("DD MMM YYYY"),
            [p.payee, "is-clipped"],
            [formatCurrency(p.amount), "has-text-weight-bold has-text-right"]
          ];
        }),
        {
          total: formatCurrency(total),
          header: `${d.postings[0].date.format("MMM YYYY")} ${d.category}`
        }
      );
    };

    bar
      .selectAll("rect")
      .data(points, (p: any) => p.category)
      .join(
        (enter) =>
          enter
            .append("rect")
            .attr("fill", function (d) {
              return z(d.category);
            })
            .attr("data-tippy-content", tooltipContent)
            .attr("x", x(0))
            .attr("y", function (d) {
              return y(d.category) + (y.bandwidth() - Math.min(y.bandwidth(), BAR_HEIGHT)) / 2;
            })
            .attr("width", function (d) {
              return x(d.total);
            })
            .attr("height", y.bandwidth()),

        (update) =>
          update
            .attr("fill", function (d) {
              return z(d.category);
            })
            .attr("data-tippy-content", tooltipContent)
            .transition(t)
            .attr("x", x(0))
            .attr("y", function (d) {
              return y(d.category) + (y.bandwidth() - Math.min(y.bandwidth(), BAR_HEIGHT)) / 2;
            })
            .attr("width", function (d) {
              return x(d.total);
            })
            .attr("height", y.bandwidth()),

        (exit) => exit.remove()
      );

    const rightLabel = (d: Point) =>
      `${formatCurrency(d.total)} ${formatFixedWidthFloat((d.total / total) * 100, 6)}%`;

    bar
      .selectAll("text")
      .data(points, (p: any) => p.category)
      .join(
        (enter) =>
          enter
            .append("text")
            .attr("text-anchor", "end")
            .attr("dominant-baseline", "middle")
            .attr("y", function (d) {
              return y(d.category) + y.bandwidth() / 2;
            })
            .attr("x", width + 135)
            .style("white-space", "pre")
            .style("font-size", "13px")
            .style("font-weight", "bold")
            .style("fill", function (d) {
              return chroma(z(d.category)).darken(0.8).hex();
            })
            .attr("class", "is-family-monospace")
            .text(rightLabel),
        (update) =>
          update
            .text(rightLabel)
            .transition(t)
            .attr("y", function (d) {
              return y(d.category) + y.bandwidth() / 2;
            }),
        (exit) => exit.remove()
      );

    return;
  };
}
