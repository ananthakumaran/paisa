<script lang="ts">
  import { ajax, helpUrl, type TransactionSequence } from "$lib/utils";
  import dayjs from "dayjs";
  import _ from "lodash";
  import { onMount } from "svelte";
  import RecurringCard from "$lib/components/RecurringCard.svelte";

  let isEmpty = false;
  let transactionSequences: TransactionSequence[] = [];

  function nextDate(ts: TransactionSequence) {
    const lastTransaction = ts.transactions[0];
    if (ts.interval >= 28 && ts.interval <= 33) {
      return lastTransaction.date.add(1, "month");
    }

    if (ts.interval >= 360 && ts.interval <= 370) {
      return lastTransaction.date.add(1, "year");
    }

    return lastTransaction.date.add(ts.interval, "day");
  }

  onMount(async () => {
    ({ transaction_sequences: transactionSequences } = await ajax("/api/recurring"));

    if (_.isEmpty(transactionSequences)) {
      isEmpty = true;
    }

    transactionSequences = _.chain(transactionSequences)
      .sortBy((ts) => {
        return Math.abs(nextDate(ts).diff(dayjs()));
      })
      .value();
  });
</script>

<section class="section" class:is-hidden={!isEmpty}>
  <div class="container is-fluid">
    <div class="columns is-centered">
      <div class="column is-4 has-text-centered">
        <article class="message is-info">
          <div class="message-body">
            <strong>Oops!</strong> You haven't configured any recurring transactions yet. Checkout
            the <a href={helpUrl("recurring")}>docs</a> page to get started.
          </div>
        </article>
      </div>
    </div>
  </div>
</section>

<div class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        {#each transactionSequences as ts}
          <RecurringCard {ts} n={nextDate(ts)} />
        {/each}
      </div>
    </div>
  </div>
</div>
