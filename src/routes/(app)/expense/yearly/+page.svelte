<script lang="ts">
  import * as d3 from "d3";
  import { onMount } from "svelte";
  import _ from "lodash";
  import { ajax, formatCurrency, formatPercentage, type Legend, type Posting } from "$lib/utils";
  import {
    renderYearlyExpensesTimeline,
    renderCurrentExpensesBreakdown,
    renderCalendar
  } from "$lib/expense/yearly";
  import { filterRefunds } from "$lib/expense";
  import { dateMin, dateMax, year } from "../../../../store";
  import { writable } from "svelte/store";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import COLORS from "$lib/colors";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import LegendCard from "$lib/components/LegendCard.svelte";
  import { _ as t } from "$lib/i18n";
  import { get } from "svelte/store";

  let groups = writable([]);
  let z: d3.ScaleOrdinal<string, string, never>,
    renderer: (ps: Posting[]) => void,
    expenses: Posting[],
    grouped_expenses: Record<string, Posting[]>,
    grouped_incomes: Record<string, Posting[]>,
    grouped_investments: Record<string, Posting[]>,
    grouped_taxes: Record<string, Posting[]>;

  // M3-G refund toggle (see monthly page for rationale). Same key so
  // both views share the user's choice.
  const SHOW_GROSS_KEY = "expense.show_gross";
  let showGross = false;
  function persistShowGross() {
    try {
      localStorage.setItem(SHOW_GROSS_KEY, showGross ? "1" : "0");
    } catch (_e) {
      // ignore
    }
  }

  let currentYearExpenses: Posting[] = [];

  let legends: Legend[] = [];

  let income = "",
    netIncome = "",
    taxRate = "",
    tax = "",
    expenseRate = "",
    expense = "",
    investment = "",
    savingRate = "";

  $: if (grouped_expenses) {
    currentYearExpenses = filterRefunds(grouped_expenses[$year] || [], showGross);
    renderCalendar(currentYearExpenses, z, $groups);

    const expenses = currentYearExpenses;
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
      const tt = get(t);
      netIncome =
        formatCurrency(sum(incomes, -1) - sum(taxes)) + " " + tt("page.expense.net_income_suffix");
      taxRate =
        formatPercentage(sum(taxes) / sum(incomes, -1)) + " " + tt("page.expense.of_income_suffix");
      expenseRate =
        formatPercentage(sum(expenses) / (sum(incomes, -1) - sum(taxes))) +
        " " +
        tt("page.expense.of_net_income_suffix");
      savingRate =
        formatPercentage(sum(investments) / (sum(incomes, -1) - sum(taxes))) +
        " " +
        tt("page.expense.of_net_income_suffix");
    }

    renderer(expenses);
  }

  function rebuildCharts() {
    if (!expenses) return;
    const filtered = filterRefunds(expenses, showGross);
    const timelineSvg = document.getElementById("d3-yearly-expense-timeline");
    if (timelineSvg) timelineSvg.innerHTML = "";
    const breakdownSvg = document.getElementById("d3-current-year-breakdown");
    if (breakdownSvg) breakdownSvg.innerHTML = "";
    ({ z, legends } = renderYearlyExpensesTimeline(filtered, groups, year));
    renderer = renderCurrentExpensesBreakdown(z);
  }

  function toggleShowGross() {
    showGross = !showGross;
    persistShowGross();
    rebuildCharts();
  }

  onMount(async () => {
    try {
      showGross = localStorage.getItem(SHOW_GROSS_KEY) === "1";
    } catch (_e) {
      showGross = false;
    }

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

    const filtered = filterRefunds(expenses, showGross);
    ({ z, legends } = renderYearlyExpensesTimeline(filtered, groups, year));

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
                  title={$t("page.expense.gross_income")}
                  value={income}
                  color={COLORS.gainText}
                  subtitle={netIncome}
                />
                <LevelItem
                  title={$t("page.expense.tax")}
                  value={tax}
                  color={COLORS.lossText}
                  subtitle={taxRate}
                />
              </nav>
            </div>
          </div>
          <div class="column is-full">
            <div>
              <nav class="level grid-2">
                <LevelItem
                  title={$t("page.expense.net_investment")}
                  value={investment}
                  color={COLORS.secondary}
                  subtitle={savingRate}
                />

                <LevelItem
                  title={$t("page.expense.expenses")}
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
            <strong>{$t("common.hurray")}</strong>
            {$t("page.expense.no_expenses_this_year")}
          </ZeroState>
          <svg id="d3-current-year-breakdown" width="100%" />
        </div>
      </div>
      <div class="column is-12">
        <div class="box">
          <ZeroState item={expenses}>
            <strong>{$t("common.oops")}</strong>
            {$t("page.expense.no_expenses")}
          </ZeroState>

          <div class="is-flex is-justify-content-space-between is-align-items-center ml-4 mr-4">
            <LegendCard {legends} />
            <label class="checkbox is-size-7 has-text-grey ml-3">
              <input
                type="checkbox"
                data-testid="expense-show-gross-toggle"
                checked={showGross}
                on:change={toggleShowGross}
              />
              {showGross ? $t("page.expense.show_gross") : $t("page.expense.show_net")}
            </label>
          </div>
          <svg id="d3-yearly-expense-timeline" width="100%" height="500" />
        </div>
      </div>
    </div>
    <BoxLabel text={$t("page.expense.yearly_expenses")} />
  </div>
</section>
