<script lang="ts">
  import { iconText } from "$lib/icon";
  import {
    ajax,
    depth,
    formatCurrency,
    formatFloat,
    lastName,
    type LiabilityBreakdown
  } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  let breakdowns: LiabilityBreakdown[] = [];
  let isEmpty = false;

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
    ({ liability_breakdowns: breakdowns } = await ajax("/api/liabilities/balance"));

    if (_.isEmpty(breakdowns)) {
      isEmpty = true;
    }
  });
</script>

<section class="section" class:is-hidden={!isEmpty}>
  <div class="container is-fluid">
    <div class="columns is-centered">
      <div class="column is-4 has-text-centered">
        <article class="message">
          <div class="message-body">
            <strong>Hurray!</strong> You have no liabilities.
          </div>
        </article>
      </div>
    </div>
  </div>
</section>

<section class="section" class:is-hidden={isEmpty}>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box overflow-x-auto">
          <table class="table is-narrow is-fullwidth is-hoverable">
            <thead>
              <tr>
                <th>Account</th>
                <th class="has-text-right">Drawn Amount</th>
                <th class="has-text-right">Repaid Amount</th>
                <th class="has-text-right">Balance Amount</th>
                <th class="has-text-right">Interest</th>
                <th class="has-text-right">APR</th>
              </tr>
            </thead>
            <tbody class="has-text-grey-dark">
              {#each Object.values(breakdowns) as b}
                {@const indent = _.repeat("&emsp;&emsp;", depth(b.group) - 1)}
                {@const changeClass = calculateChangeClass(-b.interest_amount)}
                <tr>
                  <td class="whitespace-nowrap" style="max-width: 200px; overflow: hidden;"
                    >{@html indent}<span class="has-text-grey custom-icon">{iconText(b.group)}</span
                    >
                    {lastName(b.group)}</td
                  >
                  <td class="has-text-right"
                    >{b.drawn_amount != 0 ? formatCurrency(b.drawn_amount) : ""}</td
                  >
                  <td class="has-text-right"
                    >{b.repaid_amount != 0 ? formatCurrency(b.repaid_amount) : ""}</td
                  >
                  <td class="has-text-right"
                    >{b.balance_amount != 0 ? formatCurrency(b.balance_amount) : ""}</td
                  >
                  <td class="has-text-right"
                    >{b.interest_amount != 0 ? formatCurrency(b.interest_amount) : ""}</td
                  >
                  <td class="{changeClass} has-text-right"
                    >{b.apr > 0.0001 || b.apr < -0.0001 ? formatFloat(b.apr) : ""}</td
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
