import * as d3 from "d3";
import _ from "lodash";
import {
  formatCurrency,
  formatFloat,
  textColor,
  tooltip,
  skipTicks,
  generateColorScheme,
  type PortfolioAggregate,
  type CommodityBreakdown
} from "./utils";
import COLORS from "./colors";

export function renderPortfolioBreakdown(portfolioAggregates: PortfolioAggregate[]) {
  const id = "#d3-portfolio";

  if (_.isEmpty(portfolioAggregates)) {
    return;
  }

  const commodityNames = _.chain(portfolioAggregates)
    .flatMap((pa) => pa.breakdowns)
    .map((b) => b.commodity_name)
    .uniq()
    .value();

  const byID: Record<string, PortfolioAggregate> = _.chain(portfolioAggregates)
    .map((p) => [p.id, p])
    .fromPairs()
    .value();

  const color = generateColorScheme(commodityNames);

  const BAR_HEIGHT = 25;
  const svg = d3.select(id),
    margin = { top: 20, right: 0, bottom: 10, left: 350 },
    fullWidth = document.getElementById(id.substring(1)).parentElement.clientWidth,
    width = fullWidth - margin.left - margin.right,
    height = portfolioAggregates.length * BAR_HEIGHT,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");
  svg.attr("height", height + margin.top + margin.bottom);

  const y = d3.scaleBand().range([0, height]).paddingInner(0.1).paddingOuter(0);
  y.domain(portfolioAggregates.map((t) => t.id));

  const maxX = _.chain(portfolioAggregates)
    .flatMap((t) => [t.percentage])
    .max()
    .value();
  const targetWidth = 400;
  const targetMargin = 20;
  const textGroupWidth = 150;
  const textGroupMargin = 20;
  const textGroupZero = targetWidth + targetMargin;

  const x = d3.scaleLinear().range([textGroupZero + textGroupWidth + textGroupMargin, width]);
  x.domain([0, maxX]);
  const x1 = d3.scaleLinear().range([0, targetWidth]).domain([0, maxX]);

  const paddingTop = (BAR_HEIGHT - y.bandwidth()) / 2;

  svg
    .append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")")
    .selectAll("rect")
    .data(portfolioAggregates)
    .enter()
    .append("rect")
    .attr("fill", COLORS.primary)
    .attr("data-tippy-content", "")
    .attr("x", x1(0))
    .attr("y", function (d) {
      return y(d.id) + paddingTop;
    })
    .attr("width", function (d) {
      return x1(d.percentage);
    })
    .attr("height", y.bandwidth());

  g.append("line")
    .attr("stroke", "#ddd")
    .attr("x1", 0)
    .attr("y1", height)
    .attr("x2", width)
    .attr("y2", height);

  g.append("text")
    .attr("fill", "#4a4a4a")
    .text("%")
    .attr("text-anchor", "end")

    .attr("x", textGroupZero + textGroupWidth / 2)
    .attr("y", -5);

  g.append("text")
    .attr("fill", "#4a4a4a")
    .text("Amount")
    .attr("text-anchor", "end")
    .attr("x", textGroupZero + textGroupWidth)
    .attr("y", -5);

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisTop(x1)
        .tickSize(height)
        .tickFormat(skipTicks(40, x, (n: number) => formatFloat(n, 0)))
    );

  g.append("g")
    .attr("class", "axis y dark")
    .call(d3.axisLeft(y).tickFormat((id) => formatCommodityName(byID[id].name)));

  const textGroup = g
    .append("g")
    .selectAll("g")
    .data(portfolioAggregates)
    .enter()
    .append("g")
    .attr("class", "inline-text");

  textGroup
    .append("line")
    .attr("stroke", "#ddd")
    .attr("x1", 0)
    .attr("y1", (t) => y(t.id))
    .attr("x2", width)
    .attr("y2", (t) => y(t.id));

  textGroup
    .append("text")
    .text((t) => formatFloat(t.percentage))
    .attr("text-anchor", "end")
    .attr("dominant-baseline", "middle")
    .style("fill", COLORS.primary)
    .attr("x", textGroupZero + textGroupWidth / 2)
    .attr("y", (t) => y(t.id) + BAR_HEIGHT / 2);

  textGroup
    .append("text")
    .text((t) => formatCurrency(t.amount))
    .attr("text-anchor", "end")
    .attr("dominant-baseline", "middle")
    .style("fill", "#000")
    .attr("x", textGroupZero + textGroupWidth)
    .attr("y", (t) => y(t.id) + BAR_HEIGHT / 2);

  d3.select("#d3-portfolio-treemap")
    .append("div")
    .style("height", height + margin.top + margin.bottom + "px")
    .style("position", "absolute")
    .style("width", "100%")
    .selectAll("div")
    .data(portfolioAggregates)
    .enter()
    .append("div")
    .style("position", "absolute")
    .style("left", margin.left + x(0) + "px")
    .style("top", (t) => margin.top + y(t.id) + paddingTop + "px")
    .style("height", y.bandwidth() + "px")
    .style("width", x.range()[1] - x.range()[0] + "px")
    .append("div")
    .style("position", "relative")
    .style("height", y.bandwidth() + "px")
    .each(function (pa) {
      renderPartition(this, pa.breakdowns, pa, d3.treemap(), color);
    });
}

function renderPartition(
  element: HTMLElement,
  breakdowns: CommodityBreakdown[],
  pa: PortfolioAggregate,
  hierarchy: any,
  color: d3.ScaleOrdinal<string, string>
) {
  if (_.isEmpty(breakdowns)) {
    return;
  }

  const rootBreakdown: CommodityBreakdown = {
    security_id: "",
    security_name: "",
    security_type: "",
    percentage: 0,
    commodity_name: "root",
    amount: pa.amount
  };

  breakdowns.unshift(rootBreakdown);

  const byName: Record<string, CommodityBreakdown> = _.chain(breakdowns)
    .map((b) => [b.commodity_name, b])
    .fromPairs()
    .value();

  const div = d3.select(element),
    margin = { top: 0, right: 0, bottom: 0, left: 20 },
    width = element.parentElement.clientWidth - margin.left - margin.right,
    height = +div.style("height").replace("px", "") - margin.top - margin.bottom;

  const percent = (d: d3.HierarchyNode<CommodityBreakdown>) => {
    return formatFloat((d.value / root.value) * 100) + "%";
  };

  const stratify = d3
    .stratify<CommodityBreakdown>()
    .id((d) => d.commodity_name)
    .parentId((d) => (d.commodity_name == "root" ? null : "root"));

  const partition = hierarchy.size([width, height]).round(true);

  const root = stratify(breakdowns)
    .sum((a) => a.percentage)
    .sort(function (a, b) {
      return b.height - a.height || b.value - a.value;
    });

  partition(root);

  const nodes = div.selectAll(".node").data(root.descendants());

  nodes.exit().remove();

  const cell = nodes
    .enter()
    .append("div")
    .attr("class", "node")
    .attr("data-tippy-content", (d) => {
      const breakdown = byName[d.id];
      return tooltip([
        ["Commodity", [breakdown.commodity_name, "has-text-right"]],
        ["Security ID", [breakdown.security_id, "has-text-right"]],
        ["Security Type", [breakdown.security_type, "has-text-right"]],
        ["Amount", [formatCurrency(breakdown.amount), "has-text-weight-bold has-text-right"]],
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
    .style("font-size", ".7rem")
    .attr("class", "has-text-weight-bold")
    .text((d) => `${d.id} ${formatFloat(d.value)}%`);
}

function formatCommodityName(name: string) {
  return name.replaceAll(
    /([*]|EQ - |\bINC\b|\bInc\b|\bLTD\b|\bLtd\b|\bLt\b|\bLimited\b|\bLIMITED\b|[., ]+$)/g,
    ""
  );
}
