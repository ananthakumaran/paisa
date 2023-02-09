import * as d3 from "d3";
import _ from "lodash";
import { iconText } from "./icon";
import { type Breakdown, depth, formatCurrency, formatFloat, lastName } from "./utils";

export function renderBreakdowns(breakdowns: Breakdown[]) {
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
       <td style='max-width: 200px; overflow: hidden;'>${indent}${iconText(b.group)} ${lastName(
        b.group
      )}</td>
       <td class='has-text-right'>${
         b.investment_amount != 0 ? formatCurrency(b.investment_amount) : ""
       }</td>
       <td class='has-text-right'>${
         b.withdrawal_amount != 0 ? formatCurrency(b.withdrawal_amount) : ""
       }</td>
       <td class='has-text-right'>${b.balance_units > 0 ? formatFloat(b.balance_units, 4) : ""}</td>
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
