<script lang="ts">
  import * as expense from "$lib/expense/monthly";
  import * as cashFlow from "$lib/cash_flow";
  import {
    ajax,
    sortTrantionSequence,
    type CashFlow,
    type Posting,
    type TransactionSequence,
    nextDate,
    totalRecurring,
    formatCurrencyCrude,
    intervalText,
    type Networth,
    formatCurrency,
    formatFloat,
    type Transaction,
    type Budget
  } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  import COLORS from "$lib/colors";
  import dayjs from "dayjs";
  import LastNMonths from "$lib/components/LastNMonths.svelte";
  import TransactionCard from "$lib/components/TransactionCard.svelte";

  import { MasonryGrid } from "@egjs/svelte-grid";
  import BudgetCard from "$lib/components/BudgetCard.svelte";

  let UntypedMasonryGrid = MasonryGrid as any;

  let month = dayjs().format("YYYY-MM");
  let transactionSequences: TransactionSequence[] = [];
  let cashFlows: CashFlow[] = [];
  let expenses: { [key: string]: Posting[] } = {};
  let xirr = 0;
  let networth: Networth;
  let renderer: (data: Posting[]) => void;
  let totalExpense = 0;
  let transactions: Transaction[] = [];
  let budgetsByMonth: Record<string, Budget> = {};
  let currentBudget: Budget;

  const now = dayjs();

  $: if (renderer && expenses[month]) {
    renderer(expenses[month]);
    totalExpense = _.sumBy(expenses[month], (p) => p.amount);
  }

  onMount(async () => {
    ({
      expenses,
      cashFlows,
      budget: { budgetsByMonth },
      transactionSequences,
      networth: { networth, xirr },
      transactions
    } = await ajax("/api/dashboard"));
    const postings = _.chain(expenses).values().flatten().value();
    const z = expense.colorScale(postings);
    renderer = expense.renderCurrentExpensesBreakdown(z);
    currentBudget = budgetsByMonth[month];
    renderer(expenses[month]);

    cashFlow.renderMonthlyFlow("#d3-current-cash-flow", {
      rotate: false,
      balance: _.last(cashFlows)?.balance || 0
    })(cashFlows);
    transactionSequences = _.take(sortTrantionSequence(transactionSequences), 16);
  });
</script>

<section class="section tab-networth">
  <div class="container is-fluid">
    <div class="tile is-ancestor is-align-items-start">
      <div class="tile is-4">
        <div class="tile is-vertical">
          <div class="tile is-parent">
            <div class="tile is-child">
              <div class="content">
                <p class="subtitle">
                  <a class="secondary-link" href="/assets/networth">Assets</a>
                </p>
                <div class="content">
                  <div>
                    {#if networth}
                      <nav class="level">
                        <div class="level-item is-narrow has-text-centered">
                          <div>
                            <p class="heading">Net worth</p>
                            <p class="title" style="background-color: {COLORS.primary};">
                              {formatCurrency(networth.balanceAmount)}
                            </p>
                          </div>
                        </div>
                        <div class="level-item is-narrow has-text-centered">
                          <div>
                            <p class="heading">Net Investment</p>
                            <p class="title" style="background-color: {COLORS.secondary};">
                              {formatCurrency(networth.netInvestmentAmount)}
                            </p>
                          </div>
                        </div>
                      </nav>
                      <nav class="level">
                        <div class="level-item is-narrow has-text-centered">
                          <div>
                            <p class="heading">Gain / Loss</p>
                            <p
                              class="title"
                              style="background-color: {networth.gainAmount >= 0
                                ? COLORS.gainText
                                : COLORS.lossText};"
                            >
                              {formatCurrency(networth.gainAmount)}
                            </p>
                          </div>
                        </div>
                        <div class="level-item is-narrow has-text-centered">
                          <div>
                            <p class="heading">XIRR</p>
                            <p class="title has-text-black-ter">{formatFloat(xirr)}</p>
                          </div>
                        </div>
                      </nav>
                    {/if}
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="tile is-parent">
            <article class="tile is-child">
              <p class="subtitle">
                <a class="secondary-link" href="/cash_flow/monthly">Cash Flow</a>
              </p>
              <div class="content box px-2 pb-0">
                <svg id="d3-current-cash-flow" height="250" width="100%" />
              </div>
            </article>
          </div>
          {#if currentBudget}
            <div class="tile is-parent">
              <div class="tile is-child">
                <div class="content">
                  <p class="subtitle">
                    <a class="secondary-link" href="/expense/budget">Budget</a>
                  </p>
                  <div class="content">
                    <div>
                      {#each currentBudget.accounts as accountBudget (accountBudget)}
                        <BudgetCard compact {accountBudget} selected={false} />
                      {/each}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          {/if}
        </div>
      </div>
      <div class="tile is-vertical">
        <div class="tile is-parent is-12">
          <article class="tile is-child">
            <p class="subtitle is-flex is-justify-content-space-between is-align-items-end">
              <span
                ><a class="secondary-link" href="/expense/monthly">Expenses</a>
                <span
                  class="is-size-5 has-text-weight-bold px-2 has-text-white"
                  style="background-color: {COLORS.lossText};">{formatCurrency(totalExpense)}</span
                ></span
              >
              <LastNMonths n={3} bind:value={month} />
            </p>
            <div class="content box px-3">
              <svg id="d3-current-month-breakdown" width="100%" />
            </div>
          </article>
        </div>
        {#if !_.isEmpty(transactionSequences)}
          <div class="tile">
            <div class="tile is-parent is-12">
              <article class="tile is-child">
                <div class="content">
                  <p class="subtitle">
                    <a class="secondary-link" href="/cash_flow/recurring">Recurring</a>
                  </p>
                  <div class="content box">
                    <div
                      class="is-flex is-justify-content-flex-start is-flex-wrap-wrap"
                      style="overflow: hidden; max-height: 190px"
                    >
                      {#each transactionSequences as ts}
                        {@const n = nextDate(ts)}
                        <div class="has-text-centered mb-3 mr-3 max-w-[200px] border-2">
                          <div class="is-size-7 truncate">{ts.key.tagRecurring}</div>
                          <div class="my-1">
                            <span class="tag is-light">{intervalText(ts)}</span>
                          </div>
                          <div class="has-text-grey is-size-7">
                            <span class="icon has-text-grey-light">
                              <i class="fas fa-calendar" />
                            </span>
                            {n.format("DD MMM YYYY")}
                          </div>
                          <div
                            class="m-3 du-radial-progress is-size-7"
                            style="--value: {n.isBefore(now)
                              ? '0'
                              : (n.diff(now, 'day') / ts.interval) *
                                100}; --thickness: 3px; --size: 100px; color: {n.isBefore(now)
                              ? COLORS.danger
                              : COLORS.success};"
                          >
                            <span>{formatCurrencyCrude(totalRecurring(ts))}</span>
                            <span>due {n.fromNow()}</span>
                          </div>
                        </div>
                      {/each}
                    </div>
                  </div>
                </div>
              </article>
            </div>
          </div>
        {/if}
        {#if !_.isEmpty(transactions)}
          <div class="tile">
            <div class="tile is-parent is-12">
              <article class="tile is-child">
                <div class="content">
                  <p class="subtitle">
                    <a class="secondary-link" href="/ledger/transaction">Recent Transactions</a>
                  </p>
                  <div>
                    <UntypedMasonryGrid gap={10} maxStretchColumnSize={500} align="stretch">
                      {#each _.take(transactions, 20) as t}
                        <div class="mr-3 is-flex-grow-1">
                          <TransactionCard {t} />
                        </div>
                      {/each}
                    </UntypedMasonryGrid>
                  </div>
                </div>
              </article>
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>
</section>

<style lang="scss">
  p.subtitle {
    margin-bottom: 10px;
  }
</style>
