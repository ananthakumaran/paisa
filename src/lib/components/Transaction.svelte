<script lang="ts">
  import { postingUrl, type Transaction } from "$lib/utils";
  import Postings from "$lib/components/Postings.svelte";
  import _ from "lodash";
  import PostingStatus from "$lib/components/PostingStatus.svelte";

  export let compact: boolean = false;
  export let t: Transaction;
  const debits = (t: Transaction) => {
    return _.filter(t.postings, (p) => p.amount < 0);
  };

  const credits = (t: Transaction) => {
    return _.filter(t.postings, (p) => p.amount >= 0);
  };
</script>

<div class="column is-12">
  {#if compact}
    <div class="columns is-flex-wrap-wrap transaction">
      <div class="column is-12 py-0 truncate">
        <div class="description is-size-7">
          <b>{t.date.format("DD MMM YYYY")}</b>
          <span title={t.payee}
            ><PostingStatus posting={t.postings[0]} />
            <a class="secondary-link" href={postingUrl(t.postings[0])}>{t.payee}</a></span
          >
        </div>
      </div>
      <div class="column is-6 py-0">
        <Postings postings={debits(t)} />
      </div>
      <div class="column is-6 py-0">
        <Postings postings={credits(t)} />
      </div>
    </div>
  {:else}
    <div class="columns is-flex-wrap-wrap transaction bordered">
      <div class="column is-3 py-0 truncate">
        <div class="description mt-1 is-size-7">
          <b>{t.date.format("DD MMM YYYY")}</b>
          <span title={t.payee}
            ><PostingStatus posting={t.postings[0]} />
            <a class="secondary-link" href={postingUrl(t.postings[0])}>{t.payee}</a></span
          >
        </div>
      </div>
      <div class="column is-4 py-0">
        <Postings postings={debits(t)} />
      </div>
      <div class="column is-5 py-0">
        <Postings postings={credits(t)} />
      </div>
    </div>
  {/if}
</div>

<style lang="scss">
  @import "bulma/sass/utilities/_all.sass";

  .description {
    display: inline-block;
    white-space: nowrap;
    overflow: hidden;
  }
</style>
