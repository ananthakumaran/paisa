import chroma from "chroma-js";
import * as d3 from "d3";
import { Delaunay } from "d3";
import legend from "d3-svg-legend";
import dayjs from "dayjs";
import _ from "lodash";
import COLORS from "./colors";
import tippy from "tippy.js";
import {
  formatCurrency,
  formatCurrencyCrude,
  formatFloat,
  type Gain,
  type Overview,
  tooltip,
  skipTicks,
  restName,
  type Posting
} from "./utils";
import { goto } from "$app/navigation";

const areaKeys = ["gain", "loss"];
const colors = [COLORS.gain, COLORS.loss];
const areaScale = d3.scaleOrdinal<string>().domain(areaKeys).range(colors);
const lineKeys = ["balance", "investment", "withdrawal"];
const typeScale = d3
  .scaleOrdinal<string>()
  .domain(lineKeys)
  .range([COLORS.primary, COLORS.secondary, COLORS.tertiary]);

export function renderOverview(gains: Gain[]) {
  gains = _.sortBy(gains, (g) => g.account);
  const BAR_HEIGHT = 15;
  const id = "#d3-gain-overview";
  const svg = d3.select(id),
    margin = { top: 20, right: 20, bottom: 10, left: 150 },
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

  const y2 = d3
    .scaleBand()
    .range([0, y.bandwidth()])
    .domain(["0", "1"])
    .paddingInner(0)
    .paddingOuter(0.6);

  const keys = ["balance", "investment", "withdrawal", "gain", "loss"];
  const colors = [COLORS.primary, COLORS.secondary, COLORS.tertiary, COLORS.gain, COLORS.loss];
  const z = d3.scaleOrdinal<string>(colors).domain(keys);

  const getInvestmentAmount = (g: Gain) => g.overview.investment_amount;

  const getGainAmount = (g: Gain) => g.overview.gain_amount;
  const getWithdrawalAmount = (g: Gain) => g.overview.withdrawal_amount;

  const getBalanceAmount = (g: Gain) => g.overview.balance_amount;

  const maxX = _.chain(gains)
    .map((g) => getInvestmentAmount(g) + _.max([getGainAmount(g), 0]))
    .max()
    .value();
  const xirrWidth = 250;
  const xirrTextWidth = 40;
  const xirrMargin = 20;
  const textGroupWidth = 225;
  const textGroupZero = xirrWidth + xirrTextWidth + xirrMargin;

  const x = d3.scaleLinear().range([textGroupZero + textGroupWidth, width]);
  x.domain([0, maxX]);
  const x1 = d3
    .scaleLinear()
    .range([0, xirrWidth])
    .domain([
      _.min([_.min(_.map(gains, (g) => g.xirr)), 0]),
      _.max([0, _.max(_.map(gains, (g) => g.xirr))])
    ]);

  g.append("line")
    .classed("svg-grey-lightest", true)
    .attr("x1", xirrWidth + xirrTextWidth + xirrMargin / 2)
    .attr("y1", 0)
    .attr("x2", xirrWidth + xirrTextWidth + xirrMargin / 2)
    .attr("y2", height);

  g.append("line")
    .classed("svg-grey-lightest", true)
    .attr("x1", 0)
    .attr("y1", height)
    .attr("x2", width)
    .attr("y2", height);

  svg
    .append("text")
    .classed("svg-text-grey", true)
    .text("XIRR")
    .attr("text-anchor", "middle")
    .attr("x", margin.left + xirrWidth / 2)
    .attr("y", 10);

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x)
        .tickSize(-height)
        .tickFormat(skipTicks(60, x, formatCurrencyCrude))
    );

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x1)
        .tickSize(-height)
        .tickFormat(skipTicks(40, x1, (n: number) => formatFloat(n, 1)))
    );

  g.append("g").attr("class", "axis y dark link").call(d3.axisLeft(y));

  g.selectAll(".axis.y.dark.link .tick").on("click", (_event, label) => {
    goto(`/assets/gain/Assets:${label}`);
  });

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
    .attr("dominant-baseline", "hanging")
    .attr("text-anchor", "end")
    .style("fill", (g) => (getInvestmentAmount(g) > 0 ? z("investment") : "none"))
    .attr("dx", "-3")
    .attr("dy", "3")
    .attr("x", textGroupZero + textGroupWidth / 3)
    .attr("y", (g) => y(restName(g.account)));

  textGroup
    .append("text")
    .text((g) => formatCurrency(getGainAmount(g)))
    .attr("dominant-baseline", "hanging")
    .attr("text-anchor", "end")
    .style("fill", (g) => (getGainAmount(g) > 0 ? chroma(z("gain")).darken().hex() : "none"))
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
    .text((g) => formatCurrency(getGainAmount(g)))
    .attr("text-anchor", "end")
    .style("fill", (g) => (getGainAmount(g) < 0 ? chroma(z("loss")).darken().hex() : "none"))
    .attr("dx", "-3")
    .attr("dy", "-3")
    .attr("x", textGroupZero + (textGroupWidth * 2) / 3)
    .attr("y", (g) => y(restName(g.account)) + y.bandwidth());

  textGroup
    .append("text")
    .text((g) => formatCurrency(getWithdrawalAmount(g)))
    .attr("text-anchor", "end")
    .style("fill", (g) => (getWithdrawalAmount(g) > 0 ? z("withdrawal") : "none"))
    .attr("dx", "-3")
    .attr("dy", "-3")
    .attr("x", textGroupZero + textGroupWidth)
    .attr("y", (g) => y(restName(g.account)) + y.bandwidth());

  textGroup
    .append("line")
    .classed("svg-grey-lightest", true)
    .attr("x1", 0)
    .attr("y1", (g) => y(restName(g.account)))
    .attr("x2", width)
    .attr("y2", (g) => y(restName(g.account)));

  textGroup
    .append("text")
    .text((g) => formatFloat(g.xirr))
    .attr("text-anchor", "end")
    .attr("dominant-baseline", "middle")
    .style("fill", (g) =>
      g.xirr < 0 ? chroma(z("loss")).darken().hex() : chroma(z("gain")).darken().hex()
    )
    .attr("x", xirrWidth + xirrTextWidth)
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
    .attr("rx", "5")
    .attr("fill", (d) => {
      return z(d.key);
    })
    .attr("x", (d) => x(d[0][0]))
    .attr("y", (d: any) => y2(d[0].data.i))
    .attr("height", y2.bandwidth())
    .attr("width", (d) => x(d[0][1]) - x(d[0][0]));

  const paddingTop = (y1.range()[1] - y1.bandwidth() * 2) / 2;
  g.append("g")
    .selectAll("rect")
    .data(gains)
    .enter()
    .append("rect")
    .attr("fill", (g) => (g.xirr < 0 ? z("loss") : z("gain")))
    .attr("x", (g) => (g.xirr < 0 ? x1(g.xirr) : x1(0)))
    .attr("y", (g) => y(restName(g.account)) + paddingTop)
    .attr("height", y.bandwidth() - paddingTop * 2)
    .attr("width", (g) => Math.abs(x1(0) - x1(g.xirr)));

  g.append("g")
    .selectAll("rect")
    .data(gains)
    .enter()
    .append("rect")
    .attr("fill", "transparent")
    .attr("data-tippy-content", (g: Gain) => {
      const current = g.overview;
      return tooltip([
        ["Account", [g.account, "has-text-weight-bold has-text-right"]],
        [
          "Investment",
          [formatCurrency(current.investment_amount), "has-text-weight-bold has-text-right"]
        ],
        [
          "Withdrawal",
          [formatCurrency(current.withdrawal_amount), "has-text-weight-bold has-text-right"]
        ],
        ["Gain", [formatCurrency(current.gain_amount), "has-text-weight-bold has-text-right"]],
        [
          "Balance",
          [formatCurrency(current.balance_amount), "has-text-weight-bold has-text-right"]
        ],
        ["XIRR", [formatFloat(g.xirr), "has-text-weight-bold has-text-right"]]
      ]);
    })
    .attr("x", 0)
    .attr("y", (g) => y(restName(g.account)))
    .attr("height", y.bandwidth())
    .attr("width", width);
}

export function renderAccountOverview(points: Overview[], postings: Posting[], id: string) {
  const start = _.min(_.map(points, (p) => p.date)),
    end = dayjs();

  const element = document.getElementById(id);

  const svg = d3.select(element),
    margin = { top: 5, right: 50, bottom: 20, left: 40 },
    width = element.parentElement.clientWidth - margin.left - margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const areaKeys = ["gain", "loss"];
  const colors = [COLORS.gain, COLORS.loss];

  const lineKeys = ["balance", "investment"];
  const lineScale = d3
    .scaleOrdinal<string>()
    .domain(lineKeys)
    .range([COLORS.primary, COLORS.secondary]);

  const positions = _.flatMap(points, (p) => [p.balance_amount, p.net_investment_amount]);
  positions.push(0);

  const x = d3.scaleTime().range([0, width]).domain([start, end]),
    y = d3.scaleLinear().range([height, 0]).domain(d3.extent(positions)),
    z = d3.scaleOrdinal<string>(colors).domain(areaKeys);

  const area = (y0: number, y1: (d: Overview) => number) =>
    d3
      .area<Overview>()
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

  const postingsG = g.append("g").attr("class", "postings");

  postingsG
    .selectAll("circle")
    .data(postings)
    .join("circle")
    .attr("data-tippy-content", (p) => {
      return tooltip(
        [
          ["Date", p.date.format("DD MMM YYYY")],
          ["Amount", [formatCurrency(p.amount), "has-text-weight-bold has-text-right"]]
        ],
        { header: p.payee }
      );
    })
    .attr("cx", (d) => x(d.date))
    .attr("cy", height + 3)
    .attr("r", 3)
    .attr("opacity", 0.5)
    .attr("fill", (d) => (d.amount >= 0 ? typeScale("investment") : typeScale("withdrawal")));

  const layer = g.selectAll(".layer").data([points]).enter().append("g").attr("class", "layer");

  const clipAboveID = _.uniqueId("clip-above");
  layer
    .append("clipPath")
    .attr("id", clipAboveID)
    .append("path")
    .attr(
      "d",
      area(height, (d) => {
        return y(d.gain_amount + d.investment_amount - d.withdrawal_amount);
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
        return y(d.gain_amount + d.investment_amount - d.withdrawal_amount);
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
        return y(d.investment_amount - d.withdrawal_amount);
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
        return y(d.investment_amount - d.withdrawal_amount);
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
        .x((d) => x(d.date))
        .y((d) => y(d.net_investment_amount))
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
        .x((d) => x(d.date))
        .y((d) => y(d.balance_amount))
    );

  const hoverCircle = layer.append("circle").attr("r", "3").attr("fill", "none");
  const t = tippy(hoverCircle.node(), { theme: "light", delay: 0, allowHTML: true });

  const balanceVoronoiPoints = _.map(points, (d) => [x(d.date), y(d.balance_amount)]);
  const investmentVoronoiPoints = _.map(points, (d) => [x(d.date), y(d.net_investment_amount)]);
  const voronoi = Delaunay.from(balanceVoronoiPoints.concat(investmentVoronoiPoints)).voronoi([
    0,
    0,
    width,
    height
  ]);

  layer
    .append("g")
    .selectAll("path")
    .data(
      points.map((p) => ["balance", p]).concat(points.map((p) => ["investment", p])) as [
        string,
        Overview
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
        .attr("cy", y(pointType == "balance" ? d.balance_amount : d.net_investment_amount))
        .attr("fill", lineScale(pointType));

      t.setProps({
        placement: pointType == "balance" ? "top" : "bottom",
        content: tooltip([
          ["Date", d.date.format("DD MMM YYYY")],
          ["Balance", [formatCurrency(d.balance_amount), "has-text-weight-bold has-text-right"]],
          [
            "Net Investment",
            [formatCurrency(d.net_investment_amount), "has-text-weight-bold has-text-right"]
          ],
          ["Gain / Loss", [formatCurrency(d.gain_amount), "has-text-weight-bold has-text-right"]]
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

export function renderLegend() {
  const svg = d3.select("#d3-gain-legend");
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
    .scale(typeScale);

  svg.select(".legendLine").call(legendLine as any);
}
