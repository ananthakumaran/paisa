import * as d3 from "d3";
import legend from "d3-svg-legend";
import dayjs, { Dayjs } from "dayjs";
import chroma from "chroma-js";
import _ from "lodash";
import {
  ajax,
  forEachMonth,
  formatFixedWidthFloat,
  formatCurrency,
  formatCurrencyCrude,
  Posting,
  restName,
  secondName,
  setHtml,
  skipTicks,
  tooltip,
  generateColorScheme
} from "./utils";
import COLORS from "./colors";

export default async function () {
  const {
    expenses: expenses,
    month_wise: {
      expenses: grouped_expenses,
      incomes: grouped_incomes,
      investments: grouped_investments,
      taxes: grouped_taxes
    }
  } = await ajax("/api/expense");

  let minDate = dayjs();
  _.each(expenses, (p) => (p.timestamp = dayjs(p.date)));
  const parseDate = (group: { [key: string]: Posting[] }) => {
    _.each(group, (ps) => {
      _.each(ps, (p) => {
        p.timestamp = dayjs(p.date);
        if (p.timestamp.isBefore(minDate)) {
          minDate = p.timestamp;
        }
      });
    });
  };
  parseDate(grouped_expenses);
  parseDate(grouped_incomes);
  parseDate(grouped_investments);
  parseDate(grouped_taxes);

  const max = dayjs().format("YYYY-MM");
  const min = minDate.format("YYYY-MM");
  const input = d3.select<HTMLInputElement, never>("#d3-current-month");
  input.attr("max", max);
  input.attr("min", min);

  const z = renderMonthlyExpensesTimeline(expenses, input.node());
  const renderer = renderCurrentExpensesBreakdown(z);

  const selectMonth = (month) => {
    renderSelectedMonth(
      renderer,
      grouped_expenses[month] || [],
      grouped_incomes[month] || [],
      grouped_taxes[month] || [],
      grouped_investments[month] || []
    );
  };

  input.on("input", (event) => {
    selectMonth(event.srcElement.value);
  });

  input.attr("value", max);
  selectMonth(max);
  input.node().focus();
  input.node().select();
}

function renderSelectedMonth(
  renderer: (ps: Posting[]) => void,
  expenses: Posting[],
  incomes: Posting[],
  taxes: Posting[],
  investments: Posting[]
) {
  renderer(expenses);
  setHtml("current-month-income", sum(incomes, -1), COLORS.gainText);
  setHtml("current-month-tax", sum(taxes), COLORS.lossText);
  setHtml("current-month-expenses", sum(expenses), COLORS.lossText);
  setHtml("current-month-investment", sum(investments), COLORS.secondary);
}

function sum(postings: Posting[], sign = 1) {
  return formatCurrency(sign * _.sumBy(postings, (p) => p.amount));
}

function renderMonthlyExpensesTimeline(
  postings: Posting[],
  dateSelector: HTMLInputElement
) {
  const id = "#d3-expense-timeline";
  const timeFormat = "MMM-YYYY";
  const MAX_BAR_WIDTH = 40;
  const svg = d3.select(id),
    margin = { top: 40, right: 30, bottom: 60, left: 40 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg
      .append("g")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const groups = _.chain(postings)
    .map((p) => secondName(p.account))
    .uniq()
    .sort()
    .value();

  const defaultValues = _.zipObject(
    groups,
    _.map(groups, () => 0)
  );

  const start = _.min(_.map(postings, (p) => p.timestamp)),
    end = dayjs().startOf("month");
  const ms = _.groupBy(postings, (p) => p.timestamp.format(timeFormat));
  const ys = _.chain(postings)
    .groupBy((p) => p.timestamp.format("YYYY"))
    .map((ps, k) => {
      const trend = _.chain(ps)
        .groupBy((p) => secondName(p.account))
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
    [key: string]: number | string | dayjs.Dayjs;
  }

  const points: Point[] = [];

  forEachMonth(start, end, (month) => {
    const postings = ms[month.format(timeFormat)] || [];
    const values = _.chain(postings)
      .groupBy((t) => secondName(t.account))
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
    return (d) => {
      return tooltip(
        _.flatMap(allowedGroups, (key) => {
          const total = (d.data as any)[key];
          if (total > 0) {
            return [
              [
                key,
                [formatCurrency(total), "has-text-weight-bold has-text-right"]
              ]
            ];
          }
          return [];
        })
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
    const sum = (p) => _.sum(_.map(allowedGroups, (k) => p[k]));
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

    yAxis
      .transition(t)
      .call(d3.axisLeft(y).tickSize(-width).tickFormat(formatCurrencyCrude));

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
          exit
            .selectAll("rect")
            .transition(t)
            .attr("y", y.range()[0])
            .attr("height", 0)
            .remove()
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
            .on("click", (event, data) => {
              const timestamp: Dayjs = data.data.timestamp as any;
              dateSelector.value = timestamp.format("YYYY-MM");
              dateSelector.dispatchEvent(new Event("input", { bubbles: true }));
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

  svg
    .append("g")
    .attr("class", "legendOrdinal")
    .attr("transform", "translate(40,0)");

  const legendOrdinal = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(100)
    .labels(groups)
    .on("cellclick", function () {
      const group = this.__data__;
      if (selectedGroups.length == 1 && selectedGroups[0] == group) {
        selectedGroups = groups;
        d3.selectAll(".legendOrdinal .cell .label").attr("fill", "#000");
      } else {
        selectedGroups = [group];
        d3.selectAll(".legendOrdinal .cell .label").attr("fill", "#ccc");
        d3.select(this).selectAll(".label").attr("fill", "#000");
      }
      render(selectedGroups);
    })
    .scale(z);

  svg.select(".legendOrdinal").call(legendOrdinal as any);
  return z;
}

function renderCurrentExpensesBreakdown(
  z: d3.ScaleOrdinal<string, string, never>
) {
  const id = "#d3-current-month-breakdown";
  const BAR_HEIGHT = 20;
  const svg = d3.select(id),
    margin = { top: 10, right: 160, bottom: 20, left: 100 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    g = svg
      .append("g")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

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
    const categories = _.chain(postings)
      .groupBy((p) => restName(p.account))
      .mapValues((ps, category) => {
        return {
          category: category,
          postings: ps,
          total: _.sumBy(ps, (p) => p.amount)
        };
      })
      .value();
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

    yAxis.transition(t).call(d3.axisLeft(y));

    const tooltipContent = (d: Point) => {
      return tooltip(
        d.postings.map((p) => {
          return [
            p.timestamp.format("DD MMM YYYY"),
            [p.payee, "is-clipped"],
            [formatCurrency(p.amount), "has-text-weight-bold has-text-right"]
          ];
        })
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
              return (
                y(d.category) +
                (y.bandwidth() - Math.min(y.bandwidth(), BAR_HEIGHT)) / 2
              );
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
              return (
                y(d.category) +
                (y.bandwidth() - Math.min(y.bandwidth(), BAR_HEIGHT)) / 2
              );
            })
            .attr("width", function (d) {
              return x(d.total);
            })
            .attr("height", y.bandwidth()),

        (exit) => exit.remove()
      );

    bar
      .selectAll("text")
      .data(points, (p: any) => p.category)
      .join(
        (enter) =>
          enter
            .append("text")
            .attr("text-anchor", "end")
            .attr("alignment-baseline", "middle")
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
            .text(
              (d) =>
                `${formatCurrency(d.total)} ${formatFixedWidthFloat(
                  (d.total / total) * 100,
                  6
                )}%`
            ),
        (update) =>
          update
            .text(
              (d) =>
                `${formatCurrency(d.total)} ${formatFixedWidthFloat(
                  (d.total / total) * 100,
                  6
                )}%`
            )
            .transition(t)
            .attr("y", function (d) {
              return y(d.category) + y.bandwidth() / 2;
            }),
        (exit) => exit.remove()
      );

    return;
  };
}
