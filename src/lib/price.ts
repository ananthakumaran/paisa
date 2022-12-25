import * as d3 from "d3";
import { type Price, formatCurrency } from "./utils";

export function renderPrices(prices: Price[]) {
  const tbody = d3.select(".d3-prices");
  const trs = tbody.selectAll("tr").data(prices);

  trs.exit().remove();
  trs
    .enter()
    .append("tr")
    .merge(trs as any)
    .html((p) => {
      return `
       <td>${p.commodity_name}</td>
       <td>${p.commodity_type}</td>
       <td>${p.commodity_id}</td>
       <td>${p.timestamp.format("DD MMM YYYY")}</td>
       <td class='has-text-right'>${formatCurrency(p.value, 4)}</td>
      `;
    });
}
