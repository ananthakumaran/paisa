import * as d3 from "d3";
import legend from "d3-svg-legend";
import dayjs from "dayjs";
import _ from "lodash";
import {
  ajax,
  forEachMonth,
  forEachYear,
  formatCurrency,
  formatCurrencyCrude,
  Posting,
  secondName,
  skipTicks,
  tooltip
} from "./utils";

export default async function () {
  const { postings: postings } = await ajax("/api/investment");
  _.each(postings, (p) => (p.timestamp = dayjs(p.date)));
  renderMonthlyInvestmentTimeline(postings);
  renderYearlyInvestmentTimeline(postings);
}

function renderMonthlyInvestmentTimeline(postings: Posting[]) {
  renderInvestmentTimeline(
    postings,
    "#d3-investment-timeline",
    "MMM-YYYY",
    true,
    forEachMonth
  );
}

function renderYearlyInvestmentTimeline(postings: Posting[]) {
  renderInvestmentTimeline(
    postings,
    "#d3-yearly-investment-timeline",
    "YYYY",
    false,
    forEachYear
  );
}

function renderInvestmentTimeline(
  postings: Posting[],
  id: string,
  timeFormat: string,
  showTooltip: boolean,
  iterator: (
    start: dayjs.Dayjs,
    end: dayjs.Dayjs,
    cb: (current: dayjs.Dayjs) => any
  ) => any
) {
  const MAX_BAR_WIDTH = 40;
  const svg = d3.select(id),
    margin = { top: 40, right: 30, bottom: 80, left: 40 },
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
  const groupKeys = _.flatMap(groups, (g) => [g + "-credit", g + "-debit"]);

  const defaultValues = _.zipObject(
    groupKeys,
    _.map(groupKeys, () => 0)
  );

  const start = _.min(_.map(postings, (p) => p.timestamp)),
    end = dayjs().startOf("month");
  const ts = _.groupBy(postings, (p) => p.timestamp.format(timeFormat));

  const points: {
    date: dayjs.Dayjs;
    month: string;
    [key: string]: number | string | dayjs.Dayjs;
  }[] = [];

  iterator(start, end, (month) => {
    postings = ts[month.format(timeFormat)] || [];
    const values = _.chain(ts[month.format(timeFormat)] || [])
      .groupBy((t) => secondName(t.account))
      .flatMap((postings, key) => [
        [
          key + "-credit",
          _.sum(
            _.filter(
              _.map(postings, (p) => p.amount),
              (a) => a >= 0
            )
          )
        ],
        [
          key + "-debit",
          _.sum(
            _.filter(
              _.map(postings, (p) => p.amount),
              (a) => a < 0
            )
          )
        ]
      ])
      .fromPairs()
      .value();

    points.push(
      _.merge(
        {
          month: month.format(timeFormat),
          date: month,
          postings: postings
        },
        defaultValues,
        values
      )
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
      if (!showTooltip) {
        return null;
      }
      const postings: Posting[] = (d.data as any).postings;
      return tooltip(
        _.sortBy(
          postings.map((p) => [
            _.drop(p.account.split(":")).join(":"),
            [formatCurrency(p.amount), "has-text-weight-bold has-text-right"]
          ]),
          (r) => r[0]
        )
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
