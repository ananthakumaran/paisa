import * as d3 from "d3";
import dayjs from "dayjs";
import _ from "lodash";
import { skipTicks, type TransactionSequence } from "./utils";
import { scheduleIcon } from "./transaction_sequence";

export function renderRecurring(
  element: Element,
  transactionSequence: TransactionSequence,
  showPage: (pageIndex: number) => void
) {
  const svg = d3.select(element).select("svg"),
    margin = { top: 20, right: 40, bottom: 20, left: 30 },
    width = element.parentElement.clientWidth - margin.left - margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  let schedules = _.takeRight(transactionSequence.pastSchedules, 3);
  schedules = schedules.concat(_.take(transactionSequence.futureSchedules, 1));

  const dates = _.map(schedules, (s) => s.scheduled);
  const [start, end] = d3.extent(dates);
  const x = d3.scaleTime().domain([start, end]).range([0, width]);

  g.append("g")
    .attr("class", "axis light-domain x")
    .attr("transform", "translate(0," + height + ")")
    .call(
      d3
        .axisBottom(x)
        .tickValues(dates)
        .tickFormat(skipTicks(45, x, (d: any) => dayjs(d).format("DD MMM YY")))
    );

  g.append("g")
    .selectAll("text")
    .data(schedules)
    .join("text")
    .attr("class", (d) => "svg-background-white " + scheduleIcon(d).svgColor)
    .attr("dx", "-0.5em")
    .attr("dy", "-0.3em")
    .attr("x", (d) => x(d.scheduled))
    .attr("y", "10")
    .style("cursor", (d) => (d.transaction ? "pointer" : "default"))
    .on("click", (_event, d) => {
      if (d.transaction) {
        showPage(_.findIndex(transactionSequence.transactions, (t) => t.id === d.transaction.id));
      }
    })
    .text((d) => scheduleIcon(d).glyph);
}
