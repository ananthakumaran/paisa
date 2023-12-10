<script lang="ts">
  import { accountColorStyle } from "$lib/colors";
  import { iconText } from "$lib/icon";
  import {
    formatCurrency,
    postingUrl,
    restName,
    type Posting,
    type Transaction,
    firstName
  } from "$lib/utils";
  import PostingNote from "./PostingNote.svelte";
  import PostingStatus from "./PostingStatus.svelte";

  export let t: Transaction;
  let posting: Posting;
  $: {
    posting = t.postings[0];
  }
</script>

<div class="box p-2 has-background-white">
  <div class="is-flex is-justify-content-space-between is-align-items-baseline">
    <div class="has-text-grey is-size-7 truncate">
      <PostingStatus {posting} />
      <PostingNote {posting} />
      <a class="secondary-link" href={postingUrl(posting)}>{posting.payee}</a>
    </div>
    <div class="has-text-grey min-w-[110px] has-text-right">
      <span class="icon is-small has-text-grey-light">
        <i class="fas fa-calendar" />
      </span>
      {posting.date.format("DD MMM YYYY")}
    </div>
  </div>
  <hr class="my-1" />
  {#each t.postings as posting}
    <div class="my-1 is-flex is-justify-content-space-between">
      <div class="has-text-grey truncate custom-icon" title={posting.account}>
        <span style={accountColorStyle(firstName(posting.account))}
          >{iconText(posting.account)}</span
        >
        {restName(posting.account)}
      </div>
      <div class="has-text-weight-bold is-size-6">
        {formatCurrency(posting.amount)}
      </div>
    </div>
  {/each}
</div>
