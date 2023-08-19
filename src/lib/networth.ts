import * as d3 from "d3";
import { Delaunay } from "d3";
import legend from "d3-svg-legend";
import dayjs from "dayjs";
import _ from "lodash";
import tippy from "tippy.js";
import COLORS from "./colors";
import { formatCurrency, isMobile } from "./utils";
import { formatCurrencyCrude, tooltip, type Networth } from "./utils";

function networth(d: Networth) {
  return d.investmentAmount + d.gainAmount - d.withdrawalAmount;
}

function investment(d: Networth) {
  return d.investmentAmount - d.withdrawalAmount;
}

export function renderNetworth(points: Networth[], element: Element): () => void {
  const start = _.min(_.map(points, (p) => p.date)),
    end = dayjs();

  const svg = d3.select(element);

  svg.selectAll("*").remove();

  const right = isMobile() ? 10 : 80,
    margin = { top: 40, right: right, bottom: 20, left: 40 },
    width = element.parentElement.clientWidth - margin.left - margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

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
    p.gainAmount + p.investmentAmount - p.withdrawalAmount,
    p.investmentAmount - p.withdrawalAmount
  ]);
  positions.push(0);

  const x = d3.scaleTime().range([0, width]).domain([start, end]),
    y = d3.scaleLinear().range([height, 0]).domain(d3.extent(positions)),
    z = d3.scaleOrdinal<string>(colors).domain(areaKeys);

  const area = (y0: number, y1: (d: Networth) => number) =>
    d3
      .area<Networth>()
      .curve(d3.curveBasis)
      .x((d) => x(d.date))
      .y0(y0)
      .y1(y1);

  g.append("g")
    .attr("class", "axis x")
    .attr("transform", "translate(0," + height + ")")
    .call(d3.axisBottom(x));

  if (!isMobile()) {
    g.append("g")
      .attr("class", "axis y")
      .attr("transform", `translate(${width},0)`)
      .call(d3.axisRight(y).tickPadding(5).tickFormat(formatCurrencyCrude));
  }

  g.append("g")
    .attr("class", "axis y")
    .call(d3.axisLeft(y).tickSize(-width).tickFormat(formatCurrencyCrude));

  const layer = g.selectAll(".layer").data([points]).enter().append("g").attr("class", "layer");

  const clipAboveID = _.uniqueId("clip-above");
  layer
    .append("clipPath")
    .attr("id", clipAboveID)
    .append("path")
    .attr(
      "d",
      area(height, (d) => {
        return y(d.gainAmount + d.investmentAmount - d.withdrawalAmount);
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
        return y(d.gainAmount + d.investmentAmount - d.withdrawalAmount);
      })
    );

  layer
    .append("path")
    .attr("clip-path", `url(${new URL("#" + clipAboveID, window.location.toString())})`)
    .style("fill", z("gain"))
    .style("opacity", "0.8")
    .attr(
      "d",
      area(0, (d) => {
        return y(d.investmentAmount - d.withdrawalAmount);
      })
    );

  layer
    .append("path")
    .attr("clip-path", `url(${new URL("#" + clipBelowID, window.location.toString())})`)
    .style("fill", z("loss"))
    .style("opacity", "0.8")
    .attr(
      "d",
      area(height, (d) => {
        return y(d.investmentAmount - d.withdrawalAmount);
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
        .x((d) => x(d.date))
        .y((d) => y(investment(d)))
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
        .x((d) => x(d.date))
        .y((d) => y(networth(d)))
    );

  const hoverCircle = layer.append("circle").attr("r", "3").attr("fill", "none");
  const t = tippy(hoverCircle.node(), { theme: "light", delay: 0, allowHTML: true });

  const networthVoronoiPoints = _.map(points, (d) => [x(d.date), y(networth(d))]);
  const investmentVoronoiPoints = _.map(points, (d) => [x(d.date), y(investment(d))]);
  const voronoi = Delaunay.from(networthVoronoiPoints.concat(investmentVoronoiPoints)).voronoi([
    0,
    0,
    width,
    height
  ]);

  layer
    .append("g")
    .selectAll("path")
    .data(
      points.map((p) => ["networth", p]).concat(points.map((p) => ["investment", p])) as [
        string,
        Networth
      ][]
    )
    .enter()
    .append("path")
    .style("pointer-events", "all")
    .style("fill", "none")
    .attr("d", (_, i) => {
      return voronoi.renderCell(i);
    })
    .on("mouseover", (_, [pointType, d]) => {
      hoverCircle
        .attr("cx", x(d.date))
        .attr("cy", y(pointType == "networth" ? networth(d) : investment(d)))
        .attr("fill", lineScale(pointType));

      t.setProps({
        placement: pointType == "networth" ? "top" : "bottom",
        content: tooltip([
          ["Date", d.date.format("DD MMM YYYY")],
          ["Net Worth", [formatCurrency(networth(d)), "has-text-weight-bold has-text-right"]],
          [
            "Net Investment",
            [formatCurrency(investment(d)), "has-text-weight-bold has-text-right"]
          ],
          ["Gain / Loss", [formatCurrency(d.gainAmount), "has-text-weight-bold has-text-right"]]
        ])
      });
      t.show();
    })
    .on("mouseout", () => {
      t.hide();
      hoverCircle.attr("fill", "none");
    });

  svg.append("g").attr("class", "legendOrdinal").attr("transform", "translate(265,3)");

  const legendOrdinal = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(50)
    .labels(areaKeys)
    .scale(areaScale);

  svg.select(".legendOrdinal").call(legendOrdinal as any);

  svg.append("g").attr("class", "legendLine").attr("transform", "translate(80,3)");

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

  return () => {
    t.destroy();
  };
}
