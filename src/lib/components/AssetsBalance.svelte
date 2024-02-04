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

  export let breakdowns: Record<string, AssetBreakdown>;
  export let indent = true;

  const columns: ColumnDefinition[] = [
    {
      title: "Account",
      field: "group",
      formatter: indent ? indendedAssetAccountName : accountName,
      frozen: true
    },
    {
      title: "Investment Amount",
      field: "investmentAmount",
      hozAlign: "right",
      vertAlign: "middle",
      formatter: nonZeroCurrency
    },
    {
      title: "Withdrawal Amount",
      field: "withdrawalAmount",
      hozAlign: "right",
      formatter: nonZeroCurrency
    },
    {
      title: "Balance Units",
      field: "balanceUnits",
      hozAlign: "right",
      formatter: nonZeroCurrency
    },
    { title: "Market Value", field: "marketAmount", hozAlign: "right", formatter: nonZeroCurrency },
    { title: "Change", field: "gainAmount", hozAlign: "right", formatter: formatCurrencyChange },
    { title: "XIRR", field: "xirr", hozAlign: "right", formatter: nonZeroFloatChange },
    {
      title: "Absolute Return",
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
