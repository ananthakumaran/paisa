<script lang="ts">
  import { formatCurrency, formatFloat, type CapitalGain, type YearCapitalGain } from "$lib/utils";
  import _ from "lodash";
  import CapitalGainDetailCard from "./CapitalGainDetailCard.svelte";
  import Toggleable from "./Toggleable.svelte";

  export let calendarYear: string;
  export let capitalGains: CapitalGain[];

  const yearGains: YearCapitalGain[] = _.flatMap(capitalGains, (cg) => cg.year[calendarYear] || []);

  const total = {
    withdrawn: _.sumBy(yearGains, (yg) => yg.sell_price),
    gain: _.sumBy(yearGains, (yg) => yg.tax.gain),
    taxableGain: _.sumBy(yearGains, (yg) => yg.tax.taxable),
    shortTermTax: _.sumBy(yearGains, (yg) => yg.tax.short_term),
    longTermTax: _.sumBy(yearGains, (yg) => yg.tax.long_term),
    slab: _.sumBy(yearGains, (yg) => yg.tax.slab)
  };
</script>

<div class="column is-12">
  <div class="card">
    <header class="card-header">
      <p class="card-header-title">{calendarYear}</p>
    </header>

    <div class="card-content">
      <div class="content">
        <div class="columns">
          <div class="column is-4">
            <table class="table is-narrow is-fullwidth is-hoverable">
              <tbody>
                <tr>
                  <td>Withdrawn</td>
                  <td class="has-text-right has-text-weight-bold"
                    >{formatCurrency(total["withdrawn"])}</td
                  >
                </tr>
                <tr>
                  <td>Gain</td>
                  <td class="has-text-right has-text-weight-bold"
                    >{formatCurrency(total["gain"])}</td
                  >
                </tr>
                <tr>
                  <td>Taxable Gain</td>
                  <td class="has-text-right has-text-weight-bold"
                    >{formatCurrency(total["taxableGain"])}</td
                  >
                </tr>
                <tr>
                  <td>Short Term Tax</td>
                  <td class="has-text-right has-text-weight-bold"
                    >{formatCurrency(total["shortTermTax"])}</td
                  >
                </tr>
                <tr>
                  <td>Long Term Tax</td>
                  <td class="has-text-right has-text-weight-bold"
                    >{formatCurrency(total["longTermTax"])}</td
                  >
                </tr>
                <tr>
                  <td>Taxable at Slab Rate</td>
                  <td class="has-text-right has-text-weight-bold"
                    >{formatCurrency(total["slab"])}</td
                  >
                </tr>
              </tbody>
            </table>
          </div>
          <div class="column is-8 overflow-x-auto">
            <table class="table is-narrow is-fullwidth is-hoverable">
              <thead>
                <tr>
                  <th />
                  <th>Account</th>
                  <th>Tax Category</th>
                  <th class="has-text-right">Sold Units</th>
                  <th class="has-text-right">Purchase Price</th>
                  <th class="has-text-right">Average Purchase Unit Price</th>
                  <th class="has-text-right">Sell Price</th>
                  <th class="has-text-right">Average Sell Unit Price</th>
                  <th class="has-text-right">Gain</th>
                  <th class="has-text-right">Taxable Gain</th>
                  <th class="has-text-right">Short Term Tax</th>
                  <th class="has-text-right">Long Term Tax</th>
                  <th class="has-text-right">Taxable at Slat Rate</th>
                </tr>
              </thead>
              <tbody>
                {#each capitalGains as cg}
                  {#if cg.year[calendarYear]}
                    {@const yg = cg.year[calendarYear]}
                    <Toggleable>
                      <tr
                        class={active ? "is-active has-background-white-ter" : ""}
                        style="cursor: pointer;"
                        slot="toggle"
                        let:active
                        let:onclick
                        on:click={(e) => onclick(e)}
                      >
                        <td>
                          <span class="icon has-text-link">
                            <i
                              class="fas {active ? 'fa-chevron-up' : 'fa-chevron-down'}"
                              aria-hidden="true"
                            />
                          </span>
                        </td>
                        <td>{cg.account}</td>
                        <td>{cg.tax_category}</td>
                        <td class="has-text-right">{formatFloat(yg.units)}</td>
                        <td class="has-text-right">{formatCurrency(yg.purchase_price)}</td>
                        <td class="has-text-right"
                          >{formatCurrency(yg.purchase_price / yg.units, 4)}</td
                        >
                        <td class="has-text-right">{formatCurrency(yg.sell_price)}</td>
                        <td class="has-text-right">{formatCurrency(yg.sell_price / yg.units, 4)}</td
                        >
                        <td class="has-text-right has-text-weight-bold"
                          >{formatCurrency(yg.tax.gain)}</td
                        >
                        <td class="has-text-right has-text-weight-bold"
                          >{formatCurrency(yg.tax.taxable)}</td
                        >
                        <td class="has-text-right has-text-weight-bold"
                          >{formatCurrency(yg.tax.short_term)}</td
                        >
                        <td class="has-text-right has-text-weight-bold"
                          >{formatCurrency(yg.tax.long_term)}</td
                        >
                        <td class="has-text-right has-text-weight-bold"
                          >{formatCurrency(yg.tax.slab)}</td
                        >
                      </tr>
                      <tr slot="content">
                        <td colspan="13" class="p-0">
                          <CapitalGainDetailCard yearCapitalGain={yg} />
                        </td>
                      </tr>
                    </Toggleable>
                  {/if}
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
