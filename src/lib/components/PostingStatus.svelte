<script lang="ts">
  import type { Posting } from "$lib/utils";

  export let posting: Posting;
</script>

{#if posting.status == "cleared"}
  <span class="icon is-small">
    <i class="fa-solid fa-check" />
  </span>
{:else if posting.status == "pending"}
  <span class="icon is-small">
    <i class="fa-solid fa-exclamation" />
  </span>
{/if}
