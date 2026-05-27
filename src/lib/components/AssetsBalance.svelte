<script lang="ts">
  import { type AssetBreakdown, buildTree } from "$lib/utils";
  import _ from "lodash";
  import Table from "./Table.svelte";
  import type { ColumnDefinition } from "tabulator-tables";
  import {
    accountName,
    formatCurrencyChange,
    indendedAssetAccountName,
    nonZeroCurrency,
    nonZeroFloatChange,
    nonZeroPercentageChange
  } from "$lib/table_formatters";
  import { _ as t } from "$lib/i18n";
  import { get } from "svelte/store";

  const tt = get(t);

  export let breakdowns: Record<string, AssetBreakdown>;
  export let indent = true;

  const columns: ColumnDefinition[] = [
    {
      title: tt("component.asset_breakdown.col_account"),
      field: "group",
      formatter: indent ? indendedAssetAccountName : accountName,
      frozen: true
    },
    {
      title: tt("component.asset_breakdown.col_investment_amount"),
      field: "investmentAmount",
      hozAlign: "right",
      vertAlign: "middle",
      formatter: nonZeroCurrency
    },
    {
      title: tt("component.asset_breakdown.col_withdrawal_amount"),
      field: "withdrawalAmount",
      hozAlign: "right",
      formatter: nonZeroCurrency
    },
    {
      title: tt("component.asset_breakdown.col_balance_units"),
      field: "balanceUnits",
      hozAlign: "right",
      formatter: nonZeroCurrency
    },
    {
      title: tt("component.asset_breakdown.col_market_value"),
      field: "marketAmount",
      hozAlign: "right",
      formatter: nonZeroCurrency
    },
    {
      title: tt("component.asset_breakdown.col_change"),
      field: "gainAmount",
      hozAlign: "right",
      formatter: formatCurrencyChange
    },
    {
      title: tt("component.asset_breakdown.col_xirr"),
      field: "xirr",
      hozAlign: "right",
      formatter: nonZeroFloatChange
    },
    {
      title: tt("component.asset_breakdown.col_absolute_return"),
      field: "absoluteReturn",
      hozAlign: "right",
      formatter: nonZeroPercentageChange
    }
  ];

  let tree: AssetBreakdown[] = [];
  $: if (breakdowns) {
    tree = buildTree(Object.values(breakdowns), (i) => i.group);
  }
</script>

{#if indent}
  <Table data={tree} tree {columns} />
{:else}
  <Table data={Object.values(breakdowns)} {columns} />
{/if}
