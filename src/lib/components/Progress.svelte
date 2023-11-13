<script lang="ts">
  import { formatPercentage } from "$lib/utils";
  import { dropRight, floor, range } from "lodash";

  export let small: boolean = false;
  export let progressPercent: number;
  export let showPercent: boolean = true;
  $: times = range(0, floor(progressPercent / 100));
  $: remainder = progressPercent % 100;

  $: if (remainder == 0) {
    times = dropRight(times, 1);
    if (progressPercent == 0) {
      remainder = 0;
    } else {
      remainder = 100;
    }
  }
</script>

<div>
  {#each times as _t}
    <div style="position: relative;" class="mb-1">
      <progress
        class="progress is-success {small ? 'is-extra-small' : 'is-small'}"
        value={100}
        max="100">{100}%</progress
      >
    </div>
  {/each}

  <div style="position: relative;">
    <progress
      class="progress is-success mb-1 {small ? 'is-small' : 'is-large'}"
      value={remainder}
      max="100">{remainder}%</progress
    >
    {#if small && showPercent}
      <span class="has-text-weight-bold">{formatPercentage(progressPercent / 100, 2)}</span>
    {/if}

    {#if !small && showPercent}
      <span
        class="has-text-weight-bold progress-percent {remainder < 10 && 'less-than-10'}"
        style={remainder > 10 ? `right: ${100 - remainder}%;` : `left: ${remainder}%;`}
        >{formatPercentage(progressPercent / 100, 2)}</span
      >
    {/if}
  </div>
</div>
