import type { Arima } from "arima/async";
import * as d3 from "d3";
import { Delaunay } from "d3";
import _, { first, isEmpty, last, takeRight } from "lodash";
import tippy from "tippy.js";
import COLORS from "./colors";
import type { Forecast, Point, Posting } from "./utils";
import {
  formatCurrency,
  formatCurrencyCrude,
  formatFloat,
  groupSumBy,
  isMobile,
  now,
  rem,
  skipTicks,
  sumPostings,
  tooltip
} from "./utils";
import dayjs from "dayjs";
import * as financial from "financial";
import { iconify } from "./icon";

const WHEN = financial.PaymentDueTime.Begin;

export function solvePMTOrNper(
  fv: number,
  rate: number,
  pv: number,
  pmt: number,
  targetDate: string
) {
  const empty = { pmt: 0, targetDate: "" };

  rate = rate / (100 * 12);
  const today = now().startOf("month");

  let targetDateObject = dayjs(targetDate, "YYYY-MM-DD", true);
  if (targetDateObject.isValid()) {
    const nper = targetDateObject.diff(today, "months");
    if (nper <= 0) {
      return empty;
    }
    pmt = financial.pmt(rate, nper, pv, -fv, WHEN);
    if (pmt <= 0) {
      // target will be achieved without any monthly savings
      return { pmt: 0.001, targetDate };
    }
  } else if (pmt > 0) {
    const nper = financial.nper(rate, pmt, pv, -fv, WHEN);
    targetDateObject = today.add(Math.ceil(nper), "months");
    targetDate = targetDateObject.format("YYYY-MM-DD");
  }

  return { pmt, targetDate };
}

export function project(
  fv: number,
  rate: number,
  targetDate: dayjs.Dayjs,
  pmt: number,
  pv: number
): Forecast[] {
  rate = rate / (100 * 12);
  const today = now().startOf("month");

  if (fv <= pv) {
    return [];
  }

  if (targetDate.isSameOrBefore(today, "day")) {
    return [];
  }

  const nper = targetDate.diff(today, "months");
  if (nper <= 0) {
    return [];
  }

  const points: Forecast[] = [];
  let current = today.add(1, "month");
  while (current.isSameOrBefore(targetDate, "day")) {
    const value = financial.fv(rate, current.diff(today, "months"), -pmt, -pv, WHEN);
    points.push({ date: current, value, error: 0 });
    current = current.add(1, "month");
  }

  return points;
}

export function forecast(points: Point[], target: number, ARIMA: typeof Arima): Forecast[] {
  const configs = [
    { p: 3, d: 0, q: 1, s: 0, verbose: false },
    { p: 2, d: 0, q: 1, s: 0, verbose: false }
  ];

  for (const config of configs) {
    const forecast = doForecast(config, points, target, ARIMA);
    if (!_.isEmpty(forecast)) {
      return forecast;
    }
  }
  return [];
}

function doForecast(
  config: object,
  points: Point[],
  target: number,
  ARIMA: typeof Arima
): Forecast[] {
  const values = points.map((p) => p.value);
  const arima = new ARIMA(config).train(values);

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
        if (
          point.value > 1e20 ||
          point.value < -1e20 ||
          point.error > 1e20 ||
          point.value < -1e20
        ) {
          return [];
        }
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
    if (p.value >= target * (i / 4)) {
      result.push(p);
      i++;
    }
  }

  return result;
}

export function renderProgress(
  points: Point[],
  predictions: Forecast[],
  breakPoints: Point[],
  element: Element,
  { targetSavings }: { targetSavings: number }
) {
  const start = first(points).date,
    end = (last(predictions) || last(points)).date;
  const positions = _.map(points.concat(predictions), (p) => p.value);

  const svg = d3.select(element),
    margin = { top: rem(40), right: rem(80), bottom: rem(20), left: rem(40) },
    width = Math.max(element.parentElement.clientWidth, 1000) - margin.left - margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  svg.attr("width", width + margin.left + margin.right);

  const lineKeys = ["actual", "forecast"];
  const lineScale = d3
    .scaleOrdinal<string>()
    .domain(lineKeys)
    .range([COLORS.secondary, COLORS.primary]);

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
        .curve(d3.curveMonotoneX)
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
        .curve(d3.curveMonotoneX)
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
        .curve(d3.curveMonotoneX)
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
    .attr("cx", (p) => x(p.date))
    .attr("cy", (p) => y(p.value));

  const voronoiPoints: Delaunay.Point[] = _.map(points.concat(predictions), (p) => [
    x(p.date),
    y(p.value)
  ]);
  const voronoi = Delaunay.from(voronoiPoints).voronoi([0, 0, width, height]);
  const hoverCircle = g.append("circle").attr("r", "3").attr("fill", "none");
  const t = tippy(hoverCircle.node(), { theme: "light", delay: 0, allowHTML: true });

  g.append("g")
    .selectAll("path")
    .data(points.concat(predictions))
    .enter()
    .append("path")
    .style("pointer-events", "all")
    .style("fill", "none")
    .attr("d", (_, i) => {
      return voronoi.renderCell(i);
    })
    .on("mouseover", (_, d) => {
      hoverCircle.attr("cx", x(d.date)).attr("cy", y(d.value)).attr("fill", COLORS.tertiary);

      t.setProps({
        placement: "top",
        content: tooltip([
          ["Date", d.date.format("DD MMM YYYY")],
          ["Savings", [formatCurrency(d.value), "has-text-weight-bold has-text-right"]],
          [
            "",
            [
              formatFloat((d.value / targetSavings) * 100) + "%",
              "has-text-weight-bold has-text-right"
            ]
          ]
        ])
      });
      t.show();
    })
    .on("mouseout", () => {
      t.hide();
      hoverCircle.attr("fill", "none");
    });

  return () => {
    t.destroy();
  };
}

export function renderInvestmentTimeline(postings: Posting[], element: Element, pmt: number) {
  const timeFormat = "MMM YYYY";
  const MAX_BAR_WIDTH = 40;
  const svg = d3.select(element),
    margin = { top: 10, right: 50, bottom: 50, left: 40 },
    width = element.parentElement.clientWidth - margin.left - margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const groupKeys = _.chain(postings)
    .map((p) => p.account)
    .uniq()
    .sort()
    .value();

  const defaultValues = _.zipObject(
    groupKeys,
    _.map(groupKeys, () => 0)
  );

  interface Point {
    date: dayjs.Dayjs;
    total: number;
    month: string;
    postings: Posting[];
    [key: string]: number | string | dayjs.Dayjs | Posting[];
  }
  const points: Point[] = [];
  const groupedPostings = _.groupBy(postings, (p) => p.date.format(timeFormat));
  const months = isMobile() ? 12 : 24;
  let start = now().startOf("month").subtract(months, "months");
  while (start.isBefore(now())) {
    const month = start.format(timeFormat);
    if (!groupedPostings[month]) {
      groupedPostings[month] = [];
    }

    const ps = groupedPostings[month];

    const values = _.chain(ps)
      .groupBy((p) => p.account)
      .flatMap((postings, key) => [[key, _.sumBy(postings, (p) => p.amount)]])
      .fromPairs()
      .value();

    const total = sumPostings(ps);

    const point = _.merge(
      {
        month,
        total,
        date: dayjs(month, timeFormat),
        postings: ps
      },
      defaultValues,
      values
    );

    points.push(point);
    start = start.add(1, "month");
  }

  const x = d3.scaleBand().range([0, width]).paddingInner(0.1).paddingOuter(0);
  const y = d3.scaleLinear().range([height, 0]);

  const sum = (p: Point) => p.total;
  const max = d3.max(points, sum);
  const min = d3.min([0, d3.min(points, sum)]);
  x.domain(points.map((p) => p.month));
  y.domain([min, max]);

  g.append("g")
    .attr("class", "axis x")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x)
        .ticks(5)
        .tickFormat(skipTicks(30, x, (d) => d.toString()))
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

  if (pmt > 0) {
    g.append("line")
      .attr("fill", "none")
      .attr("stroke", COLORS.primary)
      .attr("x1", 0)
      .attr("x2", width)
      .attr("y1", y(pmt))
      .attr("y2", y(pmt))
      .attr("stroke-width", "2px")
      .attr("stroke-linecap", "round")
      .attr("stroke-dasharray", "4 6");

    g.append("text")
      .style("font-size", "0.714rem")
      .attr("dx", "3px")
      .attr("dy", "0.3em")
      .attr("x", width)
      .attr("y", y(pmt))
      .attr("fill", COLORS.primary)
      .text(formatCurrencyCrude(pmt));
  }

  g.append("g")
    .selectAll("rect")
    .data(points)
    .enter()
    .append("rect")
    .attr("stroke", (p) => (p.total <= 0 ? COLORS.tertiary : COLORS.secondary))
    .attr("fill", (p) => (p.total <= 0 ? COLORS.tertiary : COLORS.secondary))
    .attr("fill-opacity", 0.5)
    .attr("data-tippy-content", (p) => {
      const group = groupSumBy(p.postings, (p) => p.account);
      return tooltip(
        _.sortBy(
          _.map(group, (amount, account) => [
            iconify(account),
            [formatCurrency(amount), "has-text-weight-bold has-text-right"]
          ]),
          (r) => r[0]
        ),
        { total: formatCurrency(p.total) }
      );
    })
    .attr("x", function (p) {
      return x(p.month) + (x.bandwidth() - Math.min(x.bandwidth(), MAX_BAR_WIDTH)) / 2;
    })
    .attr("y", function (p) {
      return p.total <= 0 ? y(0) : y(p.total);
    })
    .attr("height", function (p) {
      return p.total <= 0 ? y(p.total) - y(0) : y(0) - y(p.total);
    })
    .attr("width", Math.min(x.bandwidth(), MAX_BAR_WIDTH));
}
