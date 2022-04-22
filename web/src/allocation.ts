import * as d3 from "d3";
import legend from "d3-svg-legend";
import dayjs from "dayjs";
import _ from "lodash";
import {
  Aggregate,
  ajax,
  formatCurrency,
  formatFloat,
  lastName,
  parentName,
  rainbowScale,
  secondName,
  textColor,
  tooltip
} from "./utils";

export default async function () {
  const { aggregates: aggregates, aggregates_timeline: aggregatesTimeline } =
    await ajax("/api/allocation");
  _.each(aggregates, (a) => (a.timestamp = dayjs(a.date)));
  _.each(aggregatesTimeline, (aggregates) =>
    _.each(aggregates, (a) => (a.timestamp = dayjs(a.date)))
  );
  renderAllocation(aggregates);
  renderAllocationTimeline(aggregatesTimeline);
}

function renderAllocation(aggregates: { [key: string]: Aggregate }) {
  const allocation = function (id, hierarchy) {
    const div = d3.select("#" + id),
      margin = { top: 0, right: 0, bottom: 0, left: 20 },
      width =
        document.getElementById(id).parentElement.clientWidth -
        margin.left -
        margin.right,
      height = +div.attr("height") - margin.top - margin.bottom;

    const percent = (d) => {
      return formatFloat((d.value / root.value) * 100) + "%";
    };

    const color = rainbowScale(_.keys(aggregates));

    const stratify = d3
      .stratify<Aggregate>()
      .id((d) => d.account)
      .parentId((d) => parentName(d.account));

    const partition = hierarchy.size([width, height]).round(true);

    const root = stratify(_.sortBy(aggregates, (a) => a.account))
      .sum((a) => a.market_amount)
      .sort(function (a, b) {
        return b.height - a.height || b.value - a.value;
      });

    partition(root);

    const cell = div
      .selectAll(".node")
      .data(root.descendants())
      .enter()
      .append("div")
      .attr("class", "node")
      .attr("data-tippy-content", (d) => {
        return tooltip([
          ["Account", [d.id, "has-text-right"]],
          [
            "MarketAmount",
            [formatCurrency(d.value), "has-text-weight-bold has-text-right"]
          ],
          ["Percentage", [percent(d), "has-text-weight-bold has-text-right"]]
        ]);
      })
      .style("top", (d: any) => d.y0 + "px")
      .style("left", (d: any) => d.x0 + "px")
      .style("width", (d: any) => d.x1 - d.x0 + "px")
      .style("height", (d: any) => d.y1 - d.y0 + "px")
      .style("background", (d) => color(d.id))
      .style("color", (d) => textColor(color(d.id)));

    cell
      .append("p")
      .attr("class", "heading has-text-weight-bold")
      .text((d) => lastName(d.id));

    cell
      .append("p")
      .attr("class", "heading has-text-weight-bold")
      .style("font-size", ".5 rem")
      .text(percent);
  };

  allocation("d3-allocation-category", d3.partition());
  allocation("d3-allocation-value", d3.treemap());
}

function renderAllocationTimeline(
  aggregatesTimeline: { [key: string]: Aggregate }[]
) {
  const timeline = _.map(aggregatesTimeline, (aggregates) => {
    return _.chain(aggregates)
      .values()
      .filter((a) => a.amount != 0)
      .groupBy((a) => secondName(a.account))
      .map((aggregates, group) => {
        return {
          date: aggregates[0].date,
          account: group,
          amount: _.sum(_.map(aggregates, (a) => a.amount)),
          market_amount: _.sum(_.map(aggregates, (a) => a.market_amount)),
          timestamp: aggregates[0].timestamp
        };
      })
      .value();
  });
  const assets = _.chain(timeline)
    .last()
    .map((a) => a.account)
    .sort()
    .value();

  const defaultValues = _.zipObject(
    assets,
    _.map(assets, () => 0)
  );
  const start = timeline[0][0].timestamp,
    end = dayjs();

  interface Point {
    date: dayjs.Dayjs;
    [key: string]: number | dayjs.Dayjs;
  }
  const points: Point[] = [];
  _.each(timeline, (aggregates) => {
    const total = _.sum(_.map(aggregates, (a) => a.market_amount));
    if (total == 0) {
      return;
    }
    const kvs = _.map(aggregates, (a) => [
      a.account,
      (a.market_amount / total) * 100
    ]);
    points.push(
      _.merge(
        {
          date: aggregates[0].timestamp
        },
        defaultValues,
        _.fromPairs(kvs)
      )
    );
  });

  const svg = d3.select("#d3-allocation-timeline"),
    margin = { top: 40, right: 50, bottom: 20, left: 30 },
    width =
      document.getElementById("d3-allocation-timeline").parentElement
        .clientWidth -
      margin.left -
      margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg
      .append("g")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const x = d3.scaleTime().range([0, width]).domain([start, end]),
    y = d3.scaleLinear().range([height, 0]).domain([0, 100]),
    z = d3.scaleOrdinal(d3.schemeCategory10).domain(assets);

  const stack = d3.stack<Point>().keys(assets);
  const series = stack(points);

  const area = d3
    .area<d3.SeriesPoint<Point>>()
    .x(function (d) {
      return x(d.data.date);
    })
    .y0(function (d) {
      return y(d[0]);
    })
    .y1(function (d) {
      return y(d[1]);
    });

  g.append("g")
    .attr("class", "axis x")
    .attr("transform", "translate(0," + height + ")")
    .call(d3.axisBottom(x));

  g.append("g").attr("class", "axis y").call(d3.axisLeft(y));
  g.append("g")
    .attr("class", "axis y")
    .attr("transform", `translate(${width},0)`)
    .call(d3.axisRight(y));

  const layer = g
    .selectAll(".layer")
    .data(series)
    .enter()
    .append("g")
    .attr("class", "layer");

  layer
    .append("path")
    .attr("class", "area")
    .style("fill", function (d) {
      return z(d.key);
    })
    .attr("d", area);

  layer
    .filter(function (d) {
      return d[d.length - 1][1] - d[d.length - 1][0] > 0.01;
    })
    .append("text")
    .attr("x", width - 6)
    .attr("y", function (d) {
      return y((d[d.length - 1][0] + d[d.length - 1][1]) / 2);
    });

  svg
    .append("g")
    .attr("class", "legendOrdinal")
    .attr("transform", "translate(40,0)");

  var legendOrdinal = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(100)
    .labels(assets)
    .scale(z);

  svg.select(".legendOrdinal").call(legendOrdinal as any);
}
