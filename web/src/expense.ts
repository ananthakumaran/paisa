import * as d3 from "d3";
import legend from "d3-svg-legend";
import { sprintf } from "sprintf-js";
import dayjs, { Dayjs } from "dayjs";
import _ from "lodash";
import {
  ajax,
  forEachMonth,
  formatCurrency,
  formatCurrencyCrude,
  Posting,
  restName,
  secondName,
  setHtml,
  skipTicks,
  tooltip
} from "./utils";

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
  setHtml("current-month-income", sum(incomes, -1));
  setHtml("current-month-tax", sum(taxes));
  setHtml("current-month-expenses", sum(expenses));
  setHtml("current-month-investment", sum(investments));
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
  const ts = _.groupBy(postings, (p) => p.timestamp.format(timeFormat));

  const points: {
    month: string;
    timestamp: Dayjs;
    [key: string]: number | string | dayjs.Dayjs;
  }[] = [];

  forEachMonth(start, end, (month) => {
    const postings = ts[month.format(timeFormat)] || [];
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
          postings: postings
        },
        defaultValues,
        values
      )
    );
  });

  const x = d3.scaleBand().range([0, width]).paddingInner(0.1).paddingOuter(0);
  const y = d3.scaleLinear().range([height, 0]);

  const sum = (p) => _.sum(_.map(groups, (k) => p[k]));
  x.domain(points.map((p) => p.month));
  y.domain([0, d3.max(points, sum)]);

  const z = d3.scaleOrdinal<string>().range(d3.schemeCategory10);

  g.append("g")
    .attr("class", "axis x")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x)
        .ticks(5)
        .tickFormat(skipTicks(30, x, (d) => d.toString(), points.length))
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
      d3.stack().offset(d3.stackOffsetDiverging).keys(groups)(
        points as { [key: string]: number }[]
      )
    )
    .enter()
    .append("g")
    .attr("fill", function (d) {
      return z(d.key);
    })
    .selectAll("rect")
    .data(function (d) {
      return d;
    })
    .enter()
    .append("rect")
    .on("click", (event, data) => {
      const timestamp: Dayjs = data.data.timestamp as any;
      dateSelector.value = timestamp.format("YYYY-MM");
      dateSelector.dispatchEvent(new Event("input", { bubbles: true }));
    })
    .attr("data-tippy-content", (d) => {
      return tooltip(
        _.flatMap(groups, (key) => {
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
    })
    .attr("x", function (d) {
      return (
        x((d.data as any).month) +
        (x.bandwidth() - Math.min(x.bandwidth(), MAX_BAR_WIDTH)) / 2
      );
    })
    .attr("y", function (d) {
      return y(d[1]);
    })
    .attr("height", function (d) {
      return y(d[0]) - y(d[1]);
    })
    .attr("width", Math.min(x.bandwidth(), MAX_BAR_WIDTH));

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
            .attr("data-tippy-content", (d) => {
              return tooltip(
                d.postings.map((p) => {
                  return [
                    p.timestamp.format("DD MMM YYYY"),
                    [p.payee, "is-clipped"],
                    [
                      formatCurrency(p.amount),
                      "has-text-weight-bold has-text-right"
                    ]
                  ];
                })
              );
            })
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
            .attr("data-tippy-content", (d) => {
              return tooltip(
                d.postings.map((p) => {
                  return [
                    p.timestamp.format("DD MMM YYYY"),
                    [p.payee, "is-clipped"],
                    [
                      formatCurrency(p.amount),
                      "has-text-weight-bold has-text-right"
                    ]
                  ];
                })
              );
            })
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
            .attr("fill", function (d) {
              return z(d.category);
            })
            .attr("class", "is-family-monospace")
            .text(
              (d) =>
                `${formatCurrency(d.total)} ${sprintf(
                  "%6.2f",
                  (d.total / total) * 100
                )}%`
            ),
        (update) =>
          update
            .text(
              (d) =>
                `${formatCurrency(d.total)} ${sprintf(
                  "%6.2f",
                  (d.total / total) * 100
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
