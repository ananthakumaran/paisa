<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import _ from "lodash";
  import {
    ajax,
    secondName,
    type Posting,
    formatCurrency,
    formatPercentage,
    type Legend
  } from "$lib/utils";
  import {
    renderMonthlyExpensesTimeline,
    renderCurrentExpensesBreakdown,
    renderCalendar
  } from "$lib/expense/monthly";
  import { filterRefunds } from "$lib/expense";
  import { dateRange, month, setAllowedDateRange } from "../../../../store";
  import { writable } from "svelte/store";
  import PostingCard from "$lib/components/PostingCard.svelte";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import COLORS from "$lib/colors";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import dayjs from "dayjs";
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
    grouped_taxes: Record<string, Posting[]>,
    destroy: () => void;

  // M3-G refund toggle. Default = net (the user's intuitive "支出").
  // Persisted under `expense.show_gross` so the choice survives reloads.
  const SHOW_GROSS_KEY = "expense.show_gross";
  let showGross = false;
  function persistShowGross() {
    try {
      localStorage.setItem(SHOW_GROSS_KEY, showGross ? "1" : "0");
    } catch (_e) {
      // localStorage may be unavailable (private mode); ignore.
    }
  }

  let legends: Legend[] = [];

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
    current_month_expenses = _.chain(
      filterRefunds((grouped_expenses && grouped_expenses[$month]) || [], showGross)
    )
      .filter((e) => _.includes($groups, secondName(e.account)))
      .sortBy((e) => e.date)
      .reverse()
      .value();
  }

  $: if (grouped_expenses) {
    const monthlyExpenses = filterRefunds(grouped_expenses[$month] || [], showGross);
    renderCalendar($month, monthlyExpenses, z, $groups);

    const expenses = monthlyExpenses;
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
      const tt = get(t);
      netIncome =
        formatCurrency(sum(incomes, -1) - sum(taxes)) + " " + tt("page.expense.net_income_suffix");
      taxRate =
        formatPercentage(sum(taxes) / sum(incomes, -1)) + " " + tt("page.expense.on_income_suffix");
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

  onDestroy(async () => {
    if (destroy) {
      destroy();
    }
  });

  // Rebuild the timeline + breakdown when the refund toggle flips.
  // We can't just re-call render(): the closure captures the original
  // posting set, the d3 stack keys, and the color scale. Easiest is to
  // tear down the existing renderers, clear the SVGs, and re-init from
  // the filtered postings. The toggle is a low-frequency action, so the
  // cost of re-binding is fine.
  async function rebuildCharts() {
    if (!expenses) return;
    if (destroy) destroy();
    const filtered = filterRefunds(expenses, showGross);
    const timelineSvg = document.getElementById("d3-monthly-expense-timeline");
    if (timelineSvg) timelineSvg.innerHTML = "";
    const breakdownSvg = document.getElementById("d3-current-month-breakdown");
    if (breakdownSvg) breakdownSvg.innerHTML = "";
    ({ z, destroy, legends } = renderMonthlyExpensesTimeline(filtered, groups, month, dateRange));
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
      month_wise: {
        expenses: grouped_expenses,
        incomes: grouped_incomes,
        investments: grouped_investments,
        taxes: grouped_taxes
      }
    } = await ajax("/api/expense"));

    setAllowedDateRange(_.map(expenses, (e) => e.date));
    const filtered = filterRefunds(expenses, showGross);
    ({ z, destroy, legends } = renderMonthlyExpensesTimeline(filtered, groups, month, dateRange));
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
                  title={$t("page.expense.gross_income")}
                  value={income}
                  color={COLORS.gainText}
                  subtitle={netIncome}
                />
                <LevelItem
                  narrow
                  title={$t("page.expense.tax")}
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
                  title={$t("page.expense.net_investment")}
                  value={saving}
                  subtitle={savingRate}
                  color={COLORS.secondary}
                />

                <LevelItem
                  narrow
                  title={$t("page.expense.expenses")}
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
              <ZeroState item={grouped_expenses?.[$month]}>
                <strong>{$t("common.hurray")}</strong>
                {$t("page.expense.no_expenses_this_month")}
              </ZeroState>
              <svg id="d3-current-month-breakdown" width="100%" />
            </div>
          </div>
          <div class="column is-full">
            <div class="box">
              <ZeroState item={expenses}>
                <strong>{$t("common.oops")}</strong>
                {$t("page.expense.no_expenses")}
              </ZeroState>
              <div class="is-flex is-justify-content-space-between is-align-items-center ml-4 mr-4">
                <LegendCard {legends} clazz="overflow-x-auto" />
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
              <svg id="d3-monthly-expense-timeline" width="100%" height="400" />
            </div>
          </div>
        </div>
        <BoxLabel text={$t("page.expense.monthly_expenses")} />
      </div>
    </div>
  </div>
</section>
