import * as d3 from "d3";
import dayjs from "dayjs";
import _ from "lodash";
import { ajax, Price, formatCurrency } from "./utils";

export default async function () {
  const { prices: prices } = await ajax("/api/price");
  _.each(prices, (p) => (p.timestamp = dayjs(p.date)));
  renderPrices(prices);
}

function renderPrices(prices: Price[]) {
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
