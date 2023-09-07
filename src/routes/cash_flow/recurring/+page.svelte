<script lang="ts">
  import {
    ajax,
    helpUrl,
    nextDate,
    sortTrantionSequence,
    type TransactionSequence
  } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  import RecurringCard from "$lib/components/RecurringCard.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";

  let isEmpty = false;
  let transactionSequences: TransactionSequence[] = [];

  onMount(async () => {
    ({ transaction_sequences: transactionSequences } = await ajax("/api/recurring"));

    if (_.isEmpty(transactionSequences)) {
      isEmpty = true;
    }

    transactionSequences = sortTrantionSequence(transactionSequences);
  });
</script>

<div class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <ZeroState item={!isEmpty}>
          <strong>Oops!</strong> You haven't configured any recurring transactions yet. Checkout the
          <a href={helpUrl("recurring")}>docs</a> page to get started.
        </ZeroState>
        {#each transactionSequences as ts}
          <RecurringCard {ts} n={nextDate(ts)} />
        {/each}
      </div>
    </div>
  </div>
</div>
