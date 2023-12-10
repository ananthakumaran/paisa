<script lang="ts">
  import { formatTextAsHtml, type Transaction } from "$lib/utils";
  export let transaction: Transaction;
</script>

{#if transaction.note != ""}
  <span class="icon is-small" data-tippy-content={formatTextAsHtml(transaction.note)}>
    <i class="fa-regular fa-comment-dots" />
  </span>
{/if}
