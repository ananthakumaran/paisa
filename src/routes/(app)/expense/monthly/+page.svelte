<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import _ from "lodash";
  import { ajax, secondName, type Posting, formatCurrency, formatPercentage } from "$lib/utils";
  import {
    renderMonthlyExpensesTimeline,
    renderCurrentExpensesBreakdown,
    renderCalendar
  } from "$lib/expense/monthly";
  import { dateRange, month, setAllowedDateRange } from "../../../../store";
  import { writable } from "svelte/store";
  import PostingCard from "$lib/components/PostingCard.svelte";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import COLORS from "$lib/colors";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import dayjs from "dayjs";

  let groups = writable([]);
  let z: d3.ScaleOrdinal<string, string, never>,
    renderer: (ps: Posting[]) => void,
    expenses: Posting[],
    grouped_expenses: Record<string, Posting[]>,
    grouped_incomes: Record<string, Posting[]>,
    grouped_investments: Record<string, Posting[]>,
    grouped_taxes: Record<string, Posting[]>,
    destroy: () => void;

  let taxRate = "",
    netIncome = "",
    tax = "",
    expenseRate = "",
    expense = "",
    saving = "",
    savingRate = "",
    income = "";

  let current_month_expenses: Posting[] = [];

  $: {
    current_month_expenses = _.chain((grouped_expenses && grouped_expenses[$month]) || [])
      .filter((e) => _.includes($groups, secondName(e.account)))
      .sortBy((e) => e.date)
      .reverse()
      .value();
  }

  $: if (grouped_expenses) {
    renderCalendar($month, grouped_expenses[$month], z, $groups);

    const expenses = grouped_expenses[$month] || [];
    const incomes = grouped_incomes[$month] || [];
    const taxes = grouped_taxes[$month] || [];
    const investments = grouped_investments[$month] || [];

    income = sumCurrency(incomes, -1);
    tax = sumCurrency(taxes);
    expense = sumCurrency(expenses);
    saving = sumCurrency(investments);

    if (_.isEmpty(incomes)) {
      taxRate = "";
      expenseRate = "";
      savingRate = "";
      netIncome = "";
    } else {
      netIncome = formatCurrency(sum(incomes, -1) - sum(taxes)) + " net income";
      taxRate = formatPercentage(sum(taxes) / sum(incomes, -1)) + " on income";
      expenseRate =
        formatPercentage(sum(expenses) / (sum(incomes, -1) - sum(taxes))) + " of net income";
      savingRate =
        formatPercentage(sum(investments) / (sum(incomes, -1) - sum(taxes))) + " of net income";
    }

    renderer(expenses);
  }

  onDestroy(async () => {
    if (destroy) {
      destroy();
    }
  });

  onMount(async () => {
    ({
      expenses: expenses,
      month_wise: {
        expenses: grouped_expenses,
        incomes: grouped_incomes,
        investments: grouped_investments,
        taxes: grouped_taxes
      }
    } = await ajax("/api/expense"));

    setAllowedDateRange(_.map(expenses, (e) => e.date));
    ({ z, destroy } = renderMonthlyExpensesTimeline(expenses, groups, month, dateRange));
    renderer = renderCurrentExpensesBreakdown(z);
  });

  function sum(postings: Posting[], sign = 1) {
    return sign * _.sumBy(postings, (p) => p.amount);
  }

  function sumCurrency(postings: Posting[], sign = 1) {
    return formatCurrency(sign * _.sumBy(postings, (p) => p.amount));
  }
</script>

<section class="section tab-expense">
  <div class="container is-fluid">
    <div class="columns is-flex-wrap-wrap">
      <div class="column is-3">
        <div class="columns is-flex-wrap-wrap">
          <div class="column is-full">
            <div>
              <nav class="level grid-2">
                <LevelItem
                  narrow
                  title="Gross Income"
                  value={income}
                  color={COLORS.gainText}
                  subtitle={netIncome}
                />
                <LevelItem
                  narrow
                  title="Tax"
                  value={tax}
                  subtitle={taxRate}
                  color={COLORS.lossText}
                />
              </nav>
            </div>
          </div>
          <div class="column is-full">
            <div>
              <nav class="level grid-2">
                <LevelItem
                  narrow
                  title="Net Investment"
                  value={saving}
                  subtitle={savingRate}
                  color={COLORS.secondary}
                />

                <LevelItem
                  narrow
                  title="Expenses"
                  value={expense}
                  color={COLORS.lossText}
                  subtitle={expenseRate}
                />
              </nav>
            </div>
          </div>
          <div class="column is-full">
            {#each current_month_expenses as expense}
              <PostingCard posting={expense} color={z(secondName(expense.account))} icon={true} />
            {/each}
          </div>
        </div>
      </div>
      <div class="column is-9">
        <div class="columns is-flex-wrap-wrap">
          <div class="column is-4">
            <div class="p-3 box">
              <div id="d3-current-month-expense-calendar" class="d3-calendar">
                <div class="weekdays">
                  {#each dayjs.weekdaysShort(true) as day}
                    <div>{day}</div>
                  {/each}
                </div>
                <div class="days" />
              </div>
            </div>
          </div>
          <div class="column is-8">
            <div class="px-3 box" style="height: 100%">
              <ZeroState item={current_month_expenses}>
                <strong>Hurray!</strong> You have no expenses this month.
              </ZeroState>
              <svg id="d3-current-month-breakdown" width="100%" />
            </div>
          </div>
          <div class="column is-full">
            <div class="box">
              <ZeroState item={expenses}>
                <strong>Oops!</strong> You have no expenses.
              </ZeroState>
              <svg id="d3-monthly-expense-timeline" width="100%" height="400" />
            </div>
          </div>
        </div>
        <BoxLabel text="Monthly Expenses" />
      </div>
    </div>
  </div>
</section>
