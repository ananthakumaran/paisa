import * as d3 from "d3";
import COLORS from "./colors";
import { ajax, Issue, setHtml } from "./utils";

export default async function () {
  const { issues: issues } = await ajax("/api/diagnosis");
  setHtml(
    "diagnosis-count",
    issues.length.toString(),
    issues.length > 0 ? COLORS.lossText : COLORS.gainText
  );
  renderIssues(issues);
}

function renderIssues(issues: Issue[]) {
  const id = "#d3-diagnosis";
  const root = d3.select(id);

  `<article class="message is-danger">
  <div class="message-header">
    <p>Danger</p>
    <button class="delete" aria-label="delete"></button>
  </div>
  <div class="message-body">
    Lorem ipsum dolor sit amet, consectetur adipiscing elit. <strong>Pellentesque risus mi</strong>, tempus quis placerat ut, porta nec nulla. Vestibulum rhoncus ac ex sit amet fringilla. Nullam gravida purus diam, et dictum <a>felis venenatis</a> efficitur. Aenean ac <em>eleifend lacus</em>, in mollis lectus. Donec sodales, arcu et sollicitudin porttitor, tortor urna tempor ligula, id porttitor mi magna a neque. Donec dui urna, vehicula et sem eget, facilisis sodales sem.
  </div>
</article>`;

  const issue = root
    .selectAll("div")
    .data(issues)
    .enter()
    .append("div")
    .attr("class", "column is-6")
    .append("div")
    .attr("class", (i) => `message is-${i.level}`);

  issue
    .append("div")
    .attr("class", "message-header")
    .html((i) => `<p>${i.summary}</p>`);

  issue
    .append("div")
    .attr("class", "message-body")
    .html((i) => `${i.description} <br/> <br/> ${i.details}`);
}
