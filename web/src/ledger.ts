import * as d3 from "d3";
import dayjs from "dayjs";
import _ from "lodash";
import Clusturize from "clusterize.js";
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
  const { rows, clusterTable } = renderTransactions(postings);

  d3.select("input.d3-posting-filter").on(
    "input",
    _.debounce((event) => {
      const text = event.srcElement.value;
      const filtered = filterTransactions(rows, text);
      clusterTable.update(_.map(filtered, (r) => r.markup));
    }, 100)
  );
}

function renderTransactions(postings: Posting[]) {
  const rows = _.map(postings, (p) => {
    const purchase = formatCurrency(p.amount);
    let market = "",
      date = p.timestamp.format("DD MMM YYYY"),
      change = "",
      changePercentage = "",
      changeClass = "",
      price = "",
      units = "";
    if (p.commodity !== "INR") {
      units = formatFloat(p.quantity, 4);
      price = formatCurrency(Math.abs(p.amount / p.quantity), 4);
      const days = dayjs().diff(p.timestamp, "days");
      if (p.quantity > 0 && days > 0) {
        market = formatCurrency(p.market_amount);
        const changeAmount = p.market_amount - p.amount;
        if (changeAmount > 0) {
          changeClass = "has-text-success";
        } else if (changeAmount < 0) {
          changeClass = "has-text-danger";
        }
        const perYear = 365 / days;
        changePercentage = formatFloat(
          (changeAmount / p.amount) * 100 * perYear
        );
        change = formatCurrency(changeAmount);
      }
    }
    const markup = `
<tr class="${p.timestamp.month() % 2 == 0 ? "has-background-white-ter" : ""}">
       <td>${date}</td>
       <td>${p.payee}</td>
       <td>${p.account}</td>
       <td class='has-text-right'>${purchase}</td>
       <td class='has-text-right'>${units}</td>
       <td class='has-text-right'>${price}</td>
       <td class='has-text-right'>${market}</td>
       <td class='${changeClass} has-text-right'>${change}</td>
       <td class='${changeClass} has-text-right'>${changePercentage}</td>
</tr>
`;
    return {
      date: date,
      markup: markup,
      posting: p
    };
  });

  const clusterTable = new Clusturize({
    rows: _.map(rows, (r) => r.markup),
    scrollId: "d3-postings-container",
    contentId: "d3-postings",
    rows_in_block: 100
  });

  return { rows, clusterTable };
}

function filterTransactions(
  rows: { date: string; posting: Posting; markup: string }[],
  filter: string
) {
  let filterRegex = new RegExp(".*", "i");
  if (filter) {
    filterRegex = new RegExp(filter, "i");
  }

  return _.filter(
    rows,
    (r) =>
      filterRegex.test(r.posting.account) ||
      filterRegex.test(r.posting.payee) ||
      filterRegex.test(r.date)
  );
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
       <td class='has-text-right'>${
         b.investment_amount != 0 ? formatCurrency(b.investment_amount) : ""
       }</td>
       <td class='has-text-right'>${
         b.withdrawal_amount != 0 ? formatCurrency(b.withdrawal_amount) : ""
       }</td>
       <td class='has-text-right'>${
         b.balance_units > 0 ? formatFloat(b.balance_units, 4) : ""
       }</td>
       <td class='has-text-right'>${
         b.market_amount != 0 ? formatCurrency(b.market_amount) : ""
       }</td>
       <td class='${changeClass} has-text-right'>${
        b.investment_amount != 0 && gain != 0 ? formatCurrency(gain) : ""
      }</td>
       <td class='${changeClass} has-text-right'>${
        b.xirr > 0.0001 || b.xirr < -0.0001 ? formatFloat(b.xirr) : ""
      }</td>
      `;
    });
}
