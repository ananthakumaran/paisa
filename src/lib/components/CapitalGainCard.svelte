<script lang="ts">
  import { formatCurrency, formatFloat, type CapitalGain, type FYCapitalGain } from "$lib/utils";
  import _ from "lodash";
  import CapitalGainDetailCard from "./CapitalGainDetailCard.svelte";
  import Toggleable from "./Toggleable.svelte";

  export let financialYear: string;
  export let capitalGains: CapitalGain[];

  const fyGains: FYCapitalGain[] = _.flatMap(capitalGains, (cg) => cg.fy[financialYear] || []);

  const total = {
    withdrawn: _.sumBy(fyGains, (fy) => fy.sell_price),
    gain: _.sumBy(fyGains, (fy) => fy.tax.gain),
    taxableGain: _.sumBy(fyGains, (fy) => fy.tax.taxable),
    shortTermTax: _.sumBy(fyGains, (fy) => fy.tax.short_term),
    longTermTax: _.sumBy(fyGains, (fy) => fy.tax.long_term),
    slab: _.sumBy(fyGains, (fy) => fy.tax.slab)
  };
</script>

<div class="column is-12">
  <div class="card">
    <header class="card-header">
      <p class="card-header-title">{financialYear}</p>
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
          <div class="column is-8">
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
                  {#if cg.fy[financialYear]}
                    {@const fy = cg.fy[financialYear]}
                    <Toggleable>
                      <tr
                        class={active ? "is-active" : ""}
                        style="cursor: pointer;"
                        slot="toggle"
                        let:active
                        let:onclick
                        on:click={(e) => onclick(e)}
                      >
                        <td>
                          <span class="icon">
                            <i
                              class="fas {active ? 'fa-chevron-down' : 'fa-chevron-right'}"
                              aria-hidden="true"
                            />
                          </span>
                        </td>
                        <td>{cg.account}</td>
                        <td>{cg.tax_category}</td>
                        <td class="has-text-right">{formatFloat(fy.units)}</td>
                        <td class="has-text-right">{formatCurrency(fy.purchase_price)}</td>
                        <td class="has-text-right"
                          >{formatCurrency(fy.purchase_price / fy.units, 4)}</td
                        >
                        <td class="has-text-right">{formatCurrency(fy.sell_price)}</td>
                        <td class="has-text-right">{formatCurrency(fy.sell_price / fy.units, 4)}</td
                        >
                        <td class="has-text-right has-text-weight-bold"
                          >{formatCurrency(fy.tax.gain)}</td
                        >
                        <td class="has-text-right has-text-weight-bold"
                          >{formatCurrency(fy.tax.taxable)}</td
                        >
                        <td class="has-text-right has-text-weight-bold"
                          >{formatCurrency(fy.tax.short_term)}</td
                        >
                        <td class="has-text-right has-text-weight-bold"
                          >{formatCurrency(fy.tax.long_term)}</td
                        >
                        <td class="has-text-right has-text-weight-bold"
                          >{formatCurrency(fy.tax.slab)}</td
                        >
                      </tr>
                      <tr slot="content">
                        <td colspan="13" class="p-0">
                          <CapitalGainDetailCard fyCapitalGain={fy} />
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

<style>
  .table tr.is-active {
    background-color: #eee !important;
  }
</style>
