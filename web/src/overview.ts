import * as d3 from "d3";
import legend from "d3-svg-legend";
import dayjs from "dayjs";
import _ from "lodash";
import COLORS from "./colors";
import {
  ajax,
  formatCurrency,
  formatCurrencyCrude,
  formatFloat,
  Overview,
  setHtml
} from "./utils";

export default async function () {
  const { overview_timeline: points, xirr: xirr } = await ajax("/api/overview");
  _.each(points, (n) => (n.timestamp = dayjs(n.date)));

  const current = _.last(points);
  setHtml(
    "networth",
    formatCurrency(
      current.investment_amount +
        current.gain_amount -
        current.withdrawal_amount
    ),
    COLORS.primary
  );
  setHtml(
    "investment",
    formatCurrency(current.investment_amount - current.withdrawal_amount),
    COLORS.secondary
  );
  setHtml(
    "gains",
    formatCurrency(current.gain_amount),
    current.gain_amount >= 0 ? COLORS.gainText : COLORS.lossText
  );
  setHtml("xirr", formatFloat(xirr));

  renderOverview(points, document.getElementById("d3-overview-timeline"));
}

function renderOverview(points: Overview[], element: Element) {
  const start = _.min(_.map(points, (p) => p.timestamp)),
    end = dayjs();

  const svg = d3.select(element),
    margin = { top: 40, right: 80, bottom: 20, left: 40 },
    width = element.parentElement.clientWidth - margin.left - margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg
      .append("g")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const areaKeys = ["gain", "loss"];
  const colors = [COLORS.gain, COLORS.loss];
  const areaScale = d3.scaleOrdinal().domain(areaKeys).range(colors);

  const lineKeys = ["networth", "investment"];
  const labels = ["Net Worth", "Net Investment"];
  const lineScale = d3
    .scaleOrdinal<string>()
    .domain(lineKeys)
    .range([COLORS.primary, COLORS.secondary]);

  const positions = _.flatMap(points, (p) => [
    p.gain_amount + p.investment_amount - p.withdrawal_amount,
    p.investment_amount - p.withdrawal_amount
  ]);
  positions.push(0);

  const x = d3.scaleTime().range([0, width]).domain([start, end]),
    y = d3.scaleLinear().range([height, 0]).domain(d3.extent(positions)),
    z = d3.scaleOrdinal<string>(colors).domain(areaKeys);

  const area = (y0, y1) =>
    d3
      .area<Overview>()
      .curve(d3.curveBasis)
      .x((d) => x(d.timestamp))
      .y0(y0)
      .y1(y1);

  g.append("g")
    .attr("class", "axis x")
    .attr("transform", "translate(0," + height + ")")
    .call(d3.axisBottom(x));

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", `translate(${width},0)`)
    .call(d3.axisRight(y).tickPadding(5).tickFormat(formatCurrencyCrude));

  g.append("g")
    .attr("class", "axis y")
    .call(d3.axisLeft(y).tickSize(-width).tickFormat(formatCurrencyCrude));

  const layer = g
    .selectAll(".layer")
    .data([points])
    .enter()
    .append("g")
    .attr("class", "layer");

  const clipAboveID = _.uniqueId("clip-above");
  layer
    .append("clipPath")
    .attr("id", clipAboveID)
    .append("path")
    .attr(
      "d",
      area(height, (d) => {
        return y(d.gain_amount + d.investment_amount - d.withdrawal_amount);
      })
    );

  const clipBelowID = _.uniqueId("clip-below");
  layer
    .append("clipPath")
    .attr("id", clipBelowID)
    .append("path")
    .attr(
      "d",
      area(0, (d) => {
        return y(d.gain_amount + d.investment_amount - d.withdrawal_amount);
      })
    );

  layer
    .append("path")
    .attr(
      "clip-path",
      `url(${new URL("#" + clipAboveID, window.location.toString())})`
    )
    .style("fill", z("gain"))
    .style("opacity", "0.8")
    .attr(
      "d",
      area(0, (d) => {
        return y(d.investment_amount - d.withdrawal_amount);
      })
    );

  layer
    .append("path")
    .attr(
      "clip-path",
      `url(${new URL("#" + clipBelowID, window.location.toString())})`
    )
    .style("fill", z("loss"))
    .style("opacity", "0.8")
    .attr(
      "d",
      area(height, (d) => {
        return y(d.investment_amount - d.withdrawal_amount);
      })
    );

  layer
    .append("path")
    .style("stroke", lineScale("investment"))
    .style("fill", "none")
    .attr(
      "d",
      d3
        .line<Overview>()
        .curve(d3.curveBasis)
        .x((d) => x(d.timestamp))
        .y((d) => y(d.investment_amount - d.withdrawal_amount))
    );

  layer
    .append("path")
    .style("stroke", lineScale("networth"))
    .style("fill", "none")
    .attr(
      "d",
      d3
        .line<Overview>()
        .curve(d3.curveBasis)
        .x((d) => x(d.timestamp))
        .y((d) => y(d.investment_amount + d.gain_amount - d.withdrawal_amount))
    );

  svg
    .append("g")
    .attr("class", "legendOrdinal")
    .attr("transform", "translate(265,3)");

  const legendOrdinal = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(50)
    .labels(areaKeys)
    .scale(areaScale);

  svg.select(".legendOrdinal").call(legendOrdinal as any);

  svg
    .append("g")
    .attr("class", "legendLine")
    .attr("transform", "translate(80,3)");

  const legendLine = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(70)
    .labelOffset(22)
    .shapeHeight(3)
    .shapeWidth(25)
    .labels(labels)
    .scale(lineScale);

  svg.select(".legendLine").call(legendLine as any);
}
