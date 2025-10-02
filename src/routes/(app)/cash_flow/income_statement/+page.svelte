<script lang="ts">
  import { onMount } from "svelte";
  import _ from "lodash";
  import dayjs from "dayjs";
  import { renderIncomeStatement } from "$lib/income_statement";
  import {
    ajax,
    formatCurrency,
    formatPercentage,
    restName,
    type IncomeStatement,
    firstName
  } from "$lib/utils";
  import { dateMin, year, viewMode, selectedMonths } from "../../../../store";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import { iconify } from "$lib/icon";

  let isEmpty = false;

  let svg: Element;
  let incomeStatement: IncomeStatement;
  let renderer: (data: IncomeStatement) => void;
  let yearly: Record<string, IncomeStatement> = {};
  let monthly: Record<string, Record<string, IncomeStatement>> = {};
  let diff: number;
  let diffPercent: number;
  let years: string[] = [];

  type AccountGroupName =
    | "income"
    | "interest"
    | "equity"
    | "pnl"
    | "liabilities"
    | "tax"
    | "expenses";
  interface AccountGroup {
    key: AccountGroupName;
    accounts: string[];
    label: string;
    multiplier: number;
  }

  function formatUnlessZero(value: number) {
    if (value != 0) {
      return formatCurrency(value);
    }
    return "";
  }

  function changeClass(value: number) {
    if (value > 0) {
      return "has-text-success";
    } else if (value < 0) {
      return "has-text-danger";
    } else {
      return "has-text-grey";
    }
  }

  const sum = (object: Record<string, number>) => Object.values(object).reduce((a, b) => a + b, 0);

  const accountGroups: AccountGroup[] = [];

  $: if (yearly && monthly && renderer) {
    let currentIncomeStatement: IncomeStatement = null;

    if (isMonthlyView && $selectedMonths.length > 0) {
      // Aggregate data from selected months
      if (monthly[$year]) {
        const aggregatedStatement: IncomeStatement = {
          startingBalance: 0,
          endingBalance: 0,
          date: dayjs(),
          income: {},
          interest: {},
          equity: {},
          pnl: {},
          liabilities: {},
          tax: {},
          expenses: {}
        };

        let hasData = false;
        let firstMonthData: IncomeStatement | null = null;
        let lastMonthData: IncomeStatement | null = null;

        // Sort selected months to get first and last
        const sortedMonths = $selectedMonths.sort();

        for (const month of sortedMonths) {
          const monthKey = `${$year}-${month}`;
          if (monthly[$year][monthKey]) {
            hasData = true;
            const monthData = monthly[$year][monthKey];

            // Set first and last month data for balance calculations
            if (!firstMonthData) {
              firstMonthData = monthData;
            }
            lastMonthData = monthData;

            // Helper function to merge records
            const mergeRecords = (
              target: Record<string, number>,
              source: Record<string, number>
            ) => {
              for (const [account, amount] of Object.entries(source)) {
                target[account] = (target[account] || 0) + amount;
              }
            };

            // Aggregate all account records
            mergeRecords(aggregatedStatement.income, monthData.income);
            mergeRecords(aggregatedStatement.interest, monthData.interest);
            mergeRecords(aggregatedStatement.equity, monthData.equity);
            mergeRecords(aggregatedStatement.pnl, monthData.pnl);
            mergeRecords(aggregatedStatement.liabilities, monthData.liabilities);
            mergeRecords(aggregatedStatement.tax, monthData.tax);
            mergeRecords(aggregatedStatement.expenses, monthData.expenses);
          }
        }

        if (hasData && firstMonthData && lastMonthData) {
          // Use yearly starting balance and calculate ending balance using the same formula as backend
          aggregatedStatement.startingBalance = yearly[$year]?.startingBalance || 0;

          // Calculate ending balance: startingBalance + sum(income)*-1 + sum(interest)*-1 + sum(equity)*-1 + sum(tax)*-1 + sum(expenses)*-1 + sum(pnl) + sum(liabilities)*-1
          aggregatedStatement.endingBalance =
            aggregatedStatement.startingBalance +
            sum(aggregatedStatement.income) * -1 +
            sum(aggregatedStatement.interest) * -1 +
            sum(aggregatedStatement.equity) * -1 +
            sum(aggregatedStatement.tax) * -1 +
            sum(aggregatedStatement.expenses) * -1 +
            sum(aggregatedStatement.pnl) +
            sum(aggregatedStatement.liabilities) * -1;

          currentIncomeStatement = aggregatedStatement;
        }
      }
    } else {
      // Show yearly data
      currentIncomeStatement = yearly[$year];
    }

    if (currentIncomeStatement) {
      incomeStatement = currentIncomeStatement;
      years = _.sortBy(_.keys(yearly)).reverse();
      diff = incomeStatement.endingBalance - incomeStatement.startingBalance;
      diffPercent = diff / incomeStatement.startingBalance;

      renderer(incomeStatement);
      isEmpty = false;
    } else {
      incomeStatement = null;
      isEmpty = true;
    }
  }

  // Reactive statement to determine if we're in monthly view
  $: isMonthlyView = $viewMode === "monthly";

  // Get available months for the selected year
  $: availableMonths = monthly && monthly[$year] ? Object.keys(monthly[$year]).sort() : [];

  function uniqueAccounts(statements: IncomeStatement[], key: AccountGroupName) {
    const accounts = new Set<string>();
    for (const statement of statements) {
      for (const account of _.keys(statement[key])) {
        accounts.add(account);
      }
    }
    return Array.from(accounts).sort();
  }

  onMount(async () => {
    const response = await ajax("/api/income_statement");
    yearly = response.yearly;
    monthly = response.monthly;
    const y = _.minBy(_.values(yearly), (y) => y.date);
    renderer = renderIncomeStatement(svg);
    if (y) {
      dateMin.set(y.date);
    }

    accountGroups.push({
      key: "income",
      accounts: uniqueAccounts(_.values(yearly), "income"),
      label: "Income",
      multiplier: -1
    });

    accountGroups.push({
      key: "tax",
      accounts: uniqueAccounts(_.values(yearly), "tax"),
      label: "Tax",
      multiplier: -1
    });

    accountGroups.push({
      key: "interest",
      accounts: uniqueAccounts(_.values(yearly), "interest"),
      label: "Interest",
      multiplier: -1
    });

    accountGroups.push({
      key: "pnl",
      accounts: uniqueAccounts(_.values(yearly), "pnl"),
      label: "Gain / Loss",
      multiplier: 1
    });

    accountGroups.push({
      key: "equity",
      accounts: uniqueAccounts(_.values(yearly), "equity"),
      label: "Equity",
      multiplier: -1
    });

    accountGroups.push({
      key: "liabilities",
      accounts: uniqueAccounts(_.values(yearly), "liabilities"),
      label: "Liabilities",
      multiplier: -1
    });

    accountGroups.push({
      key: "expenses",
      accounts: uniqueAccounts(_.values(yearly), "expenses"),
      label: "Expenses",
      multiplier: -1
    });
  });
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns flex-wrap">
      {#if incomeStatement}
        <div class="column is-12">
          <div class="box py-2 my-0 overflow-x-auto">
            <div class="is-flex mr-2 is-align-items-baseline" style="min-width: fit-content">
              <div class="ml-3 custom-icon is-size-5 whitespace-nowrap">
                {$year}
              </div>
              <div class="ml-3 whitespace-nowrap">
                <span class="mr-1 is-size-7 has-text-grey">Start</span>
                <span class="has-text-weight-bold"
                  >{formatCurrency(incomeStatement.startingBalance)}</span
                >
              </div>
              <div class="ml-3 whitespace-nowrap">
                <span class="mr-1 is-size-7 has-text-grey">End</span>
                <span class="has-text-weight-bold"
                  >{formatCurrency(incomeStatement.endingBalance)}</span
                >
              </div>

              <div class="ml-3 whitespace-nowrap">
                <span class="mr-1 is-size-7 has-text-grey">change</span>
                <span class="has-text-weight-bold {changeClass(diff)}">{formatCurrency(diff)}</span>
                <span class="mr-1 is-size-7 has-text-weight-bold {changeClass(diff)}"
                  >{formatPercentage(diffPercent, 2)}</span
                >
              </div>
            </div>
          </div>
        </div>
      {/if}
      <div class="column is-12">
        <div class="box overflow-x-auto">
          <ZeroState item={!isEmpty}
            ><strong>Oops!</strong> You have not made any transactions for the selected year.</ZeroState
          >

          <svg class:is-not-visible={isEmpty} bind:this={svg}></svg>
        </div>
      </div>
    </div>
  </div>
</section>

<section class="section py-0">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 pb-0">
        <div class="box pt-0 overflow-x-auto max-h-screen max-w-fit">
          <table
            class="table is-narrow is-hoverable is-light-border has-sticky-header has-sticky-column"
          >
            <thead>
              <tr>
                <th class="py-2">Account</th>
                {#if !isMonthlyView}
                  {#each years as y}
                    <th class="py-2 has-text-right">{y}</th>
                  {/each}
                {:else}
                  <th class="py-2 has-text-right">Total</th>
                  {#each $selectedMonths.sort() as month}
                    <th class="py-2 has-text-right">{$year}-{month}</th>
                  {/each}
                {/if}
              </tr>
            </thead>
            <tbody class="has-text-grey-dark">
              {#each accountGroups as group}
                <tr class="has-text-weight-bold is-sub-header">
                  <th>{group.label}</th>
                  {#if !isMonthlyView}
                    {#each years as y}
                      <td class="has-text-right">
                        {#if yearly[y]?.[group.key]}
                          {formatUnlessZero(sum(yearly[y][group.key]) * group.multiplier)}
                        {/if}
                      </td>
                    {/each}
                  {:else}
                    <td class="has-text-right">
                      {#if incomeStatement?.[group.key]}
                        {formatUnlessZero(sum(incomeStatement[group.key]) * group.multiplier)}
                      {/if}
                    </td>
                    {#each $selectedMonths.sort() as month}
                      <td class="has-text-right">
                        {#if monthly[$year]?.[`${$year}-${month}`]?.[group.key]}
                          {formatUnlessZero(
                            sum(monthly[$year][`${$year}-${month}`][group.key]) * group.multiplier
                          )}
                        {/if}
                      </td>
                    {/each}
                  {/if}
                </tr>
                {#each group.accounts as account}
                  <tr>
                    <th class="custom-icon whitespace-nowrap"
                      ><span class="pl-3 has-text-weight-normal"
                        >{iconify(restName(account), { group: firstName(account) })}</span
                      ></th
                    >
                    {#if !isMonthlyView}
                      {#each years as y}
                        <td class="has-text-right">
                          {#if yearly[y]?.[group.key]?.[account]}
                            {formatUnlessZero(yearly[y][group.key][account] * group.multiplier)}
                          {/if}
                        </td>
                      {/each}
                    {:else}
                      <td class="has-text-right">
                        {#if incomeStatement?.[group.key]?.[account]}
                          {formatUnlessZero(incomeStatement[group.key][account] * group.multiplier)}
                        {/if}
                      </td>
                      {#each $selectedMonths.sort() as month}
                        <td class="has-text-right">
                          {#if monthly[$year]?.[`${$year}-${month}`]?.[group.key]?.[account]}
                            {formatUnlessZero(
                              monthly[$year][`${$year}-${month}`][group.key][account] *
                                group.multiplier
                            )}
                          {/if}
                        </td>
                      {/each}
                    {/if}
                  </tr>
                {/each}
                <tr
                  ><td colspan={!isMonthlyView ? years.length + 1 : $selectedMonths.length + 2}
                    >&nbsp;</td
                  ></tr
                >
              {/each}

              <tr class="has-text-weight-bold">
                <th>Change</th>
                {#if !isMonthlyView}
                  {#each years as y}
                    {#if yearly[y]}
                      {@const diff = yearly[y].endingBalance - yearly[y].startingBalance}
                      <td class="has-text-right {changeClass(diff)}">
                        <div>{formatCurrency(diff)}</div>
                        <div>{formatPercentage(diff / yearly[y].startingBalance)}</div>
                      </td>
                    {:else}
                      <td></td>
                    {/if}
                  {/each}
                {:else}
                  {#if incomeStatement}
                    {@const monthlyDiff =
                      incomeStatement.endingBalance - incomeStatement.startingBalance}
                    <td class="has-text-right {changeClass(monthlyDiff)}">
                      <div>{formatCurrency(monthlyDiff)}</div>
                      <div>{formatPercentage(monthlyDiff / incomeStatement.startingBalance)}</div>
                    </td>
                  {:else}
                    <td></td>
                  {/if}
                  {#each $selectedMonths.sort() as month}
                    {#if monthly[$year]?.[`${$year}-${month}`]}
                      {@const monthData = monthly[$year][`${$year}-${month}`]}
                      {@const diff = monthData.endingBalance - monthData.startingBalance}
                      <td class="has-text-right {changeClass(diff)}">
                        <div>{formatCurrency(diff)}</div>
                        <div>{formatPercentage(diff / monthData.startingBalance)}</div>
                      </td>
                    {:else}
                      <td></td>
                    {/if}
                  {/each}
                {/if}
              </tr>
              <tr class="has-text-weight-bold">
                <th>End Balance</th>
                {#if !isMonthlyView}
                  {#each years as y}
                    <td class="has-text-right">
                      {#if yearly[y]}
                        {formatCurrency(yearly[y].endingBalance)}
                      {/if}
                    </td>
                  {/each}
                {:else}
                  <td class="has-text-right">
                    {#if incomeStatement}
                      {formatCurrency(incomeStatement.endingBalance)}
                    {/if}
                  </td>
                  {#each $selectedMonths.sort() as month}
                    <td class="has-text-right">
                      {#if monthly[$year]?.[`${$year}-${month}`]}
                        {formatCurrency(monthly[$year][`${$year}-${month}`].endingBalance)}
                      {/if}
                    </td>
                  {/each}
                {/if}
              </tr>
              <tr class="has-text-weight-bold">
                <th>Start Balance</th>
                {#if !isMonthlyView}
                  {#each years as y}
                    <td class="has-text-right">
                      {#if yearly[y]}
                        {formatCurrency(yearly[y].startingBalance)}
                      {/if}
                    </td>
                  {/each}
                {:else}
                  <td class="has-text-right">
                    {#if incomeStatement}
                      {formatCurrency(incomeStatement.startingBalance)}
                    {/if}
                  </td>
                  {#each $selectedMonths.sort() as month}
                    <td class="has-text-right">
                      {#if monthly[$year]?.[`${$year}-${month}`]}
                        {formatCurrency(monthly[$year][`${$year}-${month}`].startingBalance)}
                      {/if}
                    </td>
                  {/each}
                {/if}
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</section>
