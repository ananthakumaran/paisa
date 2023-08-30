<script lang="ts">
  import BudgetCard from "$lib/components/BudgetCard.svelte";
  import { ajax, formatCurrency, type AccountBudget, type Budget, helpUrl } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  import { month, setAllowedDateRange } from "../../../store";
  import PostingCard from "$lib/components/PostingCard.svelte";
  import dayjs from "dayjs";
  import COLORS from "$lib/colors";

  const monthStart = dayjs().startOf("month");
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

<section class="section" class:is-hidden={!isEmpty}>
  <div class="container is-fluid">
    <div class="columns is-centered">
      <div class="column is-4 has-text-centered">
        <article class="message">
          <div class="message-body">
            <strong>Oops!</strong> You haven't set a budget yet. Checkout the
            <a href={helpUrl("budget")}>docs</a> page to get started.
          </div>
        </article>
      </div>
    </div>
  </div>
</section>

<section class="section">
  <div class="container is-fluid">
    <div class="columns is-flex-wrap-wrap is-centered">
      {#if currentMonthBudget}
        <div class="column is-12">
          <nav class="level">
            <div class="level-item has-text-centered">
              <div>
                <p class="heading">Checking Current Balance</p>
                <p class="title has-text-black-ter">
                  {formatCurrency(checkingBalance)}
                </p>
              </div>
            </div>
            <div class="level-item has-text-centered">
              <div>
                <p class="heading">
                  {availableForBudgeting >= 0 ? "Available for Budgeting" : "Budget Deficit"}
                </p>
                <p
                  class="title"
                  style="background-color: {availableForBudgeting >= 0
                    ? COLORS.gainText
                    : COLORS.lossText};"
                >
                  {formatCurrency(Math.abs(availableForBudgeting))}
                </p>
              </div>
            </div>
            {#if currentMonthBudget.date.isSameOrAfter(monthStart)}
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Available for Spending</p>
                  <p class="title has-text-black-ter">
                    {formatCurrency(currentMonthBudget.availableThisMonth)}
                  </p>
                </div>
              </div>
              <div class="level-item has-text-centered">
                <div>
                  <p class="heading">Projected Month End Balance</p>
                  <p class="title has-text-black-ter">
                    {formatCurrency(currentMonthBudget.endOfMonthBalance)}
                  </p>
                </div>
              </div>
            {/if}
          </nav>
        </div>
      {/if}
      <div class="column">
        <div class="is-flex">
          <div style="max-width: 800px; min-width: 500px; width: 100%; margin: auto;">
            {#each currentMonthAccountBudgets as accountBudget (accountBudget)}
              <BudgetCard {accountBudget} />
            {/each}
          </div>
        </div>
      </div>
    </div>
  </div>
</section>
