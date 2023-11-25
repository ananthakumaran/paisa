<script lang="ts">
  import { iconText } from "$lib/icon";
  import {
    ajax,
    depth,
    formatCurrency,
    formatFloat,
    lastName,
    type AssetBreakdown,
    isZero,
    formatPercentage
  } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  let breakdowns: AssetBreakdown[] = [];

  function calculateChangeClass(gain: number) {
    let changeClass = "";
    if (gain > 0) {
      changeClass = "has-text-success";
    } else if (gain < 0) {
      changeClass = "has-text-danger";
    }
    return changeClass;
  }

  onMount(async () => {
    ({ asset_breakdowns: breakdowns } = await ajax("/api/assets/balance"));
  });
</script>

<section class="section tab-holding">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box overflow-x-auto">
          <table class="table is-narrow is-fullwidth is-hoverable is-light-border">
            <thead>
              <tr>
                <th>Account</th>
                <th class="has-text-right">Investment Amount</th>
                <th class="has-text-right">Withdrawal Amount</th>
                <th class="has-text-right">Balance Units</th>
                <th class="has-text-right">Market Value</th>
                <th class="has-text-right">Change</th>
                <th class="has-text-right">XIRR</th>
                <th class="has-text-right">Absolute Return</th>
              </tr>
            </thead>
            <tbody class="has-text-grey-dark">
              {#each Object.values(breakdowns) as b}
                {@const indent = _.repeat("&emsp;&emsp;", depth(b.group) - 1)}
                {@const gain = b.gainAmount}
                {@const changeClass = calculateChangeClass(gain)}
                <tr>
                  <td class="whitespace-nowrap" style="max-width: 200px; overflow: hidden;"
                    >{@html indent}<span class="has-text-grey custom-icon">{iconText(b.group)}</span
                    >
                    <a href="/assets/gain/{b.group}">{lastName(b.group)}</a></td
                  >
                  <td class="has-text-right"
                    >{!isZero(b.investmentAmount) ? formatCurrency(b.investmentAmount) : ""}</td
                  >
                  <td class="has-text-right"
                    >{!isZero(b.withdrawalAmount) ? formatCurrency(b.withdrawalAmount) : ""}</td
                  >
                  <td class="has-text-right"
                    >{b.balanceUnits > 0 ? formatFloat(b.balanceUnits, 4) : ""}</td
                  >
                  <td class="has-text-right"
                    >{!isZero(b.marketAmount) ? formatCurrency(b.marketAmount) : ""}</td
                  >
                  <td class="{changeClass} has-text-right"
                    >{!isZero(b.investmentAmount) && !isZero(gain) ? formatCurrency(gain) : ""}</td
                  >
                  <td class="{changeClass} has-text-right"
                    >{!isZero(b.xirr) ? formatFloat(b.xirr) : ""}</td
                  >
                  <td class="{changeClass} has-text-right"
                    >{!isZero(b.absoluteReturn) ? formatPercentage(b.absoluteReturn, 2) : ""}</td
                  >
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</section>
