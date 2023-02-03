<script lang="ts">
  import COLORS from "$lib/colors";
  import Progress from "$lib/components/Progress.svelte";
  import { ajax, formatCurrency, formatFloat, type Point } from "$lib/utils";
  import { onMount, tick, onDestroy } from "svelte";
  import ARIMAPromise from "arima/async";
  import { forecast, renderProgress, findBreakPoints } from "$lib/retirement";
  import { isEmpty } from "lodash";

  let svg: Element;
  let savingsTotal = 0,
    targetSavings = 0,
    swr = 0,
    yearlyExpense = 0,
    progressPercent = 0,
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
    progressPercent = (savingsTotal / targetSavings) * 100;

    const ARIMA = await ARIMAPromise;
    const predictionsTimeline = forecast(savingsTimeline, targetSavings, ARIMA);
    if (isEmpty(predictionsTimeline)) {
      return;
    }
    skipProgress = false;
    await tick();
    breakPoints = findBreakPoints(savingsTimeline.concat(predictionsTimeline), targetSavings);
    destroyCallback = renderProgress(savingsTimeline, predictionsTimeline, breakPoints, svg);
  });
</script>

<section class="section">
  <div class="container">
    <nav class="level">
      <div class="level-item has-text-centered">
        <div>
          <p class="heading">Current Savings</p>
          <p class="title" style="background-color: {COLORS.primary};">
            {formatCurrency(savingsTotal)}
          </p>
        </div>
      </div>
      <div class="level-item has-text-centered">
        <div>
          <p class="heading">Yearly Expenses</p>
          <p class="title" style="background-color: {COLORS.lossText};">
            {formatCurrency(yearlyExpense)}
          </p>
        </div>
      </div>
      <div class="level-item has-text-centered">
        <div>
          <p class="heading">Target Savings</p>
          <p
            class="title"
            style="background-color: {targetSavings <= savingsTotal
              ? COLORS.gainText
              : COLORS.lossText};"
          >
            {formatCurrency(targetSavings)}
          </p>
        </div>
      </div>
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

<section class="section tab-retirement-progress" class:is-hidden={skipProgress}>
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
