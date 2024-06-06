İşte svelte bileşeni: <script lang="ts">
  import { formatPercentage } from "$lib/utils";

  export let value: number;
  $: icon = value > 0 ? "fa-arrow-up rotate-45" : "fa-arrow-down -rotate-45";
  $: color = value > 0 ? "has-text-success" : "has-text-danger";
</script>

{#if value === null}
  <span></span>
{:else if value == 0}
  <span class="has-text-grey-light">{formatPercentage(value, 2)}</span>
{:else}
  <span class="{color} whitespace-nowrap">
    <span class="icon is-small">
      <i class="fas {icon}" />
    </span>
    {formatPercentage(value, 2)}
  </span>
{/if}