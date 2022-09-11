import $ from "jquery";
import * as d3 from "d3";
import legend from "d3-svg-legend";
import dayjs from "dayjs";
import _ from "lodash";
import {
  Aggregate,
  ajax,
  AllocationTarget,
  formatCurrency,
  formatFloat,
  lastName,
  parentName,
  secondName,
  textColor,
  tooltip,
  skipTicks,
  generateColorScheme
} from "./utils";
import COLORS from "./colors";

export default async function () {
  const {
    aggregates: aggregates,
    aggregates_timeline: aggregatesTimeline,
    allocation_targets: allocationTargets
  } = await ajax("/api/allocation");
  _.each(aggregates, (a) => (a.timestamp = dayjs(a.date)));
  _.each(aggregatesTimeline, (aggregates) =>
    _.each(aggregates, (a) => (a.timestamp = dayjs(a.date)))
  );

  const color = generateColorScheme(_.keys(aggregates));

  renderAllocationTarget(allocationTargets, color);
  renderAllocation(aggregates, color);
  renderAllocationTimeline(aggregatesTimeline);
}

function renderAllocationTarget(
  allocationTargets: AllocationTarget[],
  color: d3.ScaleOrdinal<string, string>
) {
  const id = "#d3-allocation-target";

  if (_.isEmpty(allocationTargets)) {
    $(id).closest(".container").hide();
    return;
  }
  allocationTargets = _.sortBy(allocationTargets, (t) => t.name);
  const BAR_HEIGHT = 25;
  const svg = d3.select(id),
    margin = { top: 20, right: 0, bottom: 10, left: 150 },
    fullWidth = document.getElementById(id.substring(1)).parentElement
      .clientWidth,
    width = fullWidth - margin.left - margin.right,
    height = allocationTargets.length * BAR_HEIGHT * 2,
    g = svg
      .append("g")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")");
  svg.attr("height", height + margin.top + margin.bottom);

  const keys = ["target", "current"];
  const colorKeys = ["target", "current", "diff"];
  const colors = [COLORS.primary, COLORS.secondary, COLORS.diff];

  const y = d3.scaleBand().range([0, height]).paddingInner(0).paddingOuter(0);
  y.domain(allocationTargets.map((t) => t.name));

  const y1 = d3
    .scaleBand()
    .range([0, y.bandwidth()])
    .domain(keys)
    .paddingInner(0)
    .paddingOuter(0.1);

  const z = d3.scaleOrdinal<string>(colors).domain(colorKeys);

  const maxX = _.chain(allocationTargets)
    .flatMap((t) => [t.current, t.target])
    .max()
    .value();
  const targetWidth = 400;
  const targetMargin = 20;
  const textGroupWidth = 150;
  const textGroupMargin = 20;
  const textGroupZero = targetWidth + targetMargin;

  const x = d3
    .scaleLinear()
    .range([textGroupZero + textGroupWidth + textGroupMargin, width]);
  x.domain([0, maxX]);
  const x1 = d3.scaleLinear().range([0, targetWidth]).domain([0, maxX]);

  g.append("line")
    .attr("stroke", "#ddd")
    .attr("x1", 0)
    .attr("y1", height)
    .attr("x2", width)
    .attr("y2", height);

  g.append("text")
    .attr("fill", "#4a4a4a")
    .text("Target")
    .attr("text-anchor", "end")

    .attr("x", textGroupZero + (textGroupWidth * 1) / 3)
    .attr("y", -5);

  g.append("text")
    .attr("fill", "#4a4a4a")
    .text("Current")
    .attr("text-anchor", "end")
    .attr("x", textGroupZero + (textGroupWidth * 2) / 3)
    .attr("y", -5);

  g.append("text")
    .attr("fill", "#4a4a4a")
    .text("Diff")
    .attr("text-anchor", "end")
    .attr("x", textGroupZero + textGroupWidth)
    .attr("y", -5);

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x1)
        .tickSize(-height)
        .tickFormat(skipTicks(40, x, (n) => formatFloat(n, 0)))
    );

  g.append("g").attr("class", "axis y dark").call(d3.axisLeft(y));

  const textGroup = g
    .append("g")
    .selectAll("g")
    .data(allocationTargets)
    .enter()
    .append("g")
    .attr("class", "inline-text");

  textGroup
    .append("line")
    .attr("stroke", "#ddd")
    .attr("x1", 0)
    .attr("y1", (t) => y(t.name))
    .attr("x2", width)
    .attr("y2", (t) => y(t.name));

  textGroup
    .append("text")
    .text((t) => formatFloat(t.target))
    .attr("text-anchor", "end")
    .attr("alignment-baseline", "middle")
    .style("fill", z("target"))
    .attr("x", textGroupZero + (textGroupWidth * 1) / 3)
    .attr("y", (t) => y(t.name) + y.bandwidth() / 2);

  textGroup
    .append("text")
    .text((t) => formatFloat(t.current))
    .attr("text-anchor", "end")
    .attr("alignment-baseline", "middle")
    .style("fill", z("current"))
    .attr("x", textGroupZero + (textGroupWidth * 2) / 3)
    .attr("y", (t) => y(t.name) + y.bandwidth() / 2);

  textGroup
    .append("text")
    .text((t) => formatFloat(t.current - t.target))
    .attr("text-anchor", "end")
    .attr("alignment-baseline", "middle")
    .style("fill", z("diff"))
    .attr("x", textGroupZero + (textGroupWidth * 3) / 3)
    .attr("y", (t) => y(t.name) + y.bandwidth() / 2);

  const groups = g
    .append("g")
    .selectAll("g.group")
    .data(allocationTargets)
    .enter()
    .append("g")
    .attr("class", "group")
    .attr("transform", (t) => "translate(0," + y(t.name) + ")");

  groups
    .selectAll("g")
    .data((t) => [
      { key: "target", value: t.target },
      { key: "current", value: t.current }
    ])
    .enter()
    .append("rect")
    .attr("fill", (d) => {
      return z(d.key);
    })
    .attr("x", x1(0))
    .attr("y", (d) => y1(d.key))
    .attr("height", y1.bandwidth())
    .attr("width", (d) => x1(d.value));

  const paddingTop = (y1.range()[1] - y1.bandwidth() * 2) / 2;
  d3.select("#d3-allocation-target-treemap")
    .append("div")
    .style("height", height + margin.top + margin.bottom + "px")
    .style("position", "absolute")
    .style("width", "100%")
    .selectAll("div")
    .data(allocationTargets)
    .enter()
    .append("div")
    .style("position", "absolute")
    .style("left", margin.left + x(0) + "px")
    .style("top", (t) => margin.top + y(t.name) + paddingTop + "px")
    .style("height", y1.bandwidth() * 2 + "px")
    .style("width", x.range()[1] - x.range()[0] + "px")
    .append("div")
    .style("position", "relative")
    .attr("height", y1.bandwidth() * 2)
    .each(function (t) {
      renderPartition(this, t.aggregates, d3.treemap(), color);
    });
}

function renderAllocation(
  aggregates: { [key: string]: Aggregate },
  color: d3.ScaleOrdinal<string, string>
) {
  renderPartition(
    document.getElementById("d3-allocation-category"),
    aggregates,
    d3.partition(),
    color
  );
  renderPartition(
    document.getElementById("d3-allocation-value"),
    aggregates,
    d3.treemap(),
    color
  );
}

function renderPartition(
  element: HTMLElement,
  aggregates,
  hierarchy,
  color: d3.ScaleOrdinal<string, string>
) {
  if (_.isEmpty(aggregates)) {
    return;
  }

  const div = d3.select(element),
    margin = { top: 0, right: 0, bottom: 0, left: 20 },
    width = element.parentElement.clientWidth - margin.left - margin.right,
    height = +div.attr("height") - margin.top - margin.bottom;

  const percent = (d) => {
    return formatFloat((d.value / root.value) * 100) + "%";
  };

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
    margin = { top: 40, right: 60, bottom: 20, left: 35 },
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
    y = d3
      .scaleLinear()
      .range([height, 0])
      .domain([
        0,
        d3.max(d3.map(points, (p) => d3.max(_.values(_.omit(p, "date")))))
      ]),
    z = generateColorScheme(assets);

  const line = (group) =>
    d3
      .line<Point>()
      .curve(d3.curveLinear)
      .defined((p, i) => p[group] > 0 || points[i + 1]?.[group] > 0)
      .x((p) => x(p.date))
      .y((p) => y(p[group]));

  g.append("g")
    .attr("class", "axis x")
    .attr("transform", "translate(0," + height + ")")
    .call(d3.axisBottom(x));

  g.append("g")
    .attr("class", "axis y")
    .call(
      d3
        .axisLeft(y)
        .tickSize(-width)
        .tickFormat((y) => `${y}%`)
    );
  g.append("g")
    .attr("class", "axis y")
    .attr("transform", `translate(${width},0)`)
    .call(d3.axisRight(y).tickFormat((y) => `${y}%`));

  const layer = g
    .selectAll(".layer")
    .data(assets)
    .enter()
    .append("g")
    .attr("class", "layer");

  layer
    .append("path")
    .attr("fill", "none")
    .attr("stroke", (group) => z(group))
    .attr("stroke-width", "2")
    .attr("d", (group) => line(group)(points));

  svg
    .append("g")
    .attr("class", "legendOrdinal")
    .attr("transform", "translate(40,0)");

  const legendOrdinal = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(100)
    .labels(assets)
    .scale(z);

  svg.select(".legendOrdinal").call(legendOrdinal as any);
}
