<script lang="ts">
  import _ from "lodash";

  export let label: string;
  export let value: number;
  export let allowed: { min: number; max: number };
</script>

{#if allowed.max > 1}
  <div>
    <div class="mb-1">{label}</div>
    <input
      type="range"
      bind:value
      min={allowed.min}
      max={allowed.max}
      class="du-range du-range-xs"
      step={1}
    />
    <div class="is-size-7 is-flex is-justify-content-space-between">
      {#each _.range(allowed.min, allowed.max + 1) as i}
        <span>{i}</span>
      {/each}
    </div>
  </div>
{/if}
