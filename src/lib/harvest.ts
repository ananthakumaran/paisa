import * as d3 from "d3";
import dayjs from "dayjs";
import _, { round } from "lodash";
import COLORS from "./colors";
import { formatCurrency, formatFloat, restName, tooltip, type Harvestable } from "./utils";

export function renderHarvestables(harvestables: Harvestable[]) {
  const id = "#d3-harvestables";
  const root = d3.select(id);

  const card = root
    .selectAll("div.column")
    .data(_.filter(harvestables, (harvestable) => harvestable.harvestable_units > 0))
    .enter()
    .append("div")
    .attr("class", "column is-12")
    .append("div")
    .attr("class", "card");

  const header = card.append("header").attr("class", "card-header");

  header
    .append("p")
    .attr("class", "card-header-title")
    .text((h) => restName(h.account));

  header
    .append("div")
    .attr("class", "card-header-icon")
    .style("flex-grow", "1")
    .style("cursor", "auto")
    .append("div")
    .each(function (h) {
      const self = d3.select(this);
      const [units, amount, taxableGain] = unitsRequiredFromGain(h, 100000);
      self.append("span").html("If you redeem&nbsp;");
      const unitsSpan = self.append("span").text(formatFloat(units));
      self.append("span").html("&nbsp;units you will get ₹");
      const amountInput = self
        .append("input")
        .attr("class", "input is-small adjustable-input")
        .attr("type", "number")
        .attr("value", round(amount))
        .attr("step", "1000")
        .on("input", (event) => {
          const [units, amount, taxableGain] = unitsRequiredFromAmount(
            h,
            parseInt(event.srcElement.value)
          );

          unitsSpan.text(formatFloat(units));
          (taxableGainInput.node() as HTMLInputElement).value = round(taxableGain).toString();
          event.srcElement.value = round(amount);
        });
      self.append("span").html("&nbsp; and your <b>taxable</b> gain would be ₹");
      const taxableGainInput = self
        .append("input")
        .attr("class", "input is-small adjustable-input")
        .attr("type", "number")
        .attr("value", round(taxableGain))
        .attr("step", "1000")
        .on("input", (event) => {
          const [units, amount, taxableGain] = unitsRequiredFromGain(
            h,
            parseInt(event.srcElement.value)
          );
          unitsSpan.text(formatFloat(units));
          event.srcElement.value = round(taxableGain);
          (amountInput.node() as HTMLInputElement).value = round(amount).toString();
        });
    });

  header
    .append("span")
    .attr("class", "card-header-icon")
    .text(
      (harvestable) => "price as on " + dayjs(harvestable.current_unit_date).format("DD MMM YYYY")
    );

  const content = card
    .append("div")
    .attr("class", "card-content")
    .append("div")
    .attr("class", "content")
    .append("div")
    .attr("class", "columns");

  const summary = content.append("div").attr("class", "column is-4");

  summary.append("div").each(renderSingleBar);

  summary.append("div").html((h) => {
    return `
<table class="table is-narrow is-fullwidth">
  <tbody>
    <tr>
      <td>Balance Units</td>
      <td class='has-text-right has-text-weight-bold'>${formatFloat(h.total_units)}</td>
    </tr>
    <tr>
      <td>Harvestable Units</td>
      <td class='has-text-right has-text-weight-bold has-text-success'>${formatFloat(
        h.harvestable_units
      )}</td>
    </tr>
    <tr>
      <td>Tax Category</td>
      <td class='has-text-right is-uppercase'>${h.tax_category}</td>
    </tr>
    <tr>
      <td>Current Unit Price</td>
      <td class='has-text-right has-text-weight-bold'>${formatFloat(h.current_unit_price)}</td>
    </tr>
    <tr>
      <td>Unrealized Gain / Loss</td>
      <td class='has-text-right has-text-weight-bold'>${formatCurrency(h.unrealized_gain)}</td>
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
    .style("max-height", "245px")
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
    .data((harvestable) => {
      return harvestable.harvest_breakdown;
    })
    .enter()
    .append("tr")
    .html((breakdown) => {
      return `
<tr>
  <td style="white-space: nowrap">${dayjs(breakdown.purchase_date).format("DD MMM YYYY")}</td>
  <td class='has-text-right'>${formatFloat(breakdown.units)}</td>
  <td class='has-text-right'>${formatCurrency(breakdown.purchase_price)}</td>
  <td class='has-text-right'>${formatFloat(breakdown.purchase_unit_price)}</td>
  <td class='has-text-right'>${formatCurrency(breakdown.current_price)}</td>
  <td class='has-text-right has-text-weight-bold'>${formatCurrency(breakdown.unrealized_gain)}</td>
  <td class='has-text-right has-text-weight-bold'>${formatCurrency(
    breakdown.taxable_unrealized_gain
  )}</td>
</tr>
`;
    });
}

function unitsRequiredFromGain(
  harvestable: Harvestable,
  taxableGain: number
): [number, number, number] {
  let gain = 0;
  let amount = 0;
  let units = 0;
  const available = _.clone(harvestable.harvest_breakdown);
  while (taxableGain > gain && available.length > 0) {
    const breakdown = available.shift();
    if (breakdown.taxable_unrealized_gain < taxableGain - gain) {
      gain += breakdown.taxable_unrealized_gain;
      units += breakdown.units;
      amount += breakdown.current_price;
    } else {
      const u = ((taxableGain - gain) * breakdown.units) / breakdown.taxable_unrealized_gain;
      units += u;
      amount += u * harvestable.current_unit_price;
      gain = taxableGain;
    }
  }
  return [units, amount, gain];
}

function unitsRequiredFromAmount(
  harvestable: Harvestable,
  expectedAmount: number
): [number, number, number] {
  let gain = 0;
  let amount = 0;
  let units = 0;
  const available = _.clone(harvestable.harvest_breakdown);
  while (expectedAmount > amount && available.length > 0) {
    const breakdown = available.shift();
    if (breakdown.current_price < expectedAmount - amount) {
      gain += breakdown.taxable_unrealized_gain;
      units += breakdown.units;
      amount += breakdown.current_price;
    } else {
      const u = (expectedAmount - amount) / harvestable.current_unit_price;
      units += u;
      amount = expectedAmount;
      gain += u * (breakdown.taxable_unrealized_gain / breakdown.units);
    }
  }
  return [units, amount, gain];
}

function renderSingleBar(harvestable: Harvestable) {
  const selection = d3.select(this);
  const svg = selection.append("svg");

  const height = 20;
  const margin = { top: 20, right: 0, bottom: 20, left: 0 },
    width = selection.node().clientWidth - margin.left - margin.right,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  svg
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom);

  const x = d3.scaleLinear().range([0, width]).domain([0, harvestable.total_units]);

  const non_harvestable_units = harvestable.total_units - harvestable.harvestable_units;

  g.attr("data-tippy-content", () => {
    return tooltip([
      [
        ["Type", "has-text-weight-bold"],
        ["Units", "has-text-weight-bold has-text-right"],
        ["Percentage", "has-text-weight-bold has-text-right"]
      ],
      [
        "Harvestable",
        [formatFloat(harvestable.harvestable_units), "has-text-right"],
        [
          formatFloat((harvestable.harvestable_units / harvestable.total_units) * 100),
          "has-text-right"
        ]
      ],
      [
        "Non Harvestable",
        [formatFloat(non_harvestable_units), "has-text-right"],
        [formatFloat((non_harvestable_units / harvestable.total_units) * 100), "has-text-right"]
      ]
    ]);
  });

  g.selectAll("rect")
    .data([
      { start: 0, end: harvestable.harvestable_units, color: COLORS.gainText },
      {
        start: harvestable.harvestable_units,
        end: harvestable.total_units,
        color: COLORS.tertiary
      }
    ])
    .join("rect")
    .attr("fill", (d) => d.color)
    .attr("x", (d) => x(d.start))
    .attr("width", (d) => x(d.end) - x(d.start))
    .attr("y", 0)
    .attr("height", height);
}
