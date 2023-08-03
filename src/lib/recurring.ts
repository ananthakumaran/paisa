import * as d3 from "d3";
import dayjs from "dayjs";
import _ from "lodash";
import COLORS from "./colors";
import { skipTicks, type TransactionSequence } from "./utils";

export function renderRecurring(
  element: Element,
  transactionSequence: TransactionSequence,
  next: dayjs.Dayjs,
  showPage: (pageIndex: number) => void
) {
  const svg = d3.select(element).select("svg"),
    margin = { top: 20, right: 40, bottom: 20, left: 30 },
    width = element.parentElement.clientWidth - margin.left - margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const transactions = _.reverse(_.take(transactionSequence.transactions, 3));

  const dates = _.map(transactions, (d) => d.date).concat([next, dayjs()]);
  const [start, end] = d3.extent(dates);
  dates.pop();
  const x = d3.scaleTime().domain([start, end]).range([0, width]);

  g.append("g")
    .attr("class", "axis x")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x)
        .tickValues(dates)
        .tickFormat(skipTicks(45, x, (d: any) => dayjs(d).format("DD MMM")))
    );

  g.append("g")
    .selectAll("circle.past")
    .data(transactions)
    .join("circle")
    .attr("class", "past")
    .attr("r", 5)
    .attr("fill", "#d5d5d5")
    .attr("cx", (d) => x(d.date))
    .attr("cy", "10")
    .style("cursor", "pointer")
    .on("click", (_event, d) =>
      showPage(_.findIndex(transactionSequence.transactions, (t) => t.id === d.id))
    );

  function color(d: dayjs.Dayjs) {
    return d.isAfter(dayjs()) ? COLORS.success : COLORS.danger;
  }

  g.append("g")
    .selectAll("circle.next")
    .data([next])
    .join("circle")
    .attr("class", "next")
    .attr("r", 5)
    .attr("fill", color)
    .attr("cx", (d) => x(d))
    .attr("cy", "10");

  g.append("g")
    .selectAll("line")
    .data([next])
    .join("line")
    .attr("class", "next")
    .attr("stroke", color)
    .attr("stroke-linecap", "round")
    .attr("stroke-width", 3)
    .attr("x1", (d) => x(d))
    .attr("x2", x(dayjs()))
    .attr("y1", "10")
    .attr("y2", height);
}
