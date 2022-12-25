import * as d3 from "d3";
import { formatCurrency, type ScheduleALEntry } from "./utils";

export function renderBreakdowns(scheduleALEntries: ScheduleALEntry[]) {
  const tbody = d3.select(".d3-schedule-al");
  const trs = tbody.selectAll("tr").data(Object.values(scheduleALEntries));

  trs.exit().remove();
  trs
    .enter()
    .append("tr")
    .merge(trs as any)
    .html((s) => {
      return `
       <td>${s.section.code}</td>
       <td>${s.section.section}</td>
       <td>${s.section.details}</td>
       <td class='has-text-right'>${formatCurrency(s.amount)}</td>
      `;
    });
}
