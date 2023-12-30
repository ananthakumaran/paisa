import * as d3 from "d3";
import chroma from "chroma-js";
import _ from "lodash";
import {
  formatFixedWidthFloat,
  formatCurrency,
  formatPercentage,
  formatCurrencyCrude,
  type Posting,
  skipTicks,
  tooltip,
  secondName,
  financialYear,
  forEachFinancialYear,
  formatCurrencyCrudeWithPrecision,
  now,
  type Legend
} from "$lib/utils";
import COLORS, { generateColorScheme } from "$lib/colors";
import type { Writable } from "svelte/store";
import { iconify } from "$lib/icon";
import { byExpenseGroup, expenseGroup, pieData } from "$lib/expense";
import type { Dayjs } from "dayjs";

export function renderCalendar(
  expenses: Posting[],
  z: d3.ScaleOrdinal<string, string, never>,
  groups: string[]
) {
  const id = "#d3-current-year-expense-calendar";
  const alpha = d3.scaleLinear().range([0.3, 1]);

  const ALL_MONTHS = [
    "Jan",
    "Feb",
    "Mar",
    "Apr",
    "May",
    "Jun",
    "Jul",
    "Aug",
    "Sep",
    "Oct",
    "Nov",
    "Dec"
  ];

  const months: string[] = _.concat(
    _.drop(ALL_MONTHS, USER_CONFIG.financial_year_starting_month - 1),
    _.take(ALL_MONTHS, USER_CONFIG.financial_year_starting_month - 1)
  );
  const expensesByMonth: Record<string, Posting[]> = _.chain(months)
    .map((month) => {
      return [
        month,
        _.filter(
          expenses,
          (e) => _.includes(groups, expenseGroup(e)) && e.date.format("MMM") == month
        )
      ];
    })
    .fromPairs()
    .value();

  const expensesByMonthTotal = _.mapValues(expensesByMonth, (ps) => _.sumBy(ps, (p) => p.amount));

  alpha.domain(d3.extent(_.values(expensesByMonthTotal)));

  const root = d3.select(id);
  const dayDivs = root.select("div.months").selectAll("div.month").data(months);

  const tooltipContent = (month: string) => {
    const es = expensesByMonth[month];
    if (_.isEmpty(es)) {
      return null;
    }
    const byAccount: Record<string, number> = _.chain(es)
      .groupBy((e) => secondName(e.account))
      .map((ps, group) => [group, _.sumBy(ps, (p) => p.amount)])
      .fromPairs()
      .value();
    const total = _.sumBy(es, (p) => p.amount);
    return tooltip(
      _.map(byAccount, (amount, group) => {
        return [
          [iconify(group, { group: "Expenses" })],
          [formatPercentage(amount / total, 1), "has-text-right"],
          [formatCurrency(amount), "has-text-weight-bold has-text-right"]
        ];
      }),
      { total: formatCurrency(total), header: es[0].date.format("MMM YYYY") }
    );
  };

  const monthDiv = dayDivs
    .join("div")
    .attr("class", "month p-1")
    .style("position", "relative")
    .attr("data-tippy-content", tooltipContent);

  const infoDiv = monthDiv
    .selectAll("div.info")
    .data((d) => [d])
    .join("div")
    .style("width", "40px")
    .attr("class", "info");

  infoDiv
    .selectAll("div.day")
    .data((d) => [d])
    .join("div")
    .attr("class", "day has-text-grey-light")
    .text((d) => d);

  infoDiv
    .selectAll("div.total")
    .data((d) => [d])
    .join("div")
    .attr("class", "total is-size-7 has-text-weight-bold")
    .style("color", (d) => chroma(COLORS.lossText).alpha(alpha(expensesByMonthTotal[d])).hex())
    .text((d) => {
      const total = expensesByMonthTotal[d];
      if (total > 0) {
        return formatCurrencyCrudeWithPrecision(total, 1);
      }
      return "";
    });

  const width = 50;
  const height = 50;

  monthDiv
    .selectAll("svg")
    .data((d) => [d])
    .join("svg")
    .attr("width", width)
    .attr("height", height)
    .attr("viewBox", [-width / 2, -height / 2, width, height])
    .attr("style", "max-width: 100%; height: auto; height: intrinsic;")
    .selectAll("path")
    .data((d) => {
      return pieData(expensesByMonth[d]);
    })
    .join("path")
    .attr("fill", function (d) {
      return z(d.data.category);
    })
    .attr("d", (arc) => {
      return d3.arc().innerRadius(0).outerRadius(22)(arc as any);
    });
}

export function renderYearlyExpensesTimeline(
  postings: Posting[],
  groupsStore: Writable<string[]>,
  yearStore: Writable<string>
) {
  if (_.isEmpty(postings)) {
    return { z: null, legends: [] };
  }

  const id = "#d3-yearly-expense-timeline";
  const MAX_BAR_WIDTH = 40;
  const svg = d3.select(id),
    margin = { top: 15, right: 30, bottom: 60, left: 40 },
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

  const start = _.min(_.map(postings, (p) => p.date)),
    end = now().startOf("month");
  const ms = _.groupBy(postings, (p) => financialYear(p.date));

  interface Point {
    fy: string;
    timestamp: Dayjs;
    [key: string]: number | string | Dayjs;
  }

  const points: Point[] = [];

  forEachFinancialYear(start, end, (year) => {
    const postings = ms[financialYear(year)] || [];
    const values = _.chain(postings)
      .groupBy(expenseGroup)
      .map((postings, key) => [key, _.sum(_.map(postings, (p) => p.amount))])
      .fromPairs()
      .value();

    points.push(
      _.merge(
        {
          timestamp: year,
          fy: financialYear(year),
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
        {
          total: formatCurrency(grandTotal),
          header: financialYear(d.data.timestamp as any)
        }
      );
    };
  };

  const xAxis = g.append("g").attr("class", "axis x");
  const yAxis = g.append("g").attr("class", "axis y");

  const bars = g.append("g");

  const render = (allowedGroups: string[]) => {
    groupsStore.set(allowedGroups);
    const sum = (p: Point) => _.sum(_.map(allowedGroups, (k) => p[k]));
    x.domain(points.map((p) => p.fy));
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
              yearStore.set(financialYear(timestamp));
            })
            .attr("data-tippy-content", tooltipContent(allowedGroups))
            .attr("x", function (d) {
              return (
                x((d.data as any).fy) + (x.bandwidth() - Math.min(x.bandwidth(), MAX_BAR_WIDTH)) / 2
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

  const legends = groups.map(
    (group) =>
      ({
        label: iconify(group, { group: "Expenses" }),
        color: z(group),
        shape: "square",
        onClick: () => {
          if (selectedGroups.length == 1 && selectedGroups[0] == group) {
            selectedGroups = groups;
          } else {
            selectedGroups = [group];
          }

          render(selectedGroups);
        }
      }) as Legend
  );

  return { z, legends };
}

export function renderCurrentExpensesBreakdown(z: d3.ScaleOrdinal<string, string, never>) {
  const id = "#d3-current-year-breakdown";
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
      .call(
        d3
          .axisBottom(x)
          .tickSize(-height)
          .tickFormat(skipTicks(60, x, formatCurrencyCrude))
      );

    yAxis
      .transition(t)
      .call(d3.axisLeft(y).tickFormat((g) => iconify(g, { group: "Expenses", suffix: true })));

    const tooltipContent = (d: Point) => {
      const total = _.sumBy(d.postings, (p) => p.amount);
      const byMonth: Record<string, number> = _.chain(d.postings)
        .groupBy((e) => e.date.format("MMM"))
        .map((ps, group) => [group, _.sumBy(ps, (p) => p.amount)])
        .fromPairs()
        .value();
      return tooltip(
        _.map(byMonth, (amount, month) => {
          return [
            month,
            [formatPercentage(amount / total, 1), "has-text-right"],
            [formatCurrency(amount), "has-text-weight-bold has-text-right"]
          ];
        }),
        {
          total: formatCurrency(total),
          header: `${financialYear(d.postings[0].date)} ${d.category}`
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
            .style("font-size", "0.928rem")
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
