<script lang="ts">
  import Carousel from "svelte-carousel";
  import Transaction from "$lib/components/Transaction.svelte";
  import { intervalText, totalRecurring } from "$lib/transaction_sequence";
  import { formatCurrencyCrude, type TransactionSequence } from "$lib/utils";
  import dayjs from "dayjs";
  import type { Action } from "svelte/action";
  import { renderRecurring } from "$lib/recurring";
  import _ from "lodash";

  export let ts: TransactionSequence;
  export let n: dayjs.Dayjs;
  const now = dayjs();
  const HEIGHT = 50;

  let carousel: Carousel;
  let pageSize = _.min([20, ts.transactions.length]);

  function showPage(pageIndex: number) {
    carousel.goTo(pageSize - 1 - pageIndex);
  }

  const chart: Action<HTMLElement, { ts: TransactionSequence; next: dayjs.Dayjs }> = (
    element,
    props
  ) => {
    renderRecurring(element, props.ts, showPage);
    return {};
  };
</script>

<div class="columns mb-0">
  <div class="column is-12 py-0">
    <div class="is-size-5 has-text-grey">{ts.key}</div>
  </div>
</div>
<div class="columns mb-4">
  <div class="column is-4">
    <div class="box p-2">
      <div
        class="is-flex is-flex-wrap-wrap is-align-items-baseline is-justify-content-space-between"
      >
        <span class="icon-text">
          <span class="icon {n.isBefore(now) ? 'has-text-danger' : 'has-text-success'}">
            <i class="fas {n.isBefore(now) ? 'fa-hourglass-end' : 'fa-hourglass-half'}" />
          </span>
          <span class="has-text-grey">{formatCurrencyCrude(totalRecurring(ts))} due</span><span
            ><b>&nbsp;{n.fromNow()}</b></span
          >
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
      <div use:chart={{ ts: ts, next: n }}>
        <svg height={HEIGHT} width="100%" />
      </div>
      <div class="has-text-grey-light is-size-7">
        <span>{ts.key} started on</span>
        {_.last(ts.transactions).date.format("DD MMM YYYY")}, with a total of
        {ts.transactions.length} transactions so far.
      </div>
    </div>
  </div>

  <div class="column is-8">
    <Carousel bind:this={carousel} infinite={false} initialPageIndex={pageSize - 1}>
      <div
        slot="prev"
        let:showPrevPage
        on:click={showPrevPage}
        class="custom-arrow custom-arrow-prev"
      >
        <i class="fa-solid has-text-grey-light fa-angle-left" />
      </div>
      {#each _.reverse(_.take(ts.transactions, 20)) as t}
        <div class="box px-5 py-3 my-0 has-text-grey">
          <Transaction {t} compact={true} />
        </div>
      {/each}
      <div
        slot="next"
        let:showNextPage
        on:click={showNextPage}
        class="custom-arrow custom-arrow-next"
      >
        <i class="fa-solid has-text-grey-light fa-angle-right" />
      </div>
    </Carousel>
  </div>
</div>
