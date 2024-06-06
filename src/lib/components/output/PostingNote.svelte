Here is the Turkish translation:

<script lang="ts">
  import { formatTextAsHtml, type Posting } from "$lib/utils";
  export let posting: Posting;
</script>

{#if posting.note != ""}
  <span class="icon is-small" data-tippy-content={formatTextAsHtml(posting.note)}>
    <i class="fa-regular fa-comment-dots" />
  </span>
{/if}
