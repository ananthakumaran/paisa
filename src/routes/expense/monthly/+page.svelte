<script lang="ts">
  import { onMount } from "svelte";
  import _ from "lodash";
  import { ajax, formatCurrency, restName, secondName, type Posting, postingUrl } from "$lib/utils";
  import {
    renderMonthlyExpensesTimeline,
    renderCurrentExpensesBreakdown,
    renderCalendar,
    renderSelectedMonth
  } from "$lib/expense/monthly";
  import { dateRange, month, setAllowedDateRange } from "../../../store";
  import { writable } from "svelte/store";
  import { iconify } from "$lib/icon";
  import PostingStatus from "$lib/components/PostingStatus.svelte";

  let groups = writable([]);
  let z: d3.ScaleOrdinal<string, string, never>,
    renderer: (ps: Posting[]) => void,
    expenses: Posting[],
    grouped_expenses: Record<string, Posting[]>,
    grouped_incomes: Record<string, Posting[]>,
    grouped_investments: Record<string, Posting[]>,
    grouped_taxes: Record<string, Posting[]>;

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
    renderSelectedMonth(
      renderer,
      grouped_expenses[$month] || [],
      grouped_incomes[$month] || [],
      grouped_taxes[$month] || [],
      grouped_investments[$month] || []
    );
  }

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
    ({ z } = renderMonthlyExpensesTimeline(expenses, groups, month, dateRange));

    renderer = renderCurrentExpensesBreakdown(z);
  });
</script>

<section class="section tab-expense">
  <div class="container is-fluid">
    <div class="columns is-flex-wrap-wrap">
      <div class="column is-3">
        <div class="columns is-flex-wrap-wrap">
          <div class="column is-full">
            <div>
              <nav class="level">
                <div class="level-item is-narrow has-text-centered">
                  <div>
                    <p class="heading is-flex is-justify-content-space-between">Income</p>
                    <p class="d3-current-month-income title" />
                  </div>
                </div>
                <div class="level-item is-narrow has-text-centered">
                  <div>
                    <p class="heading is-flex is-justify-content-space-between">
                      <span>Tax</span><span
                        title="Tax Rate"
                        class="tag ml-2 has-text-weight-semibold d3-current-month-tax-rate"
                      />
                    </p>
                    <p class="d3-current-month-tax title" />
                  </div>
                </div>
              </nav>
            </div>
          </div>
          <div class="column is-full">
            <div>
              <nav class="level">
                <div class="level-item is-narrow has-text-centered">
                  <div>
                    <p class="heading is-flex is-justify-content-space-between">
                      <span>Net Investment</span><span
                        title="Savings Rate"
                        class="tag ml-2 has-text-weight-semibold d3-current-month-savings-rate"
                      />
                    </p>
                    <p class="d3-current-month-investment title" />
                  </div>
                </div>
                <div class="level-item is-narrow has-text-centered">
                  <div>
                    <p class="heading is-flex is-justify-content-space-between">
                      <span>Expenses</span><span
                        title="Expenses Rate"
                        class="tag ml-2 has-text-weight-semibold d3-current-month-expenses-rate"
                      />
                    </p>
                    <p class="d3-current-month-expenses title" />
                  </div>
                </div>
              </nav>
            </div>
          </div>
          <div class="column is-full">
            {#each current_month_expenses as expense}
              <div
                class="box p-2 my-2 has-background-white"
                style="border-left: 2px solid {z(secondName(expense.account))}"
              >
                <div class="is-flex is-justify-content-space-between">
                  <div class="has-text-grey is-size-7 truncate">
                    <PostingStatus posting={expense} />
                    <a href={postingUrl(expense)}>{expense.payee}</a>
                  </div>
                  <div class="has-text-grey min-w-[100px]">
                    <span class="icon is-small has-text-grey-light">
                      <i class="fas fa-calendar" />
                    </span>
                    {expense.date.format("DD MMM YYYY")}
                  </div>
                </div>
                <hr class="m-1" />
                <div class="is-flex is-flex-wrap-wrap is-justify-content-space-between">
                  <div class="has-text-grey">
                    {iconify(restName(expense.account), { group: "Expenses" })}
                  </div>
                  <div class="has-text-weight-bold is-size-6">
                    {formatCurrency(expense.amount)}
                  </div>
                </div>
              </div>
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
                  <div>Sun</div>
                  <div>Mon</div>
                  <div>Tue</div>
                  <div>Wed</div>
                  <div>Thu</div>
                  <div>Fri</div>
                  <div>Sat</div>
                </div>
                <div class="days" />
              </div>
            </div>
          </div>
          <div class="column is-8">
            <div class="px-3 box" style="height: 100%">
              <svg id="d3-current-month-breakdown" width="100%" />
            </div>
          </div>
          <div class="column is-full">
            <div class="box">
              <svg id="d3-monthly-expense-timeline" width="100%" height="400" />
            </div>
          </div>
          <div class="column is-full has-text-centered">
            <div>
              <p class="heading">Monthly Expenses</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</section>
