<script lang="ts">
  import BudgetCard from "$lib/components/BudgetCard.svelte";
  import {
    ajax,
    formatCurrency,
    type AccountBudget,
    type Budget,
    helpUrl,
    now,
    isMobile
  } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  import { month, setAllowedDateRange } from "../../../../store";
  import COLORS from "$lib/colors";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import { _ as t } from "$lib/i18n";

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
          <nav class="level {isMobile() && 'grid-2'}">
            <LevelItem
              title={$t("page.expense.checking_current_balance")}
              value={formatCurrency(checkingBalance)}
            />
            <LevelItem
              title={availableForBudgeting >= 0
                ? $t("page.expense.available_for_budgeting")
                : $t("page.expense.budget_deficit")}
              color={availableForBudgeting >= 0 ? COLORS.gainText : COLORS.lossText}
              value={formatCurrency(Math.abs(availableForBudgeting))}
            />

            {#if currentMonthBudget.date.isSameOrAfter(monthStart)}
              <LevelItem
                title={$t("page.expense.available_for_spending")}
                value={formatCurrency(currentMonthBudget.availableThisMonth)}
                subtitle="{$t('page.expense.out_of_budgeted_pre')}{formatCurrency(
                  currentMonthBudget.forecast
                )}{$t('page.expense.out_of_budgeted_post')}"
              />

              <LevelItem
                title={$t("page.expense.projected_month_end_balance")}
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
              <strong>{$t("common.oops")}</strong>
              {$t("page.expense.no_budget_pre")}<a href={helpUrl("budget")}
                >{$t("page.expense.docs")}</a
              >{$t("page.expense.no_budget_post")}
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
