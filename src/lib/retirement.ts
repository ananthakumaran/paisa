import type { Arima } from "arima/async";
import * as d3 from "d3";
import _, { first, isEmpty, last, takeRight } from "lodash";
import tippy, { type Placement } from "tippy.js";
import COLORS from "./colors";
import { formatCurrencyCrude, type Forecast, type Point } from "./utils";

export function renderProgress(
  points: Point[],
  predictions: Forecast[],
  breakPoints: Point[],
  element: Element
) {
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
        .y((d) => y(d.value))(takeRight(points, 1).concat(predictions))
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

  g.append("g")
    .selectAll("circle")
    .data(breakPoints)
    .join("circle")
    .attr("r", "3")
    .style("pointer-events", "none")
    .attr("fill", COLORS.tertiary)
    .attr("class", "axis x")
    .attr("data-tippy-placement", (_d, i) => ["top-end", "top", "bottom", "top-start"][i])
    .attr("data-tippy-content", (d, i) => {
      return `
<div class='has-text-centered'>${formatCurrencyCrude(d.value)} (${
        (i + 1) * 25
      }%)<br />${d.date.format("DD MMM YYYY")}</div>
`;
    })
    .attr("cx", (p) => x(p.date))
    .attr("cy", (p) => y(p.value));

  const instances = tippy("circle[data-tippy-content]", {
    onShow: (instance) => {
      const content = instance.reference.getAttribute("data-tippy-content");
      if (!_.isEmpty(content)) {
        instance.setContent(content);
        instance.setProps({
          placement: instance.reference.getAttribute("data-tippy-placement") as Placement
        });
      } else {
        return false;
      }
    },
    hideOnClick: false,
    allowHTML: true,
    appendTo: element.parentElement
  });

  instances.forEach((i) => i.show());

  return () => {
    instances.forEach((i) => i.destroy());
  };
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
    if (isEmpty(predictions)) {
      return [];
    }
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

export function findBreakPoints(points: Point[], target: number): Point[] {
  const result: Point[] = [];
  let i = 1;
  while (i <= 4 && !isEmpty(points)) {
    const p = points.shift();
    if (p.value > target * (i / 4)) {
      result.push(p);
      i++;
    }
  }

  return result;
}
