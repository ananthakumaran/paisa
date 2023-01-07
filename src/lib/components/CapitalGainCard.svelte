<script lang="ts">
  import { formatCurrency, formatFloat, type CapitalGain, type FYCapitalGain } from "$lib/utils";
  import { active } from "d3";
  import _ from "lodash";
  import { get } from "svelte/store";
  import CapitalGainDetailCard from "./CapitalGainDetailCard.svelte";
  import Toggleable from "./Toggleable.svelte";

  export let financialYear: string;
  export let capitalGains: CapitalGain[];

  const fyGains: FYCapitalGain[] = _.flatMap(capitalGains, (cg) => cg.fy[financialYear] || []);

  const total = {
    withdrawn: _.sumBy(fyGains, (fy) => fy.sell_price),
    gain: _.sumBy(fyGains, (fy) => fy.gain),
    taxableGain: _.sumBy(fyGains, (fy) => fy.taxable_gain),
    shortTermTax: _.sumBy(fyGains, (fy) => fy.short_term_tax),
    longTermTax: _.sumBy(fyGains, (fy) => fy.long_term_tax)
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
            <table class="table is-narrow is-fullwidth">
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
                          >{formatCurrency(fy.gain)}</td
                        >
                        <td class="has-text-right has-text-weight-bold"
                          >{formatCurrency(fy.taxable_gain)}</td
                        >
                        <td class="has-text-right has-text-weight-bold"
                          >{formatCurrency(fy.short_term_tax)}</td
                        >
                        <td class="has-text-right has-text-weight-bold"
                          >{formatCurrency(fy.long_term_tax)}</td
                        >
                      </tr>
                      <tr slot="content">
                        <td colspan="12" class="p-0">
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
