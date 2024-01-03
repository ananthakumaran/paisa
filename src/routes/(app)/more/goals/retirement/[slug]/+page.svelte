<script lang="ts">
  import COLORS from "$lib/colors";
  import {
    ajax,
    formatCurrency,
    formatFloat,
    isMobile,
    type AssetBreakdown,
    type Point,
    type Posting
  } from "$lib/utils";
  import { onMount, tick, onDestroy } from "svelte";
  import ARIMAPromise from "arima/async";
  import { forecast, renderProgress, findBreakPoints, renderInvestmentTimeline } from "$lib/goals";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import type { PageData } from "./$types";
  import { iconGlyph } from "$lib/icon";
  import _ from "lodash";
  import PostingGroup from "$lib/components/PostingGroup.svelte";
  import PostingCard from "$lib/components/PostingCard.svelte";
  import ProgressWithBreakpoints from "$lib/components/ProgressWithBreakpoints.svelte";
  import AssetsBalance from "$lib/components/AssetsBalance.svelte";
  import BoxLabel from "$lib/components/BoxLabel.svelte";

  export let data: PageData;

  let svg: Element;
  let investmentTimelineSvg: Element;
  let savingsTotal = 0,
    icon = "",
    name = "",
    targetSavings = 0,
    swr = 0,
    xirr = 0,
    yearlyExpense = 0,
    progressPercent = 0,
    savingsX = 0,
    targetX = 0,
    breakPoints: Point[] = [],
    savingsTimeline: Point[] = [],
    postings: Posting[] = [],
    latestPostings: Posting[] = [],
    balances: Record<string, AssetBreakdown> = {},
    destroyCallback = () => {};

  onDestroy(async () => {
    destroyCallback();
  });

  onMount(async () => {
    ({ savingsTotal, savingsTimeline, yearlyExpense, swr, xirr, icon, name, postings, balances } =
      await ajax("/api/goals/retirement/:name", null, data));
    targetSavings = yearlyExpense * (100 / swr);

    latestPostings = _.chain(postings)
      .sortBy((p) => p.date)
      .reverse()
      .take(100)
      .value();

    if (yearlyExpense > 0) {
      progressPercent = (savingsTotal / targetSavings) * 100;
      savingsX = savingsTotal / yearlyExpense;
      targetX = targetSavings / yearlyExpense;
    }

    if (targetX <= 0 || savingsX <= 0 || yearlyExpense <= 0) {
      return;
    }

    const ARIMA = await ARIMAPromise;
    const predictionsTimeline = forecast(savingsTimeline, targetSavings, ARIMA);
    await tick();
    breakPoints = findBreakPoints(savingsTimeline.concat(predictionsTimeline), targetSavings);
    destroyCallback = renderProgress(savingsTimeline, predictionsTimeline, breakPoints, svg, {
      targetSavings
    });

    renderInvestmentTimeline(postings, investmentTimelineSvg, 0);
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
        subtitle="{formatFloat(savingsX, 0)}x times Yearly Expenses"
      />
      <LevelItem
        title="Yearly Expenses"
        color={COLORS.lossText}
        value={formatCurrency(yearlyExpense)}
      />

      <LevelItem
        title="Target Savings"
        value={formatCurrency(targetSavings)}
        color={COLORS.secondary}
        subtitle="{formatFloat(targetX, 0)}x times Yearly Expenses"
      />
      <LevelItem title="SWR" value={formatFloat(swr)} />
      <LevelItem title="XIRR" value={formatFloat(xirr)} />
    </nav>
  </div>
</section>

<section class="section">
  <div class="container is-fluid">
    <ProgressWithBreakpoints {progressPercent} {breakPoints} />
  </div>
</section>

<section class="section tab-retirement-progress">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-9">
        <div class="columns flex-wrap">
          <div class="column is-12">
            <div class="box overflow-x-auto">
              <svg height="400" bind:this={svg} />
            </div>
          </div>
        </div>
        <BoxLabel text="{iconGlyph(icon)} {name} progress" />
        <div class="columns">
          <div class="column is-12">
            <div class="box overflow-x-auto">
              <svg height="300" width="100%" bind:this={investmentTimelineSvg} />
            </div>
          </div>
        </div>
        <BoxLabel text="Monthly Investment" />
        <div class="columns">
          <div class="column is-12 has-text-centered has-text-grey">
            <AssetsBalance breakdowns={balances} indent={false} />
          </div>
        </div>
        <BoxLabel text="Current Balance" />
      </div>
      <div class="column is-3">
        <PostingGroup postings={latestPostings} groupFormat="MMM YYYY" let:groupedPostings>
          <div>
            {#each groupedPostings as posting}
              <PostingCard
                {posting}
                color={posting.amount >= 0
                  ? posting.account.startsWith("Income:CapitalGains")
                    ? COLORS.lossText
                    : COLORS.gainText
                  : posting.account.startsWith("Income:CapitalGains")
                    ? COLORS.gainText
                    : COLORS.lossText}
              />
            {/each}
          </div>
        </PostingGroup>
      </div>
    </div>
  </div>
</section>
