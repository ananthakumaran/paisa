import * as d3 from "d3";
import legend from "d3-svg-legend";
import type dayjs from "dayjs";
import _ from "lodash";
import {
  formatCurrency,
  formatCurrencyCrude,
  type Income,
  type Posting,
  restName,
  secondName,
  skipTicks,
  tooltip,
  type IncomeYearlyCard
} from "./utils";
import { generateColorScheme } from "./colors";

const LEGEND_PADDING = 70;

export function renderMonthlyInvestmentTimeline(incomes: Income[]) {
  renderIncomeTimeline(incomes, "#d3-income-timeline", "MMM-YYYY");
}

function renderIncomeTimeline(incomes: Income[], id: string, timeFormat: string) {
  const MAX_BAR_WIDTH = 40;
  const svg = d3.select(id),
    margin = { top: 60, right: 30, bottom: 80, left: 40 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const postings = _.flatMap(incomes, (i) => i.postings);
  const groupKeys = _.chain(postings)
    .map((p) => incomeGroup(p))
    .uniq()
    .sort()
    .value();

  const groupTotal = _.chain(postings)
    .groupBy((p) => incomeGroup(p))
    .map((postings, key) => {
      const total = _.sumBy(postings, (p) => -p.amount);
      return [key, `${key} ${formatCurrency(total)}`];
    })
    .fromPairs()
    .value();

  const defaultValues = _.zipObject(
    groupKeys,
    _.map(groupKeys, () => 0)
  );

  interface Point {
    date: dayjs.Dayjs;
    month: string;
    [key: string]: number | string | dayjs.Dayjs;
  }
  let points: Point[] = [];

  points = _.map(incomes, (i) => {
    const values = _.chain(i.postings)
      .groupBy((p) => incomeGroup(p))
      .flatMap((postings, key) => [[key, _.sumBy(postings, (p) => -p.amount)]])
      .fromPairs()
      .value();

    return _.merge(
      {
        month: i.date.format(timeFormat),
        date: i.date,
        postings: i.postings
      },
      defaultValues,
      values
    );
  });

  const x = d3.scaleBand().range([0, width]).paddingInner(0.1).paddingOuter(0);
  const y = d3.scaleLinear().range([height, 0]);

  const sum = (filter: (n: number) => boolean) => (p: Point) =>
    _.sum(
      _.filter(
        _.map(groupKeys, (k) => p[k]),
        filter
      )
    );
  x.domain(points.map((p) => p.month));
  y.domain([
    d3.min(
      points,
      sum((a) => a < 0)
    ),
    d3.max(
      points,
      sum((a) => a > 0)
    )
  ]);

  const z = generateColorScheme(groupKeys);

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

  g.append("g")
    .selectAll("g")
    .data(
      d3.stack().offset(d3.stackOffsetDiverging).keys(groupKeys)(
        points as { [key: string]: number }[]
      )
    )
    .enter()
    .append("g")
    .attr("fill", function (d) {
      return z(d.key.split("-")[0]);
    })
    .selectAll("rect")
    .data(function (d) {
      return d;
    })
    .enter()
    .append("rect")
    .attr("data-tippy-content", (d) => {
      const postings: Posting[] = (d.data as any).postings;
      const total = _.sumBy(postings, (p) => -p.amount);
      return tooltip(
        _.sortBy(
          postings.map((p) => [
            restName(p.account),
            [formatCurrency(-p.amount), "has-text-weight-bold has-text-right"]
          ]),
          (r) => r[0]
        ),
        { total: formatCurrency(total) }
      );
    })
    .attr("x", function (d) {
      return (
        x((d.data as any).month) + (x.bandwidth() - Math.min(x.bandwidth(), MAX_BAR_WIDTH)) / 2
      );
    })
    .attr("y", function (d) {
      return y(d[1]);
    })
    .attr("height", function (d) {
      return y(d[0]) - y(d[1]);
    })
    .attr("width", Math.min(x.bandwidth(), MAX_BAR_WIDTH));

  svg
    .append("g")
    .attr("class", "legendOrdinal")
    .attr("transform", `translate(${LEGEND_PADDING / 2},0)`);

  const legendOrdinal = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(LEGEND_PADDING)
    .labels(({ i }: { i: number }) => {
      return groupTotal[groupKeys[i]];
    })
    .scale(z);

  (legendOrdinal as any).labelWrap(75); // type missing

  svg.select(".legendOrdinal").call(legendOrdinal as any);
}

function financialYear(card: IncomeYearlyCard) {
  return `${card.start_date.format("YYYY")} - ${card.end_date.format("YY")}`;
}

export function renderYearlyIncomeTimeline(yearlyCards: IncomeYearlyCard[]) {
  const id = "#d3-yearly-income-timeline";
  const BAR_HEIGHT = 20;
  const svg = d3.select(id),
    margin = { top: 40, right: 20, bottom: 20, left: 70 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const groups = _.chain(yearlyCards)
    .flatMap((c) => c.postings)
    .map((p) => secondName(p.account))
    .uniq()
    .sort()
    .value();

  const defaultValues = _.zipObject(
    groups,
    _.map(groups, () => 0)
  );

  const start = _.min(_.map(yearlyCards, (c) => c.start_date)),
    end = _.max(_.map(yearlyCards, (c) => c.end_date));

  const height = BAR_HEIGHT * (end.year() - start.year());
  svg.attr("height", height + margin.top + margin.bottom);

  interface Point {
    year: string;
    [key: string]: number | string | dayjs.Dayjs;
  }
  const points: Point[] = [];

  _.each(yearlyCards, (card) => {
    const postings = card.postings;
    const values = _.chain(postings)
      .groupBy((t) => secondName(t.account))
      .map((postings, key) => [key, _.sum(_.map(postings, (p) => -p.amount))])
      .fromPairs()
      .value();

    points.push(
      _.merge(
        {
          year: financialYear(card),
          postings: postings
        },
        defaultValues,
        values
      )
    );
  });

  const x = d3.scaleLinear().range([0, width]);
  const y = d3.scaleBand().range([height, 0]).paddingInner(0.1).paddingOuter(0);

  y.domain(points.map((p) => p.year));
  x.domain([0, d3.max(points, (p: Point) => _.sum(_.map(groups, (k) => p[k])))]);

  const z = generateColorScheme(groups);

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(d3.axisBottom(x).tickSize(-height).tickFormat(formatCurrencyCrude));

  g.append("g").attr("class", "axis y dark").call(d3.axisLeft(y));

  g.append("g")
    .selectAll("g")
    .data(
      d3.stack().offset(d3.stackOffsetDiverging).keys(groups)(points as { [key: string]: number }[])
    )
    .enter()
    .append("g")
    .attr("fill", function (d) {
      return z(d.key.split("-")[0]);
    })
    .selectAll("rect")
    .data(function (d) {
      return d;
    })
    .enter()
    .append("rect")
    .attr("data-tippy-content", (d) => {
      let grandTotal = 0;
      return tooltip(
        _.sortBy(
          groups.flatMap((k) => {
            const total = d.data[k];
            if (total == 0) {
              return [];
            }
            grandTotal += total;
            return [[k, [formatCurrency(total), "has-text-weight-bold has-text-right"]]];
          }),
          (r) => r[0]
        ),
        { total: formatCurrency(grandTotal) }
      );
    })
    .attr("x", function (d) {
      return x(d[0]);
    })
    .attr("y", function (d) {
      return y((d.data as any).year) + (y.bandwidth() - Math.min(y.bandwidth(), BAR_HEIGHT)) / 2;
    })
    .attr("width", function (d) {
      return x(d[1]) - x(d[0]);
    })
    .attr("height", y.bandwidth());

  svg.append("g").attr("class", "legendOrdinal").attr("transform", "translate(40,0)");

  const legendOrdinal = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(LEGEND_PADDING)
    .labels(groups)
    .scale(z);

  svg.select(".legendOrdinal").call(legendOrdinal as any);
}

export function renderYearlyTimelineOf(
  label: string,
  key: "net_tax" | "net_income",
  color: string,
  yearlyCards: IncomeYearlyCard[]
) {
  const id = `#d3-yearly-${key}-timeline`;
  const BAR_HEIGHT = 20;
  const svg = d3.select(id),
    margin = { top: 40, right: 20, bottom: 20, left: 70 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const colorKeys = [label];
  const colorScale = d3.scaleOrdinal<string>().domain(colorKeys).range([color]);

  const start = _.min(_.map(yearlyCards, (c) => c.start_date)),
    end = _.max(_.map(yearlyCards, (c) => c.end_date));

  const height = BAR_HEIGHT * (end.year() - start.year());
  svg.attr("height", height + margin.top + margin.bottom);

  interface Point {
    year: string;
    value: number;
  }

  const points: Point[] = _.map(yearlyCards, (card) => {
    return {
      year: financialYear(card),
      value: card[key]
    };
  });

  const x = d3.scaleLinear().range([0, width]);
  const y = d3.scaleBand().range([height, 0]).paddingInner(0.1).paddingOuter(0);

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
    .attr("fill", color)
    .attr("data-tippy-content", (d) => {
      return tooltip([[label, [formatCurrency(d.value), "has-text-weight-bold has-text-right"]]]);
    })
    .attr("x", x(0))
    .attr("y", function (d) {
      return y(d.year) + (y.bandwidth() - Math.min(y.bandwidth(), BAR_HEIGHT)) / 2;
    })
    .attr("width", function (d) {
      return x(d.value) - x(0);
    })
    .attr("height", y.bandwidth());

  svg.append("g").attr("class", "legendOrdinal").attr("transform", "translate(40,0)");

  const legendOrdinal = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(LEGEND_PADDING)
    .labels(colorKeys)
    .scale(colorScale);

  svg.select(".legendOrdinal").call(legendOrdinal as any);
}

export function incomeGroup(posting: Posting) {
  return secondName(posting.account);
}
