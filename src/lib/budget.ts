import * as d3 from "d3";
import { darkLightColor, type AccountBudget } from "./utils";
import _ from "lodash";
import COLORS from "./colors";
import chroma from "chroma-js";
import textures from "textures";

export function renderBudget(element: Element, accountBudget: AccountBudget) {
  const svg = d3.select(element).select("svg"),
    margin = { top: 2, right: 10, bottom: 4, left: 10 },
    width = element.clientWidth - margin.left - margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom;

  const max = _.max([
    0,
    accountBudget.forecast,
    accountBudget.actual,
    accountBudget.actual - accountBudget.rollover
  ]);
  const x = d3.scaleLinear().domain([0, max]).range([0, width]);
  const g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const gainColor = darkLightColor(COLORS.success, COLORS.gain);
  const texture = textures
    .lines()
    .strokeWidth(2)
    .size(5)
    .background(gainColor)
    .stroke(chroma(gainColor).brighten().hex());
  svg.call(texture);

  function renderLine(amount: number, color: string) {
    if (amount === 0) return;

    g.append("g")
      .selectAll("line")
      .data([amount])
      .join("line")
      .attr("class", "next")
      .attr("stroke", color)
      .attr("stroke-linecap", "round")
      .attr("stroke-width", 5)
      .attr("x1", "0")
      .attr("x2", (d) => x(d))
      .attr("y1", height)
      .attr("y2", height);
  }

  renderLine(accountBudget.forecast, texture.url());
  let yellow = 0;
  let green = 0;
  let red = 0;

  if (accountBudget.rollover > 0 && accountBudget.actual > accountBudget.forecast) {
    const rolloverUsed = _.min([
      accountBudget.actual - accountBudget.forecast,
      accountBudget.rollover
    ]);
    yellow = rolloverUsed;
  }

  if (accountBudget.actual > accountBudget.forecast) {
    red = _.max([accountBudget.actual - accountBudget.forecast, 0]);
    if (accountBudget.rollover > 0) {
      red = _.max([red - accountBudget.rollover, 0]);
    }
  }

  green = _.min([accountBudget.forecast, accountBudget.actual]);

  renderLine(green + yellow + red, darkLightColor(COLORS.danger, COLORS.loss));
  renderLine(green + yellow, darkLightColor(COLORS.warnText, COLORS.warn));
  renderLine(green, darkLightColor(COLORS.success, COLORS.gain));
}
