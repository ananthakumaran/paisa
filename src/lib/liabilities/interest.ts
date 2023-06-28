import chroma from "chroma-js";
import * as d3 from "d3";
import legend from "d3-svg-legend";
import type dayjs from "dayjs";
import _ from "lodash";
import COLORS from "$lib/colors";
import {
  formatCurrency,
  formatCurrencyCrude,
  formatFloat,
  type Interest,
  type InterestOverview,
  tooltip,
  skipTicks,
  restName
} from "$lib/utils";

const areaKeys = ["gain", "loss"];
const colors = [COLORS.gain, COLORS.loss];
const areaScale = d3.scaleOrdinal<string>().domain(areaKeys).range(colors);
const lineKeys = ["balance", "drawn", "repaid"];
const lineScale = d3
  .scaleOrdinal<string>()
  .domain(lineKeys)
  .range([COLORS.primary, COLORS.secondary, COLORS.tertiary]);

function renderTable(interest: Interest) {
  const tbody = d3.select(this);
  const current = _.last(interest.overview_timeline);
  tbody.html(function () {
    return `
<tr>
  <td>Account</td>
  <td class='has-text-right has-text-weight-bold'>${restName(interest.account)}</td>
</tr>
<tr>
  <td>Loan Drawn</td>
  <td class='has-text-right'>${formatCurrency(current.drawn_amount)}</td>
</tr>
<tr>
  <td>Loan Repaid</td>
  <td class='has-text-right'>${formatCurrency(current.repaid_amount)}</td>
</tr>
<tr>
  <td>Interest</td>
  <td class='has-text-right'>${formatCurrency(current.interest_amount)}</td>
</tr>
<tr>
  <td>Balance</td>
  <td class='has-text-right'>${formatCurrency(
    current.drawn_amount + current.interest_amount - current.repaid_amount
  )}</td>
</tr>
<tr>
  <td>APR</td>
  <td class='has-text-right'>${formatFloat(interest.apr)}</td>
</tr>
`;
  });
}

export function renderOverview(gains: Interest[]) {
  gains = _.sortBy(gains, (g) => g.account);
  const BAR_HEIGHT = 15;
  const id = "#d3-interest-overview";
  const svg = d3.select(id),
    margin = { top: 5, right: 20, bottom: 30, left: 150 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    height = gains.length * BAR_HEIGHT * 2,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");
  svg.attr("height", height + margin.top + margin.bottom);

  const y = d3.scaleBand().range([0, height]).paddingInner(0).paddingOuter(0);
  y.domain(gains.map((g) => restName(g.account)));
  const y1 = d3
    .scaleBand()
    .range([0, y.bandwidth()])
    .domain(["0", "1"])
    .paddingInner(0)
    .paddingOuter(0.1);

  const keys = ["balance", "drawn", "repaid", "gain", "loss"];
  const colors = [COLORS.primary, COLORS.secondary, COLORS.tertiary, COLORS.gain, COLORS.loss];
  const z = d3.scaleOrdinal<string>(colors).domain(keys);

  const getDrawnAmount = (g: Interest) => _.last(g.overview_timeline).drawn_amount;

  const getInterestAmount = (g: Interest) => _.last(g.overview_timeline).interest_amount;
  const getRepaidAmount = (g: Interest) => _.last(g.overview_timeline).repaid_amount;

  const getBalanceAmount = (g: Interest) => {
    const current = _.last(g.overview_timeline);
    return current.drawn_amount + current.interest_amount - current.repaid_amount;
  };

  const maxX = _.chain(gains)
    .map((g) => getDrawnAmount(g) + _.max([getInterestAmount(g), 0]))
    .max()
    .value();
  const aprWidth = 250;
  const aprTextWidth = 40;
  const aprMargin = 20;
  const textGroupWidth = 225;
  const textGroupZero = aprWidth + aprTextWidth + aprMargin;

  const x = d3.scaleLinear().range([textGroupZero + textGroupWidth, width]);
  x.domain([0, maxX]);
  const x1 = d3
    .scaleLinear()
    .range([0, aprWidth])
    .domain([
      _.min([_.min(_.map(gains, (g) => g.apr)), 0]),
      _.max([0, _.max(_.map(gains, (g) => g.apr))])
    ]);

  g.append("line")
    .classed("svg-grey-lighter", true)
    .attr("x1", aprWidth + aprTextWidth + aprMargin / 2)
    .attr("y1", 0)
    .attr("x2", aprWidth + aprTextWidth + aprMargin / 2)
    .attr("y2", height);

  g.append("line")
    .classed("svg-grey-lighter", true)
    .attr("x1", 0)
    .attr("y1", height)
    .attr("x2", width)
    .attr("y2", height);

  g.append("text")
    .classed("svg-text-grey", true)
    .text("APR")
    .attr("text-anchor", "middle")
    .attr("x", aprWidth / 2)
    .attr("y", height + 30);

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(d3.axisBottom(x).tickSize(-height).tickFormat(skipTicks(60, x, formatCurrencyCrude)));

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x1)
        .tickSize(-height)
        .tickFormat(skipTicks(40, x1, (n: number) => formatFloat(n, 1)))
    );

  g.append("g").attr("class", "axis y dark").call(d3.axisLeft(y));

  const textGroup = g
    .append("g")
    .selectAll("g")
    .data(gains)
    .enter()
    .append("g")
    .attr("class", "inline-text");

  textGroup
    .append("text")
    .text((g) => formatCurrency(getDrawnAmount(g)))
    .attr("dominant-baseline", "hanging")
    .attr("text-anchor", "end")
    .style("fill", (g) => (getDrawnAmount(g) > 0 ? z("drawn") : "none"))
    .attr("dx", "-3")
    .attr("dy", "3")
    .attr("x", textGroupZero + textGroupWidth / 3)
    .attr("y", (g) => y(restName(g.account)));

  textGroup
    .append("text")
    .text((g) => formatCurrency(getInterestAmount(g)))
    .attr("dominant-baseline", "hanging")
    .attr("text-anchor", "end")
    .style("fill", (g) => (getInterestAmount(g) > 0 ? chroma(z("loss")).darken().hex() : "none"))
    .attr("dx", "-3")
    .attr("dy", "3")
    .attr("x", textGroupZero + (textGroupWidth * 2) / 3)
    .attr("y", (g) => y(restName(g.account)));

  textGroup
    .append("text")
    .text((g) => formatCurrency(getBalanceAmount(g)))
    .attr("text-anchor", "end")
    .style("fill", (g) => (getBalanceAmount(g) > 0 ? z("balance") : "none"))
    .attr("dx", "-3")
    .attr("dy", "-3")
    .attr("x", textGroupZero + textGroupWidth / 3)
    .attr("y", (g) => y(restName(g.account)) + y.bandwidth());

  textGroup
    .append("text")
    .text((g) => formatCurrency(getInterestAmount(g)))
    .attr("text-anchor", "end")
    .style("fill", (g) => (getInterestAmount(g) < 0 ? chroma(z("gain")).darken().hex() : "none"))
    .attr("dx", "-3")
    .attr("dy", "-3")
    .attr("x", textGroupZero + (textGroupWidth * 2) / 3)
    .attr("y", (g) => y(restName(g.account)) + y.bandwidth());

  textGroup
    .append("text")
    .text((g) => formatCurrency(getRepaidAmount(g)))
    .attr("text-anchor", "end")
    .style("fill", (g) => (getRepaidAmount(g) > 0 ? z("repaid") : "none"))
    .attr("dx", "-3")
    .attr("dy", "-3")
    .attr("x", textGroupZero + textGroupWidth)
    .attr("y", (g) => y(restName(g.account)) + y.bandwidth());

  textGroup
    .append("line")
    .classed("svg-grey-lighter", true)
    .attr("x1", 0)
    .attr("y1", (g) => y(restName(g.account)))
    .attr("x2", width)
    .attr("y2", (g) => y(restName(g.account)));

  textGroup
    .append("text")
    .text((g) => formatFloat(g.apr))
    .attr("text-anchor", "end")
    .attr("dominant-baseline", "middle")
    .style("fill", (g) =>
      g.apr < 0 ? chroma(z("loss")).darken().hex() : chroma(z("gain")).darken().hex()
    )
    .attr("x", aprWidth + aprTextWidth)
    .attr("y", (g) => y(restName(g.account)) + y.bandwidth() / 2);

  const groups = g
    .append("g")
    .selectAll("g.group")
    .data(gains)
    .enter()
    .append("g")
    .attr("class", "group")
    .attr("transform", (g) => "translate(0," + y(restName(g.account)) + ")");

  groups
    .selectAll("g")
    .data((g) => [
      d3.stack().keys(["drawn", "loss"])([
        {
          i: "0",
          data: g,
          drawn: getDrawnAmount(g),
          loss: _.max([getInterestAmount(g), 0])
        }
      ] as any),
      d3.stack().keys(["balance", "gain", "repaid"])([
        {
          i: "1",
          data: g,
          balance: getBalanceAmount(g),
          repaid: getRepaidAmount(g),
          gain: Math.abs(_.min([getInterestAmount(g), 0]))
        }
      ] as any)
    ])
    .enter()
    .append("g")
    .selectAll("rect")
    .data((d) => {
      return d;
    })
    .enter()
    .append("rect")
    .attr("fill", (d) => {
      return z(d.key);
    })
    .attr("x", (d) => x(d[0][0]))
    .attr("y", (d: any) => y1(d[0].data.i))
    .attr("height", y1.bandwidth())
    .attr("width", (d) => x(d[0][1]) - x(d[0][0]));

  const paddingTop = (y1.range()[1] - y1.bandwidth() * 2) / 2;
  g.append("g")
    .selectAll("rect")
    .data(gains)
    .enter()
    .append("rect")
    .attr("fill", (g) => (g.apr < 0 ? z("loss") : z("gain")))
    .attr("x", (g) => (g.apr < 0 ? x1(g.apr) : x1(0)))
    .attr("y", (g) => y(restName(g.account)) + paddingTop)
    .attr("height", y.bandwidth() - paddingTop * 2)
    .attr("width", (g) => Math.abs(x1(0) - x1(g.apr)));

  g.append("g")
    .selectAll("rect")
    .data(gains)
    .enter()
    .append("rect")
    .attr("fill", "transparent")
    .attr("data-tippy-content", (g: Interest) => {
      const current = _.last(g.overview_timeline);
      return tooltip([
        ["Account", [g.account, "has-text-weight-bold has-text-right"]],
        [
          "Loan Drawn",
          [formatCurrency(current.drawn_amount), "has-text-weight-bold has-text-right"]
        ],
        [
          "Loan Repaid",
          [formatCurrency(current.repaid_amount), "has-text-weight-bold has-text-right"]
        ],
        [
          "Interest",
          [formatCurrency(current.interest_amount), "has-text-weight-bold has-text-right"]
        ],
        [
          "Balance",
          [
            formatCurrency(current.drawn_amount + current.interest_amount - current.repaid_amount),
            "has-text-weight-bold has-text-right"
          ]
        ],
        ["APR", [formatFloat(g.apr), "has-text-weight-bold has-text-right"]]
      ]);
    })
    .attr("x", 0)
    .attr("y", (g) => y(restName(g.account)))
    .attr("height", y.bandwidth())
    .attr("width", width);
}

export function renderPerAccountOverview(interests: Interest[]) {
  const dates = _.flatMap(interests, (g) => _.map(g.overview_timeline, (o) => o.date));
  const start = _.min(dates),
    end = _.max(dates);

  const divs = d3
    .select("#d3-interest-timeline-breakdown")
    .selectAll("div")
    .data(_.sortBy(interests, (g) => g.account));

  divs.exit().remove();

  const columns = divs.enter().append("div").attr("class", "columns");

  const leftColumn = columns.append("div").attr("class", "column is-4 is-3-desktop is-2-fullhd");
  leftColumn
    .append("table")
    .attr("class", "table is-narrow is-fullwidth is-size-7")
    .append("tbody")
    .each(renderTable);

  const rightColumn = columns.append("div").attr("class", "column");
  rightColumn
    .append("svg")
    .attr("width", "100%")
    .attr("height", "150")
    .each(function (gain) {
      renderOverviewSmall(gain.overview_timeline, this, [start, end]);
    });
}

function renderOverviewSmall(
  points: InterestOverview[],
  element: Element,
  xDomain: [dayjs.Dayjs, dayjs.Dayjs]
) {
  const svg = d3.select(element),
    margin = { top: 5, right: 80, bottom: 20, left: 40 },
    width = element.parentElement.clientWidth - margin.left - margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const x = d3.scaleTime().range([0, width]).domain(xDomain),
    y = d3
      .scaleLinear()
      .range([height, 0])
      .domain([
        0,
        d3.max<InterestOverview, number>(points, (d) => d.interest_amount + d.drawn_amount)
      ]),
    z = d3.scaleOrdinal<string>(colors).domain(areaKeys);

  const area = (y0: number, y1: (d: InterestOverview) => number) =>
    d3
      .area<InterestOverview>()
      .curve(d3.curveBasis)
      .x((d) => x(d.date))
      .y0(y0)
      .y1(y1);

  g.append("g")
    .attr("class", "axis x")
    .attr("transform", "translate(0," + height + ")")
    .call(d3.axisBottom(x));

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", `translate(${width},0)`)
    .call(d3.axisRight(y).ticks(5).tickPadding(5).tickFormat(formatCurrencyCrude));

  g.append("g")
    .attr("class", "axis y")
    .call(d3.axisLeft(y).ticks(5).tickSize(-width).tickFormat(formatCurrencyCrude));

  const layer = g.selectAll(".layer").data([points]).enter().append("g").attr("class", "layer");

  const clipAboveID = _.uniqueId("clip-above");
  layer
    .append("clipPath")
    .attr("id", clipAboveID)
    .append("path")
    .attr(
      "d",
      area(height, (d) => {
        return y(d.repaid_amount - d.interest_amount);
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
        return y(d.repaid_amount - d.interest_amount);
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
        return y(d.repaid_amount);
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
        return y(d.repaid_amount);
      })
    );

  layer
    .append("path")
    .style("stroke", lineScale("drawn"))
    .style("fill", "none")
    .attr(
      "d",
      d3
        .line<InterestOverview>()
        .curve(d3.curveBasis)
        .x((d) => x(d.date))
        .y((d) => y(d.drawn_amount))
    );

  layer
    .append("path")
    .style("stroke", lineScale("repaid"))
    .style("fill", "none")
    .attr(
      "d",
      d3
        .line<InterestOverview>()
        .curve(d3.curveBasis)
        .defined((d) => d.repaid_amount > 0)
        .x((d) => x(d.date))
        .y((d) => y(d.repaid_amount))
    );

  layer
    .append("path")
    .style("stroke", lineScale("balance"))
    .style("fill", "none")
    .attr(
      "d",
      d3
        .line<InterestOverview>()
        .curve(d3.curveBasis)
        .x((d) => x(d.date))
        .y((d) => y(d.drawn_amount + d.interest_amount - d.repaid_amount))
    );
}

export function renderLegend() {
  const svg = d3.select("#d3-interest-legend");
  svg.append("g").attr("class", "legendOrdinal").attr("transform", "translate(280,3)");

  const legendOrdinal = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(70)
    .labels(areaKeys)
    .scale(areaScale);

  svg.select(".legendOrdinal").call(legendOrdinal as any);

  svg.append("g").attr("class", "legendLine").attr("transform", "translate(30,3)");

  const legendLine = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(70)
    .labels(lineKeys)
    .scale(lineScale);

  svg.select(".legendLine").call(legendLine as any);
}
