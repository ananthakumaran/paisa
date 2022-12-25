import * as d3 from "d3";
import legend from "d3-svg-legend";
import type dayjs from "dayjs";
import _ from "lodash";
import {
  formatCurrency,
  formatCurrencyCrude,
  generateColorScheme,
  type Income,
  type Posting,
  restName,
  skipTicks,
  tooltip
} from "./utils";

export function renderMonthlyInvestmentTimeline(incomes: Income[]) {
  renderIncomeTimeline(incomes, "#d3-income-timeline", "MMM-YYYY");
}

function renderIncomeTimeline(incomes: Income[], id: string, timeFormat: string) {
  const MAX_BAR_WIDTH = 40;
  const svg = d3.select(id),
    margin = { top: 60, right: 30, bottom: 80, left: 40 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const postings = _.flatMap(incomes, (i) => i.postings);
  const groupKeys = _.chain(postings)
    .map((p) => restName(p.account))
    .uniq()
    .sort()
    .value();

  const groupTotal = _.chain(postings)
    .groupBy((p) => restName(p.account))
    .map((postings, key) => {
      const total = _.sumBy(postings, (p) => -p.amount);
      return [key, `${key} ${formatCurrency(total)}`];
    })
    .fromPairs()
    .value();

  const defaultValues = _.zipObject(
    groupKeys,
    _.map(groupKeys, () => 0)
  );

  let points: {
    date: dayjs.Dayjs;
    month: string;
    [key: string]: number | string | dayjs.Dayjs;
  }[] = [];

  points = _.map(incomes, (i) => {
    const values = _.chain(i.postings)
      .groupBy((p) => restName(p.account))
      .flatMap((postings, key) => [[key, _.sumBy(postings, (p) => -p.amount)]])
      .fromPairs()
      .value();

    return _.merge(
      {
        month: i.timestamp.format(timeFormat),
        date: i.timestamp,
        postings: i.postings
      },
      defaultValues,
      values
    );
  });

  const x = d3.scaleBand().range([0, width]).paddingInner(0.1).paddingOuter(0);
  const y = d3.scaleLinear().range([height, 0]);

  const sum = (filter) => (p) =>
    _.sum(
      _.filter(
        _.map(groupKeys, (k) => p[k]),
        filter
      )
    );
  x.domain(points.map((p) => p.month));
  y.domain([
    d3.min(
      points,
      sum((a) => a < 0)
    ),
    d3.max(
      points,
      sum((a) => a > 0)
    )
  ]);

  const z = generateColorScheme(groupKeys);

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
      d3.stack().offset(d3.stackOffsetDiverging).keys(groupKeys)(
        points as { [key: string]: number }[]
      )
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
      return tooltip(
        _.sortBy(
          postings.map((p) => [
            restName(p.account),
            [formatCurrency(-p.amount), "has-text-weight-bold has-text-right"]
          ]),
          (r) => r[0]
        )
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

  const LEGEND_PADDING = 100;

  svg
    .append("g")
    .attr("class", "legendOrdinal")
    .attr("transform", `translate(${LEGEND_PADDING / 2},0)`);

  const legendOrdinal = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(LEGEND_PADDING)
    .labels(({ i }) => {
      return groupTotal[groupKeys[i]];
    })
    .scale(z);

  (legendOrdinal as any).labelWrap(75); // type missing

  svg.select(".legendOrdinal").call(legendOrdinal as any);
}
