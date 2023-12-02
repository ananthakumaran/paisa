<script lang="ts">
  import { iconText } from "$lib/icon";
  import {
    type AssetBreakdown,
    depth,
    lastName,
    isZero,
    formatCurrency,
    formatFloat,
    formatPercentage
  } from "$lib/utils";
  import _ from "lodash";

  export let breakdowns: Record<string, AssetBreakdown>;
  export let indent = true;

  function calculateChangeClass(gain: number) {
    let changeClass = "";
    if (gain > 0) {
      changeClass = "has-text-success";
    } else if (gain < 0) {
      changeClass = "has-text-danger";
    }
    return changeClass;
  }
</script>

<div class="box overflow-x-auto max-h-screen max-w-fit pt-0">
  <table class="table is-narrow is-hoverable is-light-border has-sticky-header">
    <thead>
      <tr>
        <th class="py-2">Account</th>
        <th class="py-2 has-text-right">Investment Amount</th>
        <th class="py-2 has-text-right">Withdrawal Amount</th>
        <th class="py-2 has-text-right">Balance Units</th>
        <th class="py-2 has-text-right">Market Value</th>
        <th class="py-2 has-text-right">Change</th>
        <th class="py-2 has-text-right">XIRR</th>
        <th class="py-2 has-text-right">Absolute Return</th>
      </tr>
    </thead>
    <tbody class="has-text-grey-dark">
      {#each Object.values(breakdowns) as b}
        {@const indentWidth = indent ? _.repeat("&emsp;&emsp;", depth(b.group) - 1) : ""}
        {@const gain = b.gainAmount}
        {@const changeClass = calculateChangeClass(gain)}
        <tr>
          <td class="whitespace-nowrap has-text-left" style="max-width: 200px; overflow: hidden;"
            >{@html indentWidth}<span class="has-text-grey custom-icon">{iconText(b.group)}</span>
            <a href="/assets/gain/{b.group}">{indent ? lastName(b.group) : b.group}</a></td
          >
          <td class="has-text-right"
            >{!isZero(b.investmentAmount) ? formatCurrency(b.investmentAmount) : ""}</td
          >
          <td class="has-text-right"
            >{!isZero(b.withdrawalAmount) ? formatCurrency(b.withdrawalAmount) : ""}</td
          >
          <td class="has-text-right">{b.balanceUnits > 0 ? formatFloat(b.balanceUnits, 4) : ""}</td>
          <td class="has-text-right"
            >{!isZero(b.marketAmount) ? formatCurrency(b.marketAmount) : ""}</td
          >
          <td class="{changeClass} has-text-right"
            >{!isZero(b.investmentAmount) && !isZero(gain) ? formatCurrency(gain) : ""}</td
          >
          <td class="{changeClass} has-text-right">{!isZero(b.xirr) ? formatFloat(b.xirr) : ""}</td>
          <td class="{changeClass} has-text-right"
            >{!isZero(b.absoluteReturn) ? formatPercentage(b.absoluteReturn, 2) : ""}</td
          >
        </tr>
      {/each}
    </tbody>
  </table>
</div>
