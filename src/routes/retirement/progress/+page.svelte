<script lang="ts">
  import COLORS from "$lib/colors";
  import { ajax, formatCurrency, formatFloat, formatPercentage } from "$lib/utils";
  import { onMount } from "svelte";

  let savingsTotal = 0,
    targetSavings = 0,
    swr = 0,
    yearlyExpense = 0,
    progressPercent = 0;

  onMount(async () => {
    ({
      savings_total: savingsTotal,
      yearly_expense: yearlyExpense,
      swr
    } = await ajax("/api/retirement/progress"));
    targetSavings = yearlyExpense * (100 / swr);
    progressPercent = (savingsTotal / targetSavings) * 100;
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
  <div class="container is-fluid">
    <div style="position: relative;">
      <progress class="progress is-success is-large" value={progressPercent} max="100"
        >{progressPercent}%</progress
      >
      <span
        class="has-text-weight-bold progress-percent {progressPercent < 10 && 'less-than-10'}"
        style="left: {progressPercent}%;">{formatPercentage(progressPercent / 100, 2)}</span
      >
    </div>
  </div>
</section>

<style>
  .level-item p.title {
    padding: 5px;
    color: white;
  }

  span.progress-percent {
    position: absolute;
    top: 0;
    margin-left: -50px;
    margin-top: 2px;
    color: white;
  }

  span.progress-percent.less-than-10 {
    margin-left: 5px;
    color: #0a0a0a;
  }
</style>
