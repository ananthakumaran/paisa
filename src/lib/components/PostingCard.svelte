<script lang="ts">
  import { iconify } from "$lib/icon";
  import {
    formatCurrency,
    formatFloat,
    postingUrl,
    type Posting,
    firstName,
    restName
  } from "$lib/utils";
  import PostingStatus from "./PostingStatus.svelte";

  export let posting: Posting;
  export let color: string;
  export let icon: boolean = false;
</script>

<div class="box p-2 my-2 has-background-white" style="border-left: 2px solid {color}">
  <div class="is-flex is-justify-content-space-between">
    <div class="has-text-grey is-size-7 truncate">
      <PostingStatus {posting} />
      <a class="secondary-link" href={postingUrl(posting)}>{posting.payee}</a>
    </div>
    <div class="has-text-grey min-w-[110px] has-text-right">
      <span class="icon is-small has-text-grey-light">
        <i class="fas fa-calendar" />
      </span>
      {posting.date.format("DD MMM YYYY")}
    </div>
  </div>
  <hr class="m-1" />
  <div class="is-flex is-justify-content-space-between">
    <div class="has-text-grey truncate">
      {#if icon}
        {iconify(restName(posting.account), { group: firstName(posting.account) })}
      {:else}
        {posting.account}
      {/if}
    </div>
    <div class="is-flex is-align-items-baseline">
      <div
        class="has-text-grey mr-1 truncate is-size-7"
        class:is-hidden={posting.quantity == posting.amount}
      >
        {formatFloat(posting.quantity, 4)} @ {formatFloat(posting.amount / posting.quantity, 3)}
      </div>
      <div class="has-text-weight-bold is-size-6">{formatCurrency(posting.amount)}</div>
    </div>
  </div>
</div>
