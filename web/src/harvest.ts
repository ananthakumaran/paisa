import * as d3 from "d3";
import dayjs from "dayjs";
import _, { round } from "lodash";
import {
  ajax,
  CapitalGain,
  formatCurrency,
  formatFloat,
  restName
} from "./utils";

export default async function () {
  const { capital_gains: capital_gains } = await ajax("/api/harvest");
  renderHarvestables(capital_gains);
}

function renderHarvestables(capital_gains: CapitalGain[]) {
  const id = "#d3-harvestables";
  const root = d3.select(id);

  const card = root
    .selectAll("div.column")
    .data(_.filter(capital_gains, (cg) => cg.harvestable.harvestable_units > 0))
    .enter()
    .append("div")
    .attr("class", "column is-12")
    .append("div")
    .attr("class", "card");

  const header = card.append("header").attr("class", "card-header");

  header
    .append("p")
    .attr("class", "card-header-title")
    .text((cg) => restName(cg.account));

  header
    .append("div")
    .attr("class", "card-header-icon")
    .style("flex-grow", "1")
    .append("div")
    .each(function (cg) {
      const self = d3.select(this);
      const [units, amount, taxableGain] = unitsRequired(cg, 100000);
      self.append("span").html("Redeeming&nbsp;");
      const unitsSpan = self.append("span").text(formatFloat(units));
      self.append("span").html("&nbsp;(amount ₹");
      const amountSpan = self.append("span").text(formatCurrency(amount));
      self.append("span").html(")&nbsp;will result in <b>taxable</b> gain ₹");
      self
        .append("input")
        .attr("class", "input is-small adjustable-input")
        .attr("type", "number")
        .attr("value", round(taxableGain))
        .attr("step", "1000")
        .on("input", (event) => {
          const [units, amount, taxableGain] = unitsRequired(
            cg,
            event.srcElement.value
          );
          unitsSpan.text(formatFloat(units));
          event.srcElement.value = round(taxableGain);
          amountSpan.text(formatCurrency(amount));
        });
    });

  header
    .append("span")
    .attr("class", "card-header-icon")
    .text(
      (cg) =>
        "price as on " +
        dayjs(cg.harvestable.current_unit_date).format("DD MMM YYYY")
    );

  const content = card
    .append("div")
    .attr("class", "card-content")
    .append("div")
    .attr("class", "content")
    .append("div")
    .attr("class", "columns");

  content
    .append("div")
    .attr("class", "column is-4")
    .html((cg) => {
      const h = cg.harvestable;
      return `
<table class="table is-narrow is-fullwidth">
  <tbody>
    <tr>
      <td>Units</td>
      <td class='has-text-right has-text-weight-bold'>${formatFloat(
        h.harvestable_units
      )}</td>
    </tr>
    <tr>
      <td>Grandfathered Unit Price</td>
      <td class='has-text-right has-text-weight-bold'>${formatFloat(
        h.grandfather_unit_price
      )}</td>
    </tr>
    <tr>
      <td>Current Unit Price</td>
      <td class='has-text-right has-text-weight-bold'>${formatFloat(
        h.current_unit_price
      )}</td>
    </tr>
    <tr>
      <td>Unrealized Gain / Loss</td>
      <td class='has-text-right has-text-weight-bold'>${formatCurrency(
        h.unrealized_gain
      )}</td>
    </tr>
    <tr>
      <td>Taxable Unrealized Gain / Loss</td>
      <td class='has-text-right has-text-weight-bold'>${formatCurrency(
        h.taxable_unrealized_gain
      )}</td>
    </tr>
  </tbody>
</table>
`;
    });

  const table = content
    .append("div")
    .attr("class", "column is-8")
    .append("div")
    .attr("class", "table-container")
    .style("overflow-y", "auto")
    .style("max-height", "210px")
    .append("table")
    .attr("class", "table");

  table.append("thead").html(`
<tr>
  <th>Purchase Date</th>
  <th class='has-text-right'>Units</th>
  <th class='has-text-right'>Purchase Price</th>
  <th class='has-text-right'>Purchase Unit Price</th>
  <th class='has-text-right'>Current Price</th>
  <th class='has-text-right'>Unrealized Gain</th>
  <th class='has-text-right'>Taxable Unrealized Gain</th>
</tr>
`);

  table
    .append("tbody")
    .selectAll("tr")
    .data((cg) => {
      return cg.harvestable.harvest_breakdown;
    })
    .enter()
    .append("tr")
    .html((breakdown) => {
      return `
<tr>
  <td style="white-space: nowrap">${dayjs(breakdown.purchase_date).format(
    "DD MMM YYYY"
  )}</td>
  <td class='has-text-right'>${formatFloat(breakdown.units)}</td>
  <td class='has-text-right'>${formatCurrency(breakdown.purchase_price)}</td>
  <td class='has-text-right'>${formatFloat(breakdown.purchase_unit_price)}</td>
  <td class='has-text-right'>${formatCurrency(breakdown.current_price)}</td>
  <td class='has-text-right has-text-weight-bold'>${formatCurrency(
    breakdown.unrealized_gain
  )}</td>
  <td class='has-text-right has-text-weight-bold'>${formatCurrency(
    breakdown.taxable_unrealized_gain
  )}</td>
</tr>
`;
    });
}

function unitsRequired(
  cg: CapitalGain,
  taxableGain: number
): [number, number, number] {
  let gain = 0;
  let amount = 0;
  let units = 0;
  const available = _.clone(cg.harvestable.harvest_breakdown);
  while (taxableGain > gain && available.length > 0) {
    const breakdown = available.shift();
    if (breakdown.taxable_unrealized_gain < taxableGain - gain) {
      gain += breakdown.taxable_unrealized_gain;
      units += breakdown.units;
      amount += breakdown.current_price;
    } else {
      const u =
        ((taxableGain - gain) * breakdown.units) /
        breakdown.taxable_unrealized_gain;
      units += u;
      amount += u * cg.harvestable.current_unit_price;
      gain = taxableGain;
    }
  }
  return [units, amount, gain];
}
