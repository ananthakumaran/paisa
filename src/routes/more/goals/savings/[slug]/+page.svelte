<script lang="ts">
  import COLORS from "$lib/colors";
  import Progress from "$lib/components/Progress.svelte";
  import {
    ajax,
    formatCurrency,
    formatFloat,
    isMobile,
    type Point,
    type Posting
  } from "$lib/utils";
  import { onMount, tick, onDestroy } from "svelte";
  import ARIMAPromise from "arima/async";
  import { forecast, renderProgress, findBreakPoints } from "$lib/goals";
  import _ from "lodash";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import type { PageData } from "./$types";
  import PostingCard from "$lib/components/PostingCard.svelte";
  import PostingGroup from "$lib/components/PostingGroup.svelte";
  import { iconGlyph } from "$lib/icon";

  export let data: PageData;

  let svg: Element;
  let savingsTotal = 0,
    targetSavings = 0,
    xirr = 0,
    name = "",
    icon = "",
    progressPercent = 0,
    breakPoints: Point[] = [],
    savingsTimeline: Point[] = [],
    postings: Posting[] = [],
    destroyCallback = () => {};

  onDestroy(async () => {
    destroyCallback();
  });

  onMount(async () => {
    ({
      savingsTotal,
      savingsTimeline,
      target: targetSavings,
      postings,
      icon,
      name,
      xirr
    } = await ajax("/api/goals/savings/:name", null, data));

    postings = _.chain(postings)
      .sortBy((p) => p.date)
      .reverse()
      .take(100)
      .value();

    progressPercent = (savingsTotal / targetSavings) * 100;

    const ARIMA = await ARIMAPromise;
    const predictionsTimeline = forecast(savingsTimeline, targetSavings, ARIMA);
    await tick();
    breakPoints = findBreakPoints(savingsTimeline.concat(predictionsTimeline), targetSavings);
    destroyCallback = renderProgress(savingsTimeline, predictionsTimeline, breakPoints, svg, {
      targetSavings
    });
  });
</script>

<section class="section">
  <div class="container is-fluid">
    <nav class="level custom-icon {isMobile() && 'grid-2'}">
      <LevelItem title={name} value={iconGlyph(icon)} />
      <LevelItem
        title="Current Savings"
        value={formatCurrency(savingsTotal)}
        color={COLORS.gainText}
      />

      <LevelItem
        title="Target Savings"
        value={formatCurrency(targetSavings)}
        color={COLORS.secondary}
      />

      <LevelItem title="XIRR" value={formatFloat(xirr)} />
    </nav>
  </div>
</section>

<section class="section">
  <div class="container is-fluid">
    <Progress {progressPercent} />
  </div>
</section>

<section class="section tab-retirement-progress">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-3">
        <PostingGroup {postings} groupFormat="MMM YYYY" let:groupedPostings>
          <div>
            {#each groupedPostings as posting}
              <PostingCard
                {posting}
                color={posting.amount >= 0
                  ? posting.account.startsWith("Income:CapitalGains")
                    ? COLORS.loss
                    : COLORS.secondary
                  : posting.account.startsWith("Income:CapitalGains")
                  ? COLORS.gain
                  : COLORS.tertiary}
              />
            {/each}
          </div>
        </PostingGroup>
      </div>
      <div class="column is-9">
        <div class="columns flex-wrap">
          <div class="column is-12">
            <div class="box overflow-x-auto">
              <svg height="500" bind:this={svg} />
            </div>
          </div>
          <div class="column is-12 has-text-centered has-text-grey">
            <div>
              <p class="is-size-5 custom-icon">{iconGlyph(icon)} {name} progress</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</section>
