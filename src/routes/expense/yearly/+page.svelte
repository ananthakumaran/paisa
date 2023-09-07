<script lang="ts">
  import * as d3 from "d3";
  import { onMount } from "svelte";
  import _ from "lodash";
  import { ajax, formatCurrency, formatPercentage, type Posting } from "$lib/utils";
  import {
    renderYearlyExpensesTimeline,
    renderCurrentExpensesBreakdown,
    renderCalendar
  } from "$lib/expense/yearly";
  import { dateMin, dateMax, year } from "../../../store";
  import { writable } from "svelte/store";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import COLORS from "$lib/colors";
  import ZeroState from "$lib/components/ZeroState.svelte";

  let groups = writable([]);
  let z: d3.ScaleOrdinal<string, string, never>,
    renderer: (ps: Posting[]) => void,
    expenses: Posting[],
    grouped_expenses: Record<string, Posting[]>,
    grouped_incomes: Record<string, Posting[]>,
    grouped_investments: Record<string, Posting[]>,
    grouped_taxes: Record<string, Posting[]>;

  let currentYearExpenses: Posting[] = [];

  let income = "",
    netIncome = "",
    taxRate = "",
    tax = "",
    expenseRate = "",
    expense = "",
    investment = "",
    savingRate = "";

  $: if (grouped_expenses) {
    currentYearExpenses = grouped_expenses[$year];
    renderCalendar(currentYearExpenses, z, $groups);

    const expenses = grouped_expenses[$year] || [];
    const incomes = grouped_incomes[$year] || [];
    const taxes = grouped_taxes[$year] || [];
    const investments = grouped_investments[$year] || [];

    income = sumCurrency(incomes, -1);

    tax = sumCurrency(taxes);
    expense = sumCurrency(expenses);
    investment = sumCurrency(investments);

    if (_.isEmpty(incomes)) {
      expenseRate = "";
      taxRate = "";
      savingRate = "";
      netIncome = "";
    } else {
      netIncome = formatCurrency(sum(incomes, -1) - sum(taxes)) + " net income";
      taxRate = formatPercentage(sum(taxes) / sum(incomes, -1)) + " of income";
      expenseRate =
        formatPercentage(sum(expenses) / (sum(incomes, -1) - sum(taxes))) + " of net income";
      savingRate =
        formatPercentage(sum(investments) / (sum(incomes, -1) - sum(taxes))) + " of net income";
    }

    renderer(expenses);
  }

  onMount(async () => {
    ({
      expenses: expenses,
      year_wise: {
        expenses: grouped_expenses,
        incomes: grouped_incomes,
        investments: grouped_investments,
        taxes: grouped_taxes
      }
    } = await ajax("/api/expense"));

    const [start, end] = d3.extent(_.map(expenses, (e) => e.date));
    if (start) {
      dateMin.set(start);
      dateMax.set(end);
    }

    ({ z } = renderYearlyExpensesTimeline(expenses, groups, year));

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
                  title="Gross Income"
                  value={income}
                  color={COLORS.gainText}
                  subtitle={netIncome}
                />
                <LevelItem title="Tax" value={tax} color={COLORS.lossText} subtitle={taxRate} />
              </nav>
            </div>
          </div>
          <div class="column is-full">
            <div>
              <nav class="level grid-2">
                <LevelItem
                  title="Net Investment"
                  value={investment}
                  color={COLORS.secondary}
                  subtitle={savingRate}
                />

                <LevelItem
                  title="Expenses"
                  value={expense}
                  color={COLORS.lossText}
                  subtitle={expenseRate}
                />
              </nav>
            </div>
          </div>
        </div>
      </div>
      <div class="column is-3">
        <div class="px-3 box">
          <div id="d3-current-year-expense-calendar" class="d3-calendar">
            <div class="months" />
          </div>
        </div>
      </div>
      <div class="column is-full-tablet is-half-fullhd">
        <div class="px-3 box" style="height: 100%">
          <ZeroState item={currentYearExpenses}>
            <strong>Hurray!</strong> You have no expenses this year.
          </ZeroState>
          <svg id="d3-current-year-breakdown" width="100%" />
        </div>
      </div>
      <div class="column is-12">
        <div class="box">
          <ZeroState item={expenses}>
            <strong>Oops!</strong> You have no expenses.
          </ZeroState>

          <svg id="d3-yearly-expense-timeline" width="100%" height="500" />
        </div>
      </div>
      <div class="column is-12 has-text-centered">
        <div>
          <p class="heading">Yearly Expenses</p>
        </div>
      </div>
    </div>
  </div>
</section>
