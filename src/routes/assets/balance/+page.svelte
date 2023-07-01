<script lang="ts">
  import { iconText } from "$lib/icon";
  import {
    ajax,
    depth,
    formatCurrency,
    formatFloat,
    lastName,
    type AssetBreakdown
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
      <div class="column is-12 box">
        <table class="table is-narrow is-fullwidth is-hoverable">
          <thead>
            <tr>
              <th>Account</th>
              <th class="has-text-right">Investment Amount</th>
              <th class="has-text-right">Withdrawal Amount</th>
              <th class="has-text-right">Balance Units</th>
              <th class="has-text-right">Market Value</th>
              <th class="has-text-right">Change</th>
              <th class="has-text-right">XIRR</th>
            </tr>
          </thead>
          <tbody>
            {#each Object.values(breakdowns) as b}
              {@const indent = _.repeat("&emsp;&emsp;", depth(b.group) - 1)}
              {@const gain = b.market_amount + b.withdrawal_amount - b.investment_amount}
              {@const changeClass = calculateChangeClass(gain)}
              <tr>
                <td style="max-width: 200px; overflow: hidden;"
                  >{@html indent}<span class="has-text-grey">{iconText(b.group)}</span>
                  {lastName(b.group)}</td
                >
                <td class="has-text-right"
                  >{b.investment_amount != 0 ? formatCurrency(b.investment_amount) : ""}</td
                >
                <td class="has-text-right"
                  >{b.withdrawal_amount != 0 ? formatCurrency(b.withdrawal_amount) : ""}</td
                >
                <td class="has-text-right"
                  >{b.balance_units > 0 ? formatFloat(b.balance_units, 4) : ""}</td
                >
                <td class="has-text-right"
                  >{b.market_amount != 0 ? formatCurrency(b.market_amount) : ""}</td
                >
                <td class="{changeClass} has-text-right"
                  >{b.investment_amount != 0 && gain != 0 ? formatCurrency(gain) : ""}</td
                >
                <td class="{changeClass} has-text-right"
                  >{b.xirr > 0.0001 || b.xirr < -0.0001 ? formatFloat(b.xirr) : ""}</td
                >
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  </div>
</section>
