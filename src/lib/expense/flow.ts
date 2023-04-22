import {
  generateColorScheme,
  type Graph,
  type Node,
  formatCurrencyCrude,
  restName
} from "$lib/utils";
import legend from "d3-svg-legend";
import * as d3 from "d3";
import { sankeyCircular, sankeyJustify } from "d3-sankey-circular";
import { pathArrows } from "d3-path-arrows";
import _ from "lodash";
import { firstName } from "$lib/utils";
import { tooltip } from "$lib/utils";
import { formatCurrency } from "$lib/utils";
import { iconify } from "$lib/icon";
import { willClearTippy } from "../../store";

export function renderFlow(graph: Graph) {
  const id = "#d3-expense-flow";
  const svg = d3.select(id);
  const margin = { top: 60, right: 40, bottom: 40, left: 40 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom;

  willClearTippy.update((n) => n + 1);
  svg.selectAll("*").remove();

  const accounts = _.chain(graph.nodes)
    .map((n) => firstName(n.name))
    .uniq()
    .sort()
    .value();
  const color = generateColorScheme(accounts);

  svg.append("g").attr("class", "legendOrdinal").attr("transform", "translate(50,0)");

  const legendOrdinal = legend
    .legendColor()
    .shape("rect")
    .orient("horizontal")
    .shapePadding(50)
    .labels(accounts)
    .scale(color);

  svg.select(".legendOrdinal").call(legendOrdinal as any);

  const g = svg
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom)
    .append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const sankey = sankeyCircular()
    .nodeWidth(20)
    .nodePaddingRatio(0.7)
    .size([width, height])
    .nodeId((d: Node) => d.id)
    .nodeAlign(sankeyJustify)
    .iterations(64)
    .circularLinkGap(2);

  const linkG = g
    .append("g")
    .attr("class", "links")
    .attr("fill", "none")
    .attr("stroke-opacity", 0.2)
    .selectAll("path");

  const nodeG = g
    .append("g")
    .attr("class", "nodes")
    .attr("font-family", "sans-serif")
    .attr("font-size", 10)
    .selectAll("g");

  const sankeyData = sankey(graph);
  const sankeyNodes = sankeyData.nodes;
  const sankeyLinks = sankeyData.links;

  d3.extent(sankeyNodes, (d: any) => d.depth);

  const node = nodeG.data(sankeyNodes).enter().append("g");

  node
    .append("rect")
    .attr("x", (d: any) => d.x0)
    .attr("y", (d: any) => d.y0)
    .attr("height", (d: any) => d.y1 - d.y0)
    .attr("width", (d: any) => d.x1 - d.x0)
    .style("fill", (d: any) => color(firstName(d.name)))
    .attr("data-tippy-content", (d: any) =>
      tooltip([
        ["Account", iconify(d.name)],
        ["Total", [formatCurrency(d.value), "has-text-right has-text-weight-bold"]]
      ])
    );

  node
    .append("text")
    .attr("x", (d: any) => (d.x0 + d.x1) / 2)
    .attr("y", (d: any) => d.y0 - 12)
    .attr("dx", (d: any) => {
      if (_.isEmpty(d.sourceLinks)) {
        return "10px";
      }

      if (_.isEmpty(d.targetLinks)) {
        return "-10px";
      }

      return "0px";
    })
    .attr("dy", "0.35em")
    .attr("text-anchor", (d: any) => {
      if (_.isEmpty(d.sourceLinks)) {
        return "end";
      }

      if (_.isEmpty(d.targetLinks)) {
        return "start";
      }

      return "middle";
    })
    .attr("fill", "#4a4a4a")
    .text(
      (d: any) =>
        `${iconify(restName(d.name), { group: firstName(d.name) })} ${formatCurrencyCrude(d.value)}`
    );

  const link = linkG.data(sankeyLinks).enter().append("g");

  link
    .append("path")
    .attr("class", "sankey-link")
    .attr("d", (link: any) => link.path)
    .style("stroke-width", (d: any) => Math.max(1, d.width))
    .style("opacity", 0.5)
    .style("stroke", (d: any) => color(firstName(d.target.name)))
    .attr("data-tippy-content", (d: any) =>
      tooltip([
        ["Debit Account", iconify(d.source.name)],
        ["Credit Account", iconify(d.target.name)],
        ["Total", [formatCurrency(d.value), "has-text-right has-text-weight-bold"]]
      ])
    );

  const arrows = pathArrows()
    .arrowLength(10)
    .gapLength(150)
    .arrowHeadSize(3)
    .path((link: any) => link.path);

  linkG.data(sankeyLinks).enter().append("g").attr("class", "g-arrow").call(arrows);
}
