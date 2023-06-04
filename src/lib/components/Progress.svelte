<script lang="ts">
  import { formatPercentage } from "$lib/utils";
  import { dropRight, floor, range } from "lodash";

  export let progressPercent: number;
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

<div class="container is-fluid">
  {#each times as _t}
    <div style="position: relative;" class="mb-1">
      <progress class="progress is-success is-small" value={100} max="100">{100}%</progress>
    </div>
  {/each}

  <div style="position: relative;">
    <progress class="progress is-success is-large" value={remainder} max="100"
      >{remainder}%</progress
    >
    <span
      class="has-text-weight-bold progress-percent {remainder < 10 && 'less-than-10'}"
      style={remainder > 10 ? `right: ${100 - remainder}%;` : `left: ${remainder}%;`}
      >{formatPercentage(progressPercent / 100, 2)}</span
    >
  </div>
</div>

<style>
  span.progress-percent {
    position: absolute;
    top: 0;
    margin-right: 5px;
    margin-top: 2px;
    color: white;
  }

  span.progress-percent.less-than-10 {
    margin-left: 5px;
    color: #0a0a0a;
  }
</style>
