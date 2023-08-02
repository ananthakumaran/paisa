<script lang="ts">
  import Carousel from "svelte-carousel";
  import Transaction from "$lib/components/Transaction.svelte";
  import { ajax, helpUrl, type TransactionSequence } from "$lib/utils";
  import dayjs from "dayjs";
  import _ from "lodash";
  import { onMount } from "svelte";

  let isEmpty = false;
  const now = dayjs();
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

  function intervalText(ts: TransactionSequence) {
    if (ts.interval >= 7 && ts.interval <= 8) {
      return "weekly";
    }

    if (ts.interval >= 14 && ts.interval <= 16) {
      return "bi-weekly";
    }

    if (ts.interval >= 28 && ts.interval <= 33) {
      return "monthly";
    }

    if (ts.interval >= 87 && ts.interval <= 100) {
      return "quarterly";
    }

    if (ts.interval >= 175 && ts.interval <= 190) {
      return "half-yearly";
    }

    if (ts.interval >= 350 && ts.interval <= 395) {
      return "yearly";
    }

    return `every ${ts.interval} days`;
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
          {@const n = nextDate(ts)}
          <div class="columns">
            <div class="column is-3">
              <div class="box p-2">
                <div
                  class="is-flex is-flex-wrap-wrap is-align-items-baseline is-justify-content-space-between"
                >
                  <span
                    class="icon-text tag is-medium is-light {n.isBefore(now)
                      ? 'is-danger'
                      : 'is-success'}"
                  >
                    <span class="icon">
                      <i class="fas {n.isBefore(now) ? 'fa-hourglass-end' : 'fa-hourglass-half'}" />
                    </span>
                    <span>due {n.fromNow()}</span>
                  </span>
                  <div class="has-text-grey">
                    <span class="tag is-light">{intervalText(ts)}</span>
                    <span class="icon has-text-grey-light">
                      <i class="fas fa-calendar" />
                    </span>
                    {n.format("DD MMM YYYY")}
                  </div>
                </div>
                <hr class="m-1" />
                <div>
                  <span>Started on</span>
                  <b>{_.last(ts.transactions).date.format("DD MMM YYYY")}</b>, with a total of
                  <b>{ts.transactions.length}</b> transactions so far.
                </div>
              </div>
            </div>

            <div class="column is-9">
              <Carousel let:showPrevPage let:showNextPage infinite={false}>
                <div slot="prev" on:click={showPrevPage} class="custom-arrow custom-arrow-prev">
                  <i class="fa-solid has-text-grey-light fa-angle-left" />
                </div>
                {#each _.take(ts.transactions, 20) as t}
                  <div class="box px-5 py-3 my-0">
                    <Transaction {t} compact={true} />
                  </div>
                {/each}
                <div slot="next" on:click={showNextPage} class="custom-arrow custom-arrow-next">
                  <i class="fa-solid has-text-grey-light fa-angle-right" />
                </div>
              </Carousel>
            </div>
          </div>
        {/each}
      </div>
    </div>
  </div>
</div>

<style lang="scss">
  @import "bulma/sass/utilities/_all.sass";

  .custom-arrow-prev {
    border-radius: $radius 0 0 $radius;
    left: 0;
  }

  .custom-arrow-next {
    border-radius: 0 $radius $radius 0;
    right: 0;
  }

  .custom-arrow {
    width: 25px;
    position: absolute;
    top: 0;
    bottom: 0;
    z-index: 1;
    transition: opacity 150ms ease;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    -webkit-tap-highlight-color: transparent;
    background-color: $white-bis;

    &:hover {
      background-color: $white-ter;
    }
  }
</style>
