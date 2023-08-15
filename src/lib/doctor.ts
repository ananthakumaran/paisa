import * as d3 from "d3";
import type { Issue } from "./utils";

export function renderIssues(issues: Issue[]) {
  const id = "#d3-diagnosis";
  const root = d3.select(id);

  const issue = root
    .selectAll("div")
    .data(issues)
    .enter()
    .append("div")
    .attr("class", "column is-6")
    .append("div")
    .attr("class", (i) => `message invertable is-${i.level}`);

  issue
    .append("div")
    .attr("class", "message-header")
    .html((i) => `<p>${i.summary}</p>`);

  issue
    .append("div")
    .attr("class", "message-body")
    .html((i) => `${i.description} <br/> <br/> ${i.details}`);
}
