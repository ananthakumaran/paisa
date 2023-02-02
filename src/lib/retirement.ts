import type { Arima } from "arima/async";
import * as d3 from "d3";
import _, { first, isEmpty, last } from "lodash";
import COLORS from "./colors";
import type { Forecast, Point } from "./utils";
import { formatCurrencyCrude } from "./utils";

export function renderProgress(points: Point[], predictions: Forecast[], element: Element) {
  const start = first(points).date,
    end = last(predictions).date;
  const positions = _.map(points.concat(predictions), (p) => p.value);

  const svg = d3.select(element),
    margin = { top: 40, right: 80, bottom: 20, left: 40 },
    width = element.parentElement.clientWidth - margin.left - margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const lineKeys = ["actual", "forecast"];
  const lineScale = d3
    .scaleOrdinal<string>()
    .domain(lineKeys)
    .range([COLORS.primary, COLORS.secondary]);

  const x = d3.scaleTime().range([0, width]).domain([start, end]),
    y = d3
      .scaleLinear()
      .range([height, 0])
      .domain([0, _.max(positions)]);

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

  g.append("path")
    .style("stroke", lineScale("actual"))
    .style("fill", "none")
    .attr(
      "d",
      d3
        .line<Point>()
        .curve(d3.curveBasis)
        .x((d) => x(d.date))
        .y((d) => y(d.value))(points)
    );

  g.append("path")
    .style("stroke", lineScale("forecast"))
    .style("fill", "none")
    .attr(
      "d",
      d3
        .line<Point>()
        .curve(d3.curveBasis)
        .x((d) => x(d.date))
        .y((d) => y(d.value))(predictions)
    );

  g.append("path")
    .style("fill", lineScale("forecast"))
    .style("opacity", "0.2")
    .attr(
      "d",
      d3
        .area<Forecast>()
        .curve(d3.curveBasis)
        .x((d) => x(d.date))
        .y0((d) => y(d.value - d.error / 2))
        .y1((d) => y(d.value + d.error / 2))(predictions)
    );
}

export function forecast(points: Point[], target: number, ARIMA: typeof Arima): Forecast[] {
  const values = points.map((p) => p.value);
  const arima = new ARIMA({
    p: 3,
    d: 0,
    q: 1,
    s: 0,
    // auto: true,
    verbose: false
  }).train(values);

  const predictYears = 3;
  let i = 1;
  while (i < 10) {
    const [predictions, errors] = arima.predict(predictYears * i * 365);
    if (last(predictions) > target) {
      const predictionsTimeline: Forecast[] = [];
      let start = last(points).date;
      while (!isEmpty(predictions)) {
        start = start.add(1, "day");
        const point = { date: start, value: predictions.shift(), error: Math.sqrt(errors.shift()) };
        predictionsTimeline.push(point);
      }
      return predictionsTimeline;
    }
    i++;
  }
  return [];
}
