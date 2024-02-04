<script lang="ts">
  import Table from "$lib/components/Table.svelte";
  import {
    indendedLiabilityAccountName,
    nonZeroCurrency,
    nonZeroFloatChange
  } from "$lib/table_formatters";
  import { ajax, buildTree, type LiabilityBreakdown } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  import type { ColumnDefinition } from "tabulator-tables";

  let breakdowns: LiabilityBreakdown[] = [];
  let isEmpty = false;

  onMount(async () => {
    ({ liability_breakdowns: breakdowns } = await ajax("/api/liabilities/balance"));

    if (_.isEmpty(breakdowns)) {
      isEmpty = true;
    }
  });

  const columns: ColumnDefinition[] = [
    {
      title: "Account",
      field: "group",
      formatter: indendedLiabilityAccountName,
      frozen: true
    },
    {
      title: "Drawn Amount",
      field: "drawn_amount",
      hozAlign: "right",
      vertAlign: "middle",
      formatter: nonZeroCurrency
    },
    {
      title: "Repaid Amount",
      field: "repaid_amount",
      hozAlign: "right",
      formatter: nonZeroCurrency
    },
    {
      title: "Balance Amount",
      field: "balance_amount",
      hozAlign: "right",
      formatter: nonZeroCurrency
    },
    {
      title: "Interest",
      field: "interest_amount",
      hozAlign: "right",
      formatter: nonZeroCurrency
    },
    { title: "APR", field: "apr", hozAlign: "right", formatter: nonZeroFloatChange }
  ];

  let tree: LiabilityBreakdown[] = [];
  $: if (breakdowns) {
    tree = buildTree(Object.values(breakdowns), (i) => i.group);
  }
</script>

<section class="section" class:is-hidden={!isEmpty}>
  <div class="container is-fluid">
    <div class="columns is-centered">
      <div class="column is-4 has-text-centered">
        <article class="message">
          <div class="message-body">
            <strong>Hurray!</strong> You have no liabilities.
          </div>
        </article>
      </div>
    </div>
  </div>
</section>

<section class="section pb-0" class:is-hidden={isEmpty}>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 pb-0">
        <Table data={tree} tree {columns} />
      </div>
    </div>
  </div>
</section>
