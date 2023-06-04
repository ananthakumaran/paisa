<script lang="ts">
  import COLORS from "$lib/colors";
  import Progress from "$lib/components/Progress.svelte";
  import { ajax, formatCurrency, formatFloat, type Point } from "$lib/utils";
  import { onMount, tick, onDestroy } from "svelte";
  import ARIMAPromise from "arima/async";
  import { forecast, renderProgress, findBreakPoints } from "$lib/retirement";
  import { isEmpty } from "lodash";
  import LevelItem from "$lib/components/LevelItem.svelte";

  let svg: Element;
  let savingsTotal = 0,
    targetSavings = 0,
    swr = 0,
    yearlyExpense = 0,
    progressPercent = 0,
    savingsX = 0,
    targetX = 0,
    skipProgress = true,
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
      swr
    } = await ajax("/api/retirement/progress"));
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
    if (isEmpty(predictionsTimeline)) {
      return;
    }
    skipProgress = false;
    await tick();
    breakPoints = findBreakPoints(savingsTimeline.concat(predictionsTimeline), targetSavings);
    destroyCallback = renderProgress(savingsTimeline, predictionsTimeline, breakPoints, svg, {
      targetSavings,
      yearlyExpense
    });
  });
</script>

<section class="section">
  <div class="container">
    <nav class="level">
      <LevelItem
        title="Current Savings"
        value={formatCurrency(savingsTotal)}
        color={COLORS.primary}
        subtitle="X times Yearly Expenses"
        subvalue="{formatFloat(savingsX, 0)}x"
        subcolor={COLORS.primary}
      />
      <div class="level-item has-text-centered">
        <div>
          <p class="heading">Yearly Expenses</p>
          <p class="title" style="background-color: {COLORS.lossText};">
            {formatCurrency(yearlyExpense)}
          </p>
        </div>
      </div>
      <LevelItem
        title="Target Savings"
        value={formatCurrency(targetSavings)}
        color={targetSavings <= savingsTotal ? COLORS.gainText : COLORS.lossText}
        subtitle="X times Yearly Expenses"
        subvalue="{formatFloat(targetX, 0)}x"
        subcolor={targetSavings <= savingsTotal ? COLORS.gainText : COLORS.lossText}
      />
      <div class="level-item has-text-centered">
        <div>
          <p class="heading">SWR</p>
          <p class="title has-text-black">{formatFloat(swr)}</p>
        </div>
      </div>
    </nav>
  </div>
</section>

<section class="section">
  <Progress {progressPercent} />
</section>

<section
  class="section tab-retirement-progress"
  style="visibility: {skipProgress ? 'hidden' : 'visible'};"
>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <svg width="100%" height="500" bind:this={svg} />
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

<style>
  .level-item p.title {
    padding: 5px;
    color: white;
  }
</style>
