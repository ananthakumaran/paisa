import * as d3 from "d3";
import dayjs from "dayjs";
import _ from "lodash";
import {
  ajax,
  Breakdown,
  depth,
  formatCurrency,
  formatFloat,
  lastName,
  Posting
} from "./utils";

export default async function () {
  const { postings: postings, breakdowns: breakdowns } = await ajax(
    "/api/ledger"
  );
  _.each(postings, (p) => (p.timestamp = dayjs(p.date)));

  renderBreakdowns(breakdowns);
  renderTransactions(postings);

  d3.select("input.d3-posting-filter").on(
    "input",
    _.debounce((event) => {
      const text = event.srcElement.value;
      renderTransactions(filterTransactions(postings, text));
    }, 300)
  );
}

function renderTransactions(postings: Posting[]) {
  const tbody = d3.select(".d3-postings");
  const trs = tbody.selectAll("tr").data(postings);

  trs.exit().remove();

  trs
    .enter()
    .append("tr")
    .merge(trs as any)
    .attr("class", (t) =>
      t.timestamp.month() % 2 == 0 ? "has-background-white-ter" : ""
    )
    .html((p) => {
      const purchase = formatCurrency(p.amount);
      let market = "",
        change = "",
        changePercentage = "",
        changeClass = "";
      if (p.commodity !== "INR") {
        market = formatCurrency(p.market_amount);
        const changeAmount = p.market_amount - p.amount;
        if (changeAmount > 0) {
          changeClass = "has-text-success";
        } else if (changeAmount < 0) {
          changeClass = "has-text-danger";
        }
        const perYear = 365 / dayjs().diff(p.timestamp, "days");
        changePercentage = formatFloat(
          (changeAmount / p.amount) * 100 * perYear
        );
        change = formatCurrency(changeAmount);
      }
      return `
       <td>${p.timestamp.format("DD MMM YYYY")}</td>
       <td>${p.account}</td>
       <td class='has-text-right'>${purchase}</td>
       <td class='has-text-right'>${market}</td>
       <td class='${changeClass} has-text-right'>${change}</td>
       <td class='${changeClass} has-text-right'>${changePercentage}</td>
      `;
    });
}

function filterTransactions(postings: Posting[], filter: string) {
  let filterRegex = new RegExp(".*", "i");
  if (filter) {
    filterRegex = new RegExp(filter, "i");
  }

  return _.filter(postings, (t) => filterRegex.test(t.account));
}

function renderBreakdowns(breakdowns: Breakdown[]) {
  const tbody = d3.select(".d3-postings-breakdown");
  const trs = tbody.selectAll("tr").data(Object.values(breakdowns));

  trs.exit().remove();
  trs
    .enter()
    .append("tr")
    .merge(trs as any)
    .html((b) => {
      let changeClass = "";

      const gain = b.market_amount + b.withdrawal_amount - b.investment_amount;
      if (gain > 0) {
        changeClass = "has-text-success";
      } else if (gain < 0) {
        changeClass = "has-text-danger";
      }
      const indent = _.repeat("&emsp;&emsp;", depth(b.group) - 1);
      return `
       <td style='max-width: 200px; overflow: hidden;'>${indent}${lastName(
        b.group
      )}</td>
       <td class='has-text-right'>${formatCurrency(b.investment_amount)}</td>
       <td class='has-text-right'>${formatCurrency(b.withdrawal_amount)}</td>
       <td class='has-text-right'>${formatCurrency(b.market_amount)}</td>
       <td class='${changeClass} has-text-right'>${formatCurrency(gain)}</td>
       <td class='${changeClass} has-text-right'>${formatFloat(b.xirr)}</td>
      `;
    });
}
