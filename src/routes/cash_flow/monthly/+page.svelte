<script lang="ts">
  import _ from "lodash";
  import { renderMonthlyFlow } from "$lib/cash_flow";
  import { ajax, type CashFlow } from "$lib/utils";
  import { onMount } from "svelte";
  import { dateRange, setAllowedDateRange } from "../../../store";

  let cashFlows: CashFlow[] = [];
  let renderer: (cashflows: CashFlow[]) => void;

  $: if (!_.isEmpty(cashFlows)) {
    renderer(
      _.filter(
        cashFlows,
        (c) => c.date.isSameOrBefore($dateRange.to) && c.date.isAfter($dateRange.from)
      )
    );
  }

  onMount(async () => {
    ({ cash_flows: cashFlows } = await ajax("/api/cash_flow"));
    setAllowedDateRange(_.map(cashFlows, (c) => c.date));
    renderer = renderMonthlyFlow("#d3-monthly-cash-flow", {
      rotate: true,
      balance: _.last(cashFlows)?.balance || 0
    });
  });
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <svg id="d3-monthly-cash-flow" width="100%" height="500" />
        </div>
      </div>
    </div>
  </div>
</section>
