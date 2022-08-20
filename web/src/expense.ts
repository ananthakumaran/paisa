import * as d3 from "d3";
import legend from "d3-svg-legend";
import dayjs from "dayjs";
import _ from "lodash";
import {
  ajax,
  forEachMonth,
  formatCurrency,
  formatCurrencyCrude,
  Posting,
  secondName,
  skipTicks,
  tooltip
} from "./utils";

export default async function () {
  const { expenses: expenses } = await ajax("/api/expense");
  _.each(expenses, (p) => (p.timestamp = dayjs(p.date)));
  renderMonthlyExpensesTimeline(expenses);
}

function renderMonthlyExpensesTimeline(postings: Posting[]) {
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
}
