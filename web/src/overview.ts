import * as d3 from "d3";
window.d3 = d3;
import legend from "d3-svg-legend";
import dayjs from "dayjs";
import _ from "lodash";
import {
  ajax,
  formatCurrency,
  formatCurrencyCrude,
  Networth,
  setHtml
} from "./utils";

export default async function () {
  const { networth_timeline: points } = await ajax("/api/overview");
  _.each(points, (n) => (n.timestamp = dayjs(n.date)));

  const current = _.last(points);
  setHtml(
    "networth",
    formatCurrency(
      current.investment_amount +
        current.gain_amount -
        current.withdrawal_amount
    )
  );
  setHtml("investment", formatCurrency(current.investment_amount));
  setHtml("withdrawal", formatCurrency(current.withdrawal_amount));
  setHtml("gains", formatCurrency(current.gain_amount));

  const start = _.min(_.map(points, (p) => p.timestamp)),
    end = dayjs();

  const svg = d3.select("#d3-networth-timeline"),
    margin = { top: 40, right: 80, bottom: 20, left: 40 },
    width =
      document.getElementById("d3-networth-timeline").parentElement
        .clientWidth -
      margin.left -
      margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg
      .append("g")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const areaKeys = ["gain", "loss"];
  const colors = ["#b2df8a", "#fb9a99"];
  const areaScale = d3.scaleOrdinal().domain(areaKeys).range(colors);

  const lineKeys = ["networth", "investment", "withdrawal"];
  const lineScale = d3
    .scaleOrdinal<string>()
    .domain(lineKeys)
    .range(["#1f77b4", "#17becf", "#ff7f0e"]);

  svg
    .append("g")
    .attr("class", "legendOrdinal")
    .attr("transform", "translate(365,3)");

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
    .labels(lineKeys)
    .scale(lineScale);

  svg.select(".legendLine").call(legendLine as any);

  const x = d3.scaleTime().range([0, width]).domain([start, end]),
    y = d3
      .scaleLinear()
      .range([height, 0])
      .domain([
        0,
        d3.max<Networth, number>(
          points,
          (d) => d.gain_amount + d.investment_amount
        )
      ]),
    z = d3.scaleOrdinal<string>(colors).domain(areaKeys);

  let area = (y0, y1) =>
    d3
      .area<Networth>()
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
    .call(
      d3
        .axisRight(y)
        .tickPadding(5)
        .tickSize(10)
        .tickFormat(formatCurrencyCrude)
    );

  g.append("g")
    .attr("class", "axis y")
    .call(d3.axisLeft(y).tickSize(-width).tickFormat(formatCurrencyCrude));

  const layer = g
    .selectAll(".layer")
    .data([points])
    .enter()
    .append("g")
    .attr("class", "layer");

  layer
    .append("clipPath")
    .attr("id", `clip-above`)
    .append("path")
    .attr(
      "d",
      area(height, (d) => {
        return y(d.gain_amount + d.investment_amount);
      })
    );

  layer
    .append("clipPath")
    .attr("id", `clip-below`)
    .append("path")
    .attr(
      "d",
      area(0, (d) => {
        return y(d.gain_amount + d.investment_amount);
      })
    );

  layer
    .append("path")
    .attr(
      "clip-path",
      `url(${new URL("#clip-above", window.location.toString())})`
    )
    .style("fill", z("gain"))
    .style("opacity", "0.8")
    .attr(
      "d",
      area(0, (d) => {
        return y(d.investment_amount);
      })
    );

  layer
    .append("path")
    .attr(
      "clip-path",
      `url(${new URL("#clip-below", window.location.toString())})`
    )
    .style("fill", z("loss"))
    .style("opacity", "0.8")
    .attr(
      "d",
      area(height, (d) => {
        return y(d.investment_amount);
      })
    );

  layer
    .append("path")
    .style("stroke", lineScale("investment"))
    .style("fill", "none")
    .attr(
      "d",
      d3
        .line<Networth>()
        .curve(d3.curveBasis)
        .x((d) => x(d.timestamp))
        .y((d) => y(d.investment_amount))
    );

  layer
    .append("path")
    .style("stroke", lineScale("withdrawal"))
    .style("fill", "none")
    .attr(
      "d",
      d3
        .line<Networth>()
        .curve(d3.curveBasis)
        .defined((d) => d.withdrawal_amount > 0)
        .x((d) => x(d.timestamp))
        .y((d) => y(d.withdrawal_amount))
    );

  layer
    .append("path")
    .style("stroke", lineScale("networth"))
    .style("fill", "none")
    .attr(
      "d",
      d3
        .line<Networth>()
        .curve(d3.curveBasis)
        .x((d) => x(d.timestamp))
        .y((d) => y(d.investment_amount + d.gain_amount - d.withdrawal_amount))
    );
}
