<script lang="ts">
  import Progress from "$lib/components/Progress.svelte";
  import { formatCurrencyCrude, type Point } from "$lib/utils";
  import _ from "lodash";
  export let progressPercent: number;
  export let breakPoints: Point[];

  $: spacers = _.range(breakPoints.length, 4);
</script>

<div>
  {#if !_.isEmpty(breakPoints)}
    <div class="flex justify-between">
      <div></div>
      {#each breakPoints as point, i}
        <div class="breakpoint is-hidden-mobile box py-1 px-4 mb-3 has-text-centered">
          <div class="has-text-grey-light has-text-weight-bold is-size-7">
            {point.date.format("DD MMM YYYY")}
          </div>
          <div class="flex flex-row justify-center items-baseline">
            <div class="mr-2">
              <span
                class="icon is-small {progressPercent >= (i + 1) * 25
                  ? 'has-text-success'
                  : 'has-text-grey'}"
              >
                <i class="fas fa-check-circle" />
              </span>
            </div>
            <div>{(i + 1) * 25}%</div>
            <div class="ml-1 has-text-grey-light has-text-weight-bold is-size-7">
              {formatCurrencyCrude(point.value)}
            </div>
          </div>
        </div>
      {/each}
      {#each spacers as _s}
        <div></div>
      {/each}
    </div>
  {/if}
  <Progress {progressPercent} />
</div>

<style lang="scss">
</style>