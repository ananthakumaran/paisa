<script lang="ts">
  import * as cashFlow from "$lib/cash_flow";
  import COLORS from "$lib/colors";
  import LastNMonths from "$lib/components/LastNMonths.svelte";
  import TransactionCard from "$lib/components/TransactionCard.svelte";
  import * as expense from "$lib/expense/monthly";
  import { enrichTrantionSequence, sortTrantionSequence } from "$lib/transaction_sequence";
  import {
    ajax,
    formatCurrency,
    formatFloat,
    type Budget,
    type CashFlow,
    type Networth,
    type Posting,
    type Transaction,
    type TransactionSequence,
    type Legend,
    now,
    type GoalSummary,
    type AssetBreakdown
  } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  import BudgetCard from "$lib/components/BudgetCard.svelte";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import { MasonryGrid } from "@egjs/svelte-grid";
  import { refresh } from "../../store";
  import UpcomingCard from "$lib/components/UpcomingCard.svelte";
  import GoalSummaryCard from "$lib/components/GoalSummaryCard.svelte";
  import LegendCard from "$lib/components/LegendCard.svelte";
  import BalanceCard from "$lib/components/BalanceCard.svelte";
  import { _ as t } from "$lib/i18n";

  let UntypedMasonryGrid = MasonryGrid as any;

  let cashflowLegends: Legend[] = [];
  let month = now().format("YYYY-MM");
  let goalSummaries: GoalSummary[] = [];
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
  let checkingBalances: Record<string, AssetBreakdown> = {};

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
      goalSummaries,
      budget: { budgetsByMonth },
      transactionSequences,
      networth: { networth, xirr },
      checkingBalances: { asset_breakdowns: checkingBalances },
      transactions
    } = await ajax("/api/dashboard"));

    goalSummaries = _.sortBy(goalSummaries, (g) => -g.priority);

    if (_.isEmpty(transactions)) {
      isEmpty = true;
    } else {
      isEmpty = false;
    }

    const postings = _.chain(expenses).values().flatten().value();
    const z = expense.colorScale(postings);
    renderer = expense.renderCurrentExpensesBreakdown(z);
    currentBudget = budgetsByMonth[month];

    const { renderer: cashflowRenderer, legends } = cashFlow.renderMonthlyFlow(
      "#d3-current-cash-flow",
      {
        rotate: false,
        balance: _.last(cashFlows)?.balance || 0
      }
    );
    cashflowRenderer(cashFlows);
    cashflowLegends = legends;
    transactionSequences = _.take(
      sortTrantionSequence(enrichTrantionSequence(transactionSequences)),
      16
    );
  });
</script>

<section class="section" class:is-hidden={!isEmpty}>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <ZeroState item={!isEmpty}>
          <div class="has-text-left" style="max-width: 640px;">
            <p class="mb-2">
              {$t("page.dashboard.new_user_intro")}
            </p>
            <div>
              <p class="is-size-4">{$t("page.dashboard.get_started_title")}</p>
              <ol class="ml-5 mt-2 mb-4">
                <li>
                  {$t("page.dashboard.get_started_step1_pre")}<a href="/more/config"
                    >{$t("page.dashboard.configuration_link")}</a
                  >{$t("page.dashboard.get_started_step1_post")}
                </li>
                <li>
                  {$t("page.dashboard.get_started_step2_pre")}<a href="/ledger/editor"
                    >{$t("page.dashboard.editor_link")}</a
                  >{$t("page.dashboard.get_started_step2_post")}
                </li>
              </ol>
              <p class="is-size-4">{$t("page.dashboard.demo_title")}</p>
              <p class="ml-3"></p>
              <ol class="ml-5 mt-2 mb-4">
                <li>
                  {$t("page.dashboard.demo_step1")}
                </li>
                <li>
                  {$t("page.dashboard.demo_step2_pre")}<a href="/ledger/editor"
                    >{$t("page.dashboard.editor_link")}</a
                  >{$t("page.dashboard.demo_step2_post")}
                </li>
                <li>
                  {$t("page.dashboard.demo_step3_pre")}<a href="/more/config"
                    >{$t("page.dashboard.configuration_link")}</a
                  >{$t("page.dashboard.demo_step3_post")}
                </li>
              </ol>

              <a on:click={(_e) => initDemo()} class="button is-link"
                >{$t("page.dashboard.setup_demo")}</a
              >
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
      <div class="tile is-4 is-vertical">
        <div class="tile is-parent">
          <div class="tile is-child">
            <div class="content">
              <p class="subtitle">
                <a class="secondary-link has-text-grey" href="/assets/networth"
                  >{$t("page.dashboard.section_assets")}</a
                >
              </p>
              <div class="content">
                <div>
                  {#if networth}
                    <nav class="level grid-2">
                      <LevelItem
                        narrow
                        title={$t("page.dashboard.net_worth")}
                        color={COLORS.primary}
                        value={formatCurrency(networth.balanceAmount)}
                      />

                      <LevelItem
                        narrow
                        title={$t("page.dashboard.net_investment")}
                        color={COLORS.secondary}
                        value={formatCurrency(networth.netInvestmentAmount)}
                      />
                    </nav>
                    <nav class="level grid-2">
                      <LevelItem
                        narrow
                        title={$t("page.dashboard.gain_loss")}
                        color={networth.gainAmount >= 0 ? COLORS.gainText : COLORS.lossText}
                        value={formatCurrency(networth.gainAmount)}
                      />

                      <LevelItem
                        narrow
                        title={$t("page.dashboard.xirr")}
                        value={formatFloat(xirr)}
                      />
                    </nav>
                  {/if}
                </div>
              </div>
            </div>
          </div>
        </div>

        {#if !_.isEmpty(checkingBalances)}
          <div class="tile is-parent">
            <article class="tile is-child">
              <div class="content">
                <p class="subtitle">
                  <a class="secondary-link has-text-grey" href="/assets/balance"
                    >{$t("page.dashboard.section_checking_balance")}</a
                  >
                </p>
                <div class="content">
                  <UntypedMasonryGrid gap={10} maxStretchColumnSize={400} align="stretch">
                    {#each _.values(checkingBalances) as assetBreakdown}
                      <div class="is-flex-grow-1">
                        <BalanceCard {assetBreakdown} />
                      </div>
                    {/each}
                  </UntypedMasonryGrid>
                </div>
              </div>
            </article>
          </div>
        {/if}

        <div class="tile is-parent">
          <article class="tile is-child min-w-0">
            <p class="subtitle">
              <a class="secondary-link has-text-grey" href="/cash_flow/monthly"
                >{$t("page.dashboard.section_cash_flow")}</a
              >
            </p>
            <div class="content box px-2 pb-0">
              <ZeroState item={cashFlows}>
                <strong>{$t("common.oops")}</strong>
                {$t("page.dashboard.no_cash_flow_recent")}
              </ZeroState>

              <LegendCard legends={cashflowLegends} clazz="mb-2 overflow-x-auto" />

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
                  <a class="secondary-link has-text-grey" href="/expense/budget"
                    >{$t("page.dashboard.section_budget")}</a
                  >
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
        {#if !_.isEmpty(goalSummaries)}
          <div class="tile">
            <div class="tile is-parent is-12">
              <article class="tile is-child">
                <div class="content">
                  <p class="subtitle">
                    <a class="secondary-link has-text-grey" href="/more/goals"
                      >{$t("page.dashboard.section_goals")}</a
                    >
                  </p>
                  <div class="content">
                    {#each goalSummaries as goal}
                      <GoalSummaryCard {goal} small />
                    {/each}
                  </div>
                </div>
              </article>
            </div>
          </div>
        {/if}
      </div>
      <div class="tile is-vertical">
        <div class="tile is-parent is-12">
          <article class="tile is-child">
            <p class="subtitle is-flex is-justify-content-space-between is-align-items-end">
              <span
                ><a class="secondary-link has-text-grey" href="/expense/monthly"
                  >{$t("page.dashboard.section_expenses")}</a
                >
                <span class="is-size-5 has-text-weight-bold px-2" style="color: {COLORS.expenses}"
                  >{formatCurrency(totalExpense)}</span
                ></span
              >
              <LastNMonths n={3} bind:value={month} />
            </p>
            <div class="content box px-3">
              <ZeroState item={selectedExpenses}>
                <strong>{$t("common.hurray")}</strong>
                {$t("page.dashboard.no_expenses_this_month")}
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
                    <a class="secondary-link has-text-grey" href="/cash_flow/recurring"
                      >{$t("page.dashboard.section_recurring")}</a
                    >
                  </p>
                  <div class="content box">
                    <div
                      class="grid grid-rows-1 overflow-hidden"
                      style="grid-auto-rows: 0px; grid-template-columns: repeat(auto-fit, minmax(130px, 150px));"
                    >
                      {#each transactionSequences as ts (ts)}
                        <UpcomingCard transactionSequece={ts} />
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
                    <a class="secondary-link has-text-grey" href="/ledger/transaction"
                      >{$t("page.dashboard.section_recent_transactions")}</a
                    >
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
    margin-bottom: 0.5rem !important;
  }

  p.subtitle a.secondary-link {
    text-transform: uppercase;
    font-size: 1rem;
  }
</style>
