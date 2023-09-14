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
    type Budget,
    formatPercentage
  } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  import COLORS from "$lib/colors";
  import dayjs from "dayjs";
  import LastNMonths from "$lib/components/LastNMonths.svelte";
  import TransactionCard from "$lib/components/TransactionCard.svelte";

  import { MasonryGrid } from "@egjs/svelte-grid";
  import BudgetCard from "$lib/components/BudgetCard.svelte";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import { refresh } from "../store";

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
  let selectedExpenses: Posting[] = [];
  let isEmpty = false;

  const now = dayjs();

  $: if (renderer) {
    selectedExpenses = expenses[month] || [];
    renderer(selectedExpenses);
    totalExpense = _.sumBy(selectedExpenses, (p) => p.amount);
  }

  async function initDemo() {
    await ajax("/api/init", { method: "POST" });
    refresh();
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

    if (_.isEmpty(transactions)) {
      isEmpty = true;
    } else {
      isEmpty = false;
    }

    const postings = _.chain(expenses).values().flatten().value();
    const z = expense.colorScale(postings);
    renderer = expense.renderCurrentExpensesBreakdown(z);
    currentBudget = budgetsByMonth[month];

    cashFlow.renderMonthlyFlow("#d3-current-cash-flow", {
      rotate: false,
      balance: _.last(cashFlows)?.balance || 0
    })(cashFlows);
    transactionSequences = _.take(sortTrantionSequence(transactionSequences), 16);
  });
</script>

<section class="section" class:is-hidden={!isEmpty}>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <ZeroState item={!isEmpty}>
          <div class="has-text-left" style="max-width: 640px;">
            <p class="mb-2">
              Looks like you are new here, you can either get started or look at a demo setup
            </p>
            <div>
              <p class="is-size-4">I want to get started</p>
              <ol class="ml-5 mt-2 mb-4">
                <li>
                  Go to <a href="/config">config</a> page and set your default currency and locale.
                </li>
                <li>
                  Go to <a href="/ledger/editor">editor</a> page and start adding transactions to your
                  journal.
                </li>
              </ol>
              <p class="is-size-4">I want to view a Demo</p>
              <p class="ml-3"></p>
              <ol class="ml-5 mt-2 mb-4">
                <li>
                  Click the button below to load a demo setup. This will load a demo journal with
                  relevant config.
                </li>
                <li>
                  Once you are done playing around, you can go to <a href="/ledger/editor">editor</a
                  > page and select all the content and delete them.
                </li>
                <li>
                  Go to <a href="/config">config</a> page and click the reset to defaults button.
                </li>
              </ol>

              <a on:click={(_e) => initDemo()} class="button is-link">Setup Demo</a>
            </div>
          </div>
        </ZeroState>
      </div>
    </div>
  </div>
</section>

<section class="section tab-networth" class:is-hidden={isEmpty}>
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
                      <nav class="level grid-2">
                        <LevelItem
                          narrow
                          title="Net worth"
                          color={COLORS.primary}
                          value={formatCurrency(networth.balanceAmount)}
                        />

                        <LevelItem
                          narrow
                          title="Net Investment"
                          color={COLORS.secondary}
                          value={formatCurrency(networth.netInvestmentAmount)}
                        />
                      </nav>
                      <nav class="level grid-2">
                        <LevelItem
                          narrow
                          title="Gain / Loss"
                          color={networth.gainAmount >= 0 ? COLORS.gainText : COLORS.lossText}
                          value={formatCurrency(networth.gainAmount)}
                        />

                        <LevelItem
                          narrow
                          title="XIRR"
                          value={formatFloat(xirr)}
                          subtitle="{formatPercentage(networth.absoluteReturn, 2)} absolute return"
                        />
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
                <ZeroState item={cashFlows}>
                  <strong>Oops!</strong> You have not made any transactions in the last 3 months.
                </ZeroState>

                <svg
                  class:is-not-visible={_.isEmpty(cashFlows)}
                  id="d3-current-cash-flow"
                  height="250"
                  width="100%"
                />
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
                        <BudgetCard compact {accountBudget} />
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
              <ZeroState item={selectedExpenses}>
                <strong>Hurray!</strong> You have no expenses this month.
              </ZeroState>
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
