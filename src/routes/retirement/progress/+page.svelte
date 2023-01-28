<script lang="ts">
  import COLORS from "$lib/colors";
  import Progress from "$lib/components/Progress.svelte";
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
  <Progress {progressPercent} />
</section>

<style>
  .level-item p.title {
    padding: 5px;
    color: white;
  }
</style>
