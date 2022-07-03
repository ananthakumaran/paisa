import * as d3 from "d3";
import legend from "d3-svg-legend";
import dayjs from "dayjs";
import _ from "lodash";
import {
  ajax,
  formatCurrency,
  formatCurrencyCrude,
  formatFloat,
  Gain,
  Overview,
  tooltip,
  skipTicks,
  restName
} from "./utils";

export default async function () {
  const { gain_timeline_breakdown: gains } = await ajax("/api/gain");
  _.each(gains, (g) =>
    _.each(g.overview_timeline, (o) => (o.timestamp = dayjs(o.date)))
  );

  renderLegend();
  renderOverview(gains);
  renderPerAccountOverview(gains);
}

const areaKeys = ["gain", "loss"];
const colors = ["#b2df8a", "#fb9a99"];
const areaScale = d3.scaleOrdinal<string>().domain(areaKeys).range(colors);
const lineKeys = ["balance", "investment", "withdrawal"];
const lineScale = d3
  .scaleOrdinal<string>()
  .domain(lineKeys)
  .range(["#1f77b4", "#17becf", "#ff7f0e"]);

function renderTable(gain: Gain) {
  const tbody = d3.select(this);
  const current = _.last(gain.overview_timeline);
  tbody.html(function () {
    return `
<tr>
  <td>Account</td>
  <td class='has-text-right has-text-weight-bold'>${gain.account}</td>
</tr>
<tr>
  <td>Investment</td>
  <td class='has-text-right'>${formatCurrency(current.investment_amount)}</td>
</tr>
<tr>
  <td>Withdrawal</td>
  <td class='has-text-right'>${formatCurrency(current.withdrawal_amount)}</td>
</tr>
<tr>
  <td>Gain</td>
  <td class='has-text-right'>${formatCurrency(current.gain_amount)}</td>
</tr>
<tr>
  <td>Balance</td>
  <td class='has-text-right'>${formatCurrency(
    current.investment_amount + current.gain_amount - current.withdrawal_amount
  )}</td>
</tr>
<tr>
  <td>XIRR</td>
  <td class='has-text-right'>${formatFloat(gain.xirr)}</td>
</tr>
`;
  });
}

function renderOverview(gains: Gain[]) {
  gains = _.sortBy(gains, (g) => g.account);
  const BAR_HEIGHT = 16;
  const id = "#d3-gain-overview";
  const svg = d3.select(id),
    margin = { top: 40, right: 30, bottom: 80, left: 150 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    height = gains.length * BAR_HEIGHT,
    g = svg
      .append("g")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")");
  svg.attr("height", height + margin.top + margin.bottom);

  const y = d3.scaleBand().range([0, height]).paddingInner(0.1).paddingOuter(0);
  y.domain(gains.map((g) => restName(g.account)));

  const getInvestmentAmount = (g: Gain) =>
    _.last(g.overview_timeline).investment_amount;

  const getGainAmount = (g: Gain) => _.last(g.overview_timeline).gain_amount;

  const maxInvestment = _.chain(gains).map(getInvestmentAmount).max().value();
  const maxGain = _.chain(gains).map(getGainAmount).max().value();
  const maxLoss = _.min([_.chain(gains).map(getGainAmount).min().value(), 0]);
  const maxX = maxInvestment + maxGain + Math.abs(maxLoss);
  const x = d3.scaleLinear().range([0, width]);
  x.domain([0, maxX]);
  const x1 = d3
    .scaleLinear()
    .range([0, x(maxInvestment)])
    .domain([0, maxInvestment]);
  const x2 = d3
    .scaleLinear()
    .range([x(maxInvestment), width])
    .domain([maxLoss, maxGain]);

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x1)
        .tickSize(-height)
        .tickFormat(
          skipTicks(
            50,
            x(maxInvestment),
            x1.ticks().length,
            formatCurrencyCrude
          )
        )
    );
  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x2)
        .tickSize(-height)
        .tickFormat(
          skipTicks(
            50,
            width - x(maxInvestment),
            x2.ticks().length,
            formatCurrencyCrude
          )
        )
    );

  g.append("g").attr("class", "axis y dark").call(d3.axisLeft(y));

  g.append("g")
    .selectAll("rect")
    .data(gains)
    .enter()
    .append("rect")
    .attr("fill", lineScale("investment"))
    .attr("x", x(0))
    .attr("y", (g) => y(restName(g.account)))
    .attr("height", y.bandwidth())
    .attr("width", (g) => x(getInvestmentAmount(g)));

  g.append("g")
    .selectAll("rect")
    .data(gains)
    .enter()
    .append("rect")
    .attr("fill", (g) =>
      getGainAmount(g) < 0 ? areaScale("loss") : areaScale("gain")
    )
    .attr("x", (g) => (getGainAmount(g) < 0 ? x2(getGainAmount(g)) : x2(0)))
    .attr("y", (g) => y(restName(g.account)))
    .attr("height", y.bandwidth())
    .attr("width", (g) => x(Math.abs(getGainAmount(g))));

  g.append("g")
    .selectAll("rect")
    .data(gains)
    .enter()
    .append("rect")
    .attr("fill", "transparent")
    .attr("data-tippy-content", (g: Gain) => {
      const current = _.last(g.overview_timeline);
      return tooltip([
        ["Account", [g.account, "has-text-weight-bold has-text-right"]],
        [
          "Investment",
          [
            formatCurrency(current.investment_amount),
            "has-text-weight-bold has-text-right"
          ]
        ],
        [
          "Withdrawal",
          [
            formatCurrency(current.withdrawal_amount),
            "has-text-weight-bold has-text-right"
          ]
        ],
        [
          "Gain",
          [
            formatCurrency(current.gain_amount),
            "has-text-weight-bold has-text-right"
          ]
        ],
        [
          "Balance",
          [
            formatCurrency(
              current.investment_amount +
                current.gain_amount -
                current.withdrawal_amount
            ),
            "has-text-weight-bold has-text-right"
          ]
        ],
        ["XIRR", [formatFloat(g.xirr), "has-text-weight-bold has-text-right"]]
      ]);
    })
    .attr("x", x(0))
    .attr("y", (g) => y(restName(g.account)))
    .attr("height", y.bandwidth())
    .attr("width", x(maxX));
}

function renderPerAccountOverview(gains: Gain[]) {
  const start = _.min(
      _.flatMap(gains, (g) => _.map(g.overview_timeline, (o) => o.timestamp))
    ),
    end = dayjs();

  const divs = d3
    .select("#d3-gain-timeline-breakdown")
    .selectAll("div")
    .data(_.sortBy(gains, (g) => g.account));

  divs.exit().remove();

  const columns = divs.enter().append("div").attr("class", "columns");

  const leftColumn = columns.append("div").attr("class", "column is-2");
  leftColumn
    .append("table")
    .attr("class", "table is-narrow is-fullwidth is-size-7")
    .append("tbody")
    .each(renderTable);

  const rightColumn = columns.append("div").attr("class", "column is-10");
  rightColumn
    .append("svg")
    .attr("width", "100%")
    .attr("height", "150")
    .each(function (gain) {
      renderOverviewSmall(gain.overview_timeline, this, [start, end]);
    });
}

function renderOverviewSmall(
  points: Overview[],
  element: Element,
  xDomain: [dayjs.Dayjs, dayjs.Dayjs]
) {
  const svg = d3.select(element),
    margin = { top: 5, right: 80, bottom: 20, left: 40 },
    width = element.parentElement.clientWidth - margin.left - margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg
      .append("g")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const x = d3.scaleTime().range([0, width]).domain(xDomain),
    y = d3
      .scaleLinear()
      .range([height, 0])
      .domain([
        0,
        d3.max<Overview, number>(
          points,
          (d) => d.gain_amount + d.investment_amount
        )
      ]),
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
    .call(
      d3.axisRight(y).ticks(5).tickPadding(5).tickFormat(formatCurrencyCrude)
    );

  g.append("g")
    .attr("class", "axis y")
    .call(
      d3.axisLeft(y).ticks(5).tickSize(-width).tickFormat(formatCurrencyCrude)
    );

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
        return y(d.gain_amount + d.investment_amount);
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
        return y(d.gain_amount + d.investment_amount);
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
        return y(d.investment_amount);
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
        .line<Overview>()
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
        .line<Overview>()
        .curve(d3.curveBasis)
        .defined((d) => d.withdrawal_amount > 0)
        .x((d) => x(d.timestamp))
        .y((d) => y(d.withdrawal_amount))
    );

  layer
    .append("path")
    .style("stroke", lineScale("balance"))
    .style("fill", "none")
    .attr(
      "d",
      d3
        .line<Overview>()
        .curve(d3.curveBasis)
        .x((d) => x(d.timestamp))
        .y((d) => y(d.investment_amount + d.gain_amount - d.withdrawal_amount))
    );
}

function renderLegend() {
  const svg = d3.select("#d3-gain-legend");
  svg
    .append("g")
    .attr("class", "legendOrdinal")
    .attr("transform", "translate(315,3)");

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
    .attr("transform", "translate(30,3)");

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
}
