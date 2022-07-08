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
  const xirrTextWidth = 40;
  const BAR_HEIGHT = 13;
  const id = "#d3-gain-overview";
  const svg = d3.select(id),
    margin = { top: 40, right: 40 + xirrTextWidth, bottom: 80, left: 150 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    height = gains.length * BAR_HEIGHT * 2,
    g = svg
      .append("g")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")");
  svg.attr("height", height + margin.top + margin.bottom);

  const y = d3.scaleBand().range([0, height]).paddingInner(0).paddingOuter(0);
  y.domain(gains.map((g) => restName(g.account)));
  const y1 = d3
    .scaleBand()
    .range([0, y.bandwidth()])
    .domain(["0", "1"])
    .paddingInner(0)
    .paddingOuter(0);

  const keys = ["balance", "investment", "withdrawal", "gain", "loss"];
  const colors = ["#1f77b4", "#17becf", "#ff7f0e", "#b2df8a", "#fb9a99"];
  const z = d3.scaleOrdinal<string>(colors).domain(keys);

  const getInvestmentAmount = (g: Gain) =>
    _.last(g.overview_timeline).investment_amount;

  const getGainAmount = (g: Gain) => _.last(g.overview_timeline).gain_amount;
  const getWithdrawalAmount = (g: Gain) =>
    _.last(g.overview_timeline).withdrawal_amount;

  const getBalanceAmount = (g: Gain) => {
    const current = _.last(g.overview_timeline);
    return (
      current.investment_amount +
      current.gain_amount -
      current.withdrawal_amount
    );
  };

  const maxInvestment = _.chain(gains).map(getInvestmentAmount).max().value();
  const maxX = _.chain(gains)
    .map((g) => getInvestmentAmount(g) + _.max([getGainAmount(g), 0]))
    .max()
    .value();
  const xirrWidth = 250;
  const textGroupWidth = 225;
  const x = d3.scaleLinear().range([0, width - xirrWidth - textGroupWidth]);
  x.domain([0, maxX]);
  const x1 = d3
    .scaleLinear()
    .range([width - xirrWidth + 20, width])
    .domain([
      _.min([_.min(_.map(gains, (g) => g.xirr)), 0]),
      _.max([0, _.max(_.map(gains, (g) => g.xirr))])
    ]);

  g.append("line")
    .attr("stroke", "#ddd")
    .attr("x1", x(maxX) + textGroupWidth + 10)
    .attr("y1", 0)
    .attr("x2", x(maxX) + textGroupWidth + 10)
    .attr("y2", height);

  g.append("line")
    .attr("stroke", "#ddd")
    .attr("x1", 0)
    .attr("y1", height)
    .attr("x2", width + xirrTextWidth)
    .attr("y2", height);

  g.append("text")
    .attr("fill", "#4a4a4a")
    .text("XIRR")
    .attr("text-anchor", "middle")
    .attr("x", width + 20 - xirrWidth / 2)
    .attr("y", height + 40);

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x)
        .tickSize(-height)
        .tickFormat(
          skipTicks(50, x(maxInvestment), x.ticks().length, formatCurrencyCrude)
        )
    );
  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x1)
        .tickSize(-height)
        .tickFormat(
          skipTicks(40, xirrWidth, x1.ticks().length, (n) => formatFloat(n, 1))
        )
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
    .text((g) => formatCurrency(getInvestmentAmount(g)))
    .attr("alignment-baseline", "hanging")
    .attr("text-anchor", "end")
    .style("fill", (g) =>
      getInvestmentAmount(g) > 0 ? z("investment") : "none"
    )
    .attr("dx", "-3")
    .attr("dy", "3")
    .attr("x", x(maxX) + textGroupWidth / 3)
    .attr("y", (g) => y(restName(g.account)));

  textGroup
    .append("text")
    .text((g) => formatCurrency(getGainAmount(g)))
    .attr("alignment-baseline", "hanging")
    .attr("text-anchor", "end")
    .style("fill", (g) => (getGainAmount(g) > 0 ? z("gain") : "none"))
    .attr("dx", "3")
    .attr("dy", "3")
    .attr("x", x(maxX) + (textGroupWidth * 2) / 3)
    .attr("y", (g) => y(restName(g.account)));

  textGroup
    .append("text")
    .text((g) => formatCurrency(getBalanceAmount(g)))
    .attr("text-anchor", "end")
    .style("fill", (g) => (getBalanceAmount(g) > 0 ? z("balance") : "none"))
    .attr("dx", "-3")
    .attr("dy", "-3")
    .attr("x", x(maxX) + textGroupWidth / 3)
    .attr("y", (g) => y(restName(g.account)) + y.bandwidth());

  textGroup
    .append("text")
    .text((g) => formatCurrency(getGainAmount(g)))
    .attr("text-anchor", "end")
    .style("fill", (g) => (getGainAmount(g) < 0 ? z("loss") : "none"))
    .attr("dx", "-3")
    .attr("dy", "-3")
    .attr("x", x(maxX) + (textGroupWidth * 2) / 3)
    .attr("y", (g) => y(restName(g.account)) + y.bandwidth());

  textGroup
    .append("text")
    .text((g) => formatCurrency(getWithdrawalAmount(g)))
    .attr("text-anchor", "end")
    .style("fill", (g) =>
      getWithdrawalAmount(g) > 0 ? z("withdrawal") : "none"
    )
    .attr("dx", "3")
    .attr("dy", "-3")
    .attr("x", x(maxX) + textGroupWidth)
    .attr("y", (g) => y(restName(g.account)) + y.bandwidth());

  textGroup
    .append("line")
    .attr("stroke", "#ddd")
    .attr("x1", x(0))
    .attr("y1", (g) => y(restName(g.account)))
    .attr("x2", x(0) + width + xirrTextWidth)
    .attr("y2", (g) => y(restName(g.account)));

  textGroup
    .append("text")
    .text((g) => formatFloat(g.xirr))
    .attr("text-anchor", "end")
    .attr("alignment-baseline", "middle")
    .style("fill", (g) => (g.xirr < 0 ? z("loss") : z("gain")))
    .attr("x", width + xirrTextWidth)
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
      d3.stack().keys(["investment", "gain"])([
        {
          i: "0",
          data: g,
          investment: getInvestmentAmount(g),
          gain: _.max([getGainAmount(g), 0])
        }
      ] as any),
      d3.stack().keys(["balance", "loss", "withdrawal"])([
        {
          i: "1",
          data: g,
          balance: getBalanceAmount(g),
          withdrawal: getWithdrawalAmount(g),
          loss: Math.abs(_.min([getGainAmount(g), 0]))
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

  g.append("g")
    .selectAll("rect")
    .data(gains)
    .enter()
    .append("rect")
    .attr("fill", (g) => (g.xirr < 0 ? z("loss") : z("gain")))
    .attr("x", (g) => (g.xirr < 0 ? x1(g.xirr) : x1(0)))
    .attr("y", (g) => y(restName(g.account)))
    .attr("height", y.bandwidth())
    .attr("width", (g) => Math.abs(x1(0) - x1(g.xirr)));

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
    .attr("width", width);
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
    .attr("transform", "translate(280,3)");

  const legendOrdinal = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(70)
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
    .labels(lineKeys)
    .scale(lineScale);

  svg.select(".legendLine").call(legendLine as any);
}
