<script lang="ts">
  import BudgetCard from "$lib/components/BudgetCard.svelte";
  import { ajax, formatCurrency, type AccountBudget, type Budget, helpUrl, now } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  import { month, setAllowedDateRange } from "../../../store";
  import COLORS from "$lib/colors";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";

  const monthStart = now().startOf("month");
  let budgetsByMonth: Record<string, Budget> = {};
  let currentMonthAccountBudgets: AccountBudget[] = [];
  let currentMonthBudget: Budget;
  let checkingBalance: number, availableForBudgeting: number;
  let isEmpty = false;

  $: {
    currentMonthBudget = budgetsByMonth[$month];
    currentMonthAccountBudgets = budgetsByMonth[$month]?.accounts || [];
  }

  onMount(async () => {
    ({ budgetsByMonth, checkingBalance, availableForBudgeting } = await ajax("/api/budget"));
    setAllowedDateRange(
      _.chain(budgetsByMonth)
        .values()
        .flatten()
        .map((b) => b.date)
        .value()
    );

    if (_.isEmpty(budgetsByMonth)) {
      isEmpty = true;
    }
  });
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns is-flex-wrap-wrap is-centered">
      {#if currentMonthBudget}
        <div class="column is-12">
          <nav class="level">
            <LevelItem title="Checking Current Balance" value={formatCurrency(checkingBalance)} />
            <LevelItem
              title={availableForBudgeting >= 0 ? "Available for Budgeting" : "Budget Deficit"}
              color={availableForBudgeting >= 0 ? COLORS.gainText : COLORS.lossText}
              value={formatCurrency(Math.abs(availableForBudgeting))}
            />

            {#if currentMonthBudget.date.isSameOrAfter(monthStart)}
              <LevelItem
                title="Available for Spending"
                value={formatCurrency(currentMonthBudget.availableThisMonth)}
                subtitle="out of {formatCurrency(currentMonthBudget.forecast)} budgeted"
              />

              <LevelItem
                title="Projected Month End Balance"
                value={formatCurrency(currentMonthBudget.endOfMonthBalance)}
              />
            {/if}
          </nav>
        </div>
      {/if}
      <div class="column">
        <div class="is-flex">
          <div style="max-width: 800px; min-width: 350px; width: 100%; margin: auto;">
            <ZeroState item={!isEmpty}>
              <strong>Oops!</strong> You haven't set a budget yet. Checkout the
              <a href={helpUrl("budget")}>docs</a> page to get started.
            </ZeroState>

            {#each currentMonthAccountBudgets as accountBudget (accountBudget)}
              <BudgetCard {accountBudget} />
            {/each}
          </div>
        </div>
      </div>
    </div>
  </div>
</section>
