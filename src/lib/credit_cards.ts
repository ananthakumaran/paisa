import * as d3 from "d3";
import { formatCurrencyCrude, tooltip, formatCurrency } from "./utils";
import _ from "lodash";
import COLORS from "./colors";

export function renderYearlySpends(
  svgNode: SVGElement,
  yearlySpends: { [year: string]: { [month: string]: number } }
) {
  const BAR_HEIGHT = 20;
  const svg = d3.select(svgNode),
    margin = { top: 15, right: 20, bottom: 20, left: 70 },
    width = svgNode.parentElement.clientWidth - margin.left - margin.right,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const color = COLORS.expenses;

  const height = BAR_HEIGHT * Object.keys(yearlySpends).length;
  svg.attr("height", height + margin.top + margin.bottom);

  interface Point {
    year: string;
    value: number;
    breakdown: { [month: string]: number };
  }

  const points: Point[] = _.chain(yearlySpends)
    .map((breakdown, year) => {
      const value = _.sum(_.values(breakdown));
      return { year, breakdown, value };
    })
    .sortBy((p) => p.year)
    .value();

  const x = d3.scaleLinear().range([0, width]);
  const y = d3.scaleBand().range([height, 0]).paddingInner(0.2).paddingOuter(0);

  y.domain(points.map((p) => p.year));
  x.domain([0, d3.max(points, (p: Point) => p.value)]);

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(d3.axisBottom(x).tickSize(-height).tickFormat(formatCurrencyCrude));

  g.append("g").attr("class", "axis y dark").call(d3.axisLeft(y));

  g.append("g")
    .selectAll("rect")
    .data(points)
    .join("rect")
    .attr("stroke-opacity", 0.6)
    .attr("fill-opacity", 0.4)
    .attr("stroke", color)
    .attr("fill", color)
    .attr("data-tippy-content", (d) => {
      return tooltip(
        _.map(d.breakdown, (value, month) => {
          return [month, [formatCurrency(value), "has-text-right has-text-weight-bold"]];
        }),
        { total: formatCurrency(d.value), header: d.year }
      );
    })
    .attr("x", x(0))
    .attr("y", function (d) {
      return y(d.year) + (y.bandwidth() - Math.min(y.bandwidth(), BAR_HEIGHT)) / 2;
    })
    .attr("width", function (d) {
      return x(d.value) - x(0);
    })
    .attr("height", y.bandwidth());
}
