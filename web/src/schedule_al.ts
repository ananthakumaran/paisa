import * as d3 from "d3";
import dayjs from "dayjs";
import { ajax, formatCurrency, ScheduleALEntry, setHtml } from "./utils";

export default async function () {
  const { schedule_al_entries: scheduleALEntries, date: date } = await ajax(
    "/api/schedule_al"
  );

  setHtml("schedule-al-date", dayjs(date).format("DD MMM YYYY"));
  renderBreakdowns(scheduleALEntries);
}

function renderBreakdowns(scheduleALEntries: ScheduleALEntry[]) {
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
