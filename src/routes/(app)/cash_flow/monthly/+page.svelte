<script lang="ts">
  import _ from "lodash";
  import { renderMonthlyFlow } from "$lib/cash_flow";
  import { ajax, type CashFlow } from "$lib/utils";
  import { onMount } from "svelte";
  import { dateRange, setAllowedDateRange } from "../../../../store";
  import ZeroState from "$lib/components/ZeroState.svelte";

  let cashFlows: CashFlow[] = [];
  let renderer: (cashflows: CashFlow[]) => void;

  $: if (!_.isEmpty(cashFlows)) {
    renderer(
      _.filter(
        cashFlows,
        (c) => c.date.isSameOrBefore($dateRange.to) && c.date.isSameOrAfter($dateRange.from)
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
          <ZeroState item={cashFlows}>
            <strong>Oops!</strong> You have not made any transactions.
          </ZeroState>

          <svg
            class:is-not-visible={_.isEmpty(cashFlows)}
            id="d3-monthly-cash-flow"
            width="100%"
            height="500"
          />
        </div>
      </div>
    </div>
  </div>
</section>
