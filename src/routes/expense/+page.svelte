<script lang="ts">
  import { onMount } from "svelte";
  import dayjs from "dayjs";
  import _ from "lodash";
  import { ajax, type Posting } from "$lib/utils";
  import {
    renderMonthlyExpensesTimeline,
    renderCurrentExpensesBreakdown,
    renderCalendar,
    renderSelectedMonth
  } from "$lib/expense";
  import { month } from "../../store";
  import { writable } from "svelte/store";

  const max = dayjs().format("YYYY-MM");
  let groups = writable([]);
  let z,
    renderer,
    expenses,
    grouped_expenses,
    grouped_incomes,
    grouped_investments,
    grouped_taxes,
    min: string;

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

    let minDate = dayjs();
    _.each(expenses, (p) => (p.timestamp = dayjs(p.date)));
    const parseDate = (group: { [key: string]: Posting[] }) => {
      _.each(group, (ps) => {
        _.each(ps, (p) => {
          p.timestamp = dayjs(p.date);
          if (p.timestamp.isBefore(minDate)) {
            minDate = p.timestamp;
          }
        });
      });
    };
    parseDate(grouped_expenses);
    parseDate(grouped_incomes);
    parseDate(grouped_investments);
    parseDate(grouped_taxes);

    min = minDate.format("YYYY-MM");

    ({ z } = renderMonthlyExpensesTimeline(expenses, groups, month));

    renderer = renderCurrentExpensesBreakdown(z);
  });
</script>

<section class="section tab-expense">
  <div class="container is-fluid">
    <div class="columns is-flex-wrap-wrap">
      <div class="column is-3">
        <div class="columns is-flex-wrap-wrap">
          <div class="column is-full">
            <div class="p-3">
              <nav class="level">
                <div class="level-item has-text-centered">
                  <div>
                    <p class="heading">Income</p>
                    <p class="d3-current-month-income title" />
                  </div>
                </div>
                <div class="level-item has-text-centered">
                  <div>
                    <p class="heading">Tax</p>
                    <p class="d3-current-month-tax title" />
                  </div>
                </div>
              </nav>
            </div>
          </div>
          <div class="column is-full">
            <div class="p-3">
              <nav class="level">
                <div class="level-item has-text-centered">
                  <div>
                    <p class="heading">
                      <span>Net Investment</span><span
                        title="Savings Rate"
                        class="tag ml-2 has-text-weight-semibold d3-current-month-savings-rate"
                      />
                    </p>
                    <p class="d3-current-month-investment title" />
                  </div>
                </div>
                <div class="level-item has-text-centered">
                  <div>
                    <p class="heading">Expenses</p>
                    <p class="d3-current-month-expenses title" />
                  </div>
                </div>
              </nav>
            </div>
          </div>
        </div>
      </div>
      <div class="column is-3">
        <div class="p-3">
          <div id="d3-current-month-expense-calendar" class="d3-calendar">
            <div class="has-text-centered py-1">
              <input
                style="width: 175px"
                class="input is-medium is-size-6"
                required
                type="month"
                id="d3-current-month"
                bind:value={$month}
                {max}
                {min}
                autofocus
              />
            </div>
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
      <div class="column is-full-tablet is-half-fullhd">
        <div class="p-3">
          <svg id="d3-current-month-breakdown" width="100%" />
        </div>
      </div>
    </div>
  </div>
</section>

<section class="section tab-expense">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <svg id="d3-expense-timeline" width="100%" height="500" />
      </div>
    </div>
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div>
          <p class="heading">Monthly Expenses</p>
        </div>
      </div>
    </div>
  </div>
</section>
