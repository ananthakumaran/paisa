<script lang="ts">
  import Carousel from "svelte-carousel";
  import Transaction from "$lib/components/Transaction.svelte";
  import {
    formatCurrencyCrude,
    intervalText,
    totalRecurring,
    type TransactionSequence
  } from "$lib/utils";
  import dayjs from "dayjs";
  import type { Action } from "svelte/action";
  import { renderRecurring } from "$lib/recurring";
  import _ from "lodash";

  export let ts: TransactionSequence;
  export let n: dayjs.Dayjs;
  const now = dayjs();

  let carousel: Carousel;
  let pageSize = _.min([20, ts.transactions.length]);

  function showPage(pageIndex: number) {
    carousel.goTo(pageSize - 1 - pageIndex);
  }

  const chart: Action<HTMLElement, { ts: TransactionSequence; next: dayjs.Dayjs }> = (
    element,
    props
  ) => {
    renderRecurring(element, props.ts, props.next, showPage);
    return {};
  };
</script>

<div class="columns">
  <div class="column is-4">
    <div class="box p-2">
      <div
        class="is-flex is-flex-wrap-wrap is-align-items-baseline is-justify-content-space-between"
      >
        <span
          class="icon-text tag is-medium is-light {n.isBefore(now) ? 'is-danger' : 'is-success'}"
        >
          <span class="icon">
            <i class="fas {n.isBefore(now) ? 'fa-hourglass-end' : 'fa-hourglass-half'}" />
          </span>
          <span>{formatCurrencyCrude(totalRecurring(ts))} due {n.fromNow()}</span>
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
        <svg height="50" width="100%" />
      </div>
      <div class="">
        <span><b>{ts.key.tagRecurring}</b> started on</span>
        <b>{_.last(ts.transactions).date.format("DD MMM YYYY")}</b>, with a total of
        <b>{ts.transactions.length}</b> transactions so far.
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
        <div class="box px-5 py-3 my-0">
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
