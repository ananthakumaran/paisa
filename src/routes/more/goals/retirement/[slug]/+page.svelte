<script lang="ts">
  import COLORS from "$lib/colors";
  import Progress from "$lib/components/Progress.svelte";
  import { ajax, formatCurrency, formatFloat, type Point } from "$lib/utils";
  import { onMount, tick, onDestroy } from "svelte";
  import ARIMAPromise from "arima/async";
  import { forecast, renderProgress, findBreakPoints } from "$lib/goals";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  let svg: Element;
  let savingsTotal = 0,
    targetSavings = 0,
    swr = 0,
    xirr = 0,
    yearlyExpense = 0,
    progressPercent = 0,
    savingsX = 0,
    targetX = 0,
    breakPoints: Point[] = [],
    savingsTimeline: Point[] = [],
    destroyCallback = () => {};

  onDestroy(async () => {
    destroyCallback();
  });

  onMount(async () => {
    ({
      savings_total: savingsTotal,
      savings_timeline: savingsTimeline,
      yearly_expense: yearlyExpense,
      swr,
      xirr
    } = await ajax("/api/goals/retirement/:name", null, data));
    targetSavings = yearlyExpense * (100 / swr);

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
  });
</script>

<section class="section">
  <div class="container is-fluid">
    <nav class="level">
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
    <Progress {progressPercent} />
  </div>
</section>

<section class="section tab-retirement-progress">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box overflow-x-auto">
          <svg height="500" bind:this={svg} />
        </div>
      </div>
    </div>
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div>
          <p class="heading">Retirement Progress</p>
        </div>
      </div>
    </div>
  </div>
</section>
