<script lang="ts">
  import {
    renderAllocation,
    renderAllocationTarget,
    renderAllocationTimeline
  } from "$lib/allocation";
  import COLORS, { generateColorScheme } from "$lib/colors";
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import LegendCard from "$lib/components/LegendCard.svelte";
  import Table from "$lib/components/Table.svelte";
  import { accountName, nonZeroCurrency } from "$lib/table_formatters";
  import { ajax, formatPercentage, rem, type Aggregate, type Legend } from "$lib/utils";
  import _ from "lodash";
  import { onMount, tick } from "svelte";
  import type { ColumnDefinition, ProgressBarParams } from "tabulator-tables";
  import { _ as t } from "$lib/i18n";
  import { get } from "svelte/store";

  const tt = get(t);
  let showAllocation = false;
  let depth = 2;
  let allocationTimelineLegends: Legend[] = [];
  let aggregateLeafNodes: Aggregate[] = [];
  let total = 0;

  const columns: ColumnDefinition[] = [
    { title: tt("page.assets.col_account"), field: "account", formatter: accountName },
    {
      title: tt("page.assets.col_market_value"),
      field: "market_amount",
      hozAlign: "right",
      formatter: nonZeroCurrency
    },
    {
      title: tt("page.assets.col_percent"),
      field: "percent",
      hozAlign: "right",
      formatter: (cell) => formatPercentage(cell.getValue() / 100, 2)
    },
    {
      title: "%",
      field: "percent",
      hozAlign: "right",
      formatter: "progress",
      cssClass: "has-text-left",
      minWidth: rem(250),
      formatterParams: {
        color: COLORS.assets,
        min: 0
      }
    }
  ];

  onMount(async () => {
    const {
      aggregates: aggregates,
      aggregates_timeline: aggregatesTimeline,
      allocation_targets: allocationTargets
    } = await ajax("/api/allocation");
    const accounts = _.keys(aggregates);
    // For the table we want one row per real ledger account, not the
    // synthetic root or kind-bucket nodes. Filter to leaves with a value
    // and prefer the human-readable `original_account` for the link
    // formatter — the synthetic `account` path contains '/' separators.
    aggregateLeafNodes = _.chain(_.values(aggregates))
      .filter((a) => a.market_amount > 0 && Boolean(a.original_account))
      .map((a) => ({ ...a, account: a.original_account as string }))
      .value();
    total = _.sumBy(aggregateLeafNodes, (a) => a.market_amount);
    aggregateLeafNodes = _.map(aggregateLeafNodes, (a) => {
      a.percent = (a.market_amount / total) * 100;
      return a;
    });
    const max = _.max(_.map(aggregateLeafNodes, (a) => a.percent)) || 100;
    (_.last(columns).formatterParams as ProgressBarParams).max = max;
    const color = generateColorScheme(accounts);
    depth = _.max(_.map(accounts, (account) => account.split(":").length));

    if (!_.isEmpty(allocationTargets)) {
      showAllocation = true;
    }
    await tick();

    renderAllocationTarget(allocationTargets, color);
    renderAllocation(aggregates, color);
    allocationTimelineLegends = renderAllocationTimeline(aggregatesTimeline);
  });
</script>

<section class="section tab-allocation" style={showAllocation ? "" : "display: none"}>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div class="box overflow-x-auto">
          <div id="d3-allocation-target-treemap" style="width: 100%; position: relative" />
          <svg id="d3-allocation-target" />
        </div>
      </div>
    </div>
    <BoxLabel text={$t("page.assets.allocation_targets")} />
  </div>
</section>
<section class="section tab-allocation">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div id="d3-allocation-category" style="width: 100%; height: {depth * 100}px" />
      </div>
    </div>
    <BoxLabel text={$t("page.assets.allocation_by_category")} />
  </div>
</section>
<section class="section tab-allocation">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div id="d3-allocation-value" style="width: 100%; height: 300px" />
      </div>
    </div>
    <BoxLabel text={$t("page.assets.allocation_by_value")} />
  </div>
</section>
<section class="section tab-allocation">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <LegendCard legends={allocationTimelineLegends} clazz="ml-4" />
          <svg id="d3-allocation-timeline" width="100%" height="300" />
        </div>
      </div>
    </div>
    <BoxLabel text={$t("page.assets.allocation_timeline")} />
  </div>
</section>
<section class="section tab-allocation">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <Table data={aggregateLeafNodes} tree {columns} />
      </div>
    </div>
    <BoxLabel text={$t("page.assets.allocation_table")} />
  </div>
</section>
