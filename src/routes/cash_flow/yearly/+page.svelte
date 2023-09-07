<script lang="ts">
  import { onMount } from "svelte";
  import _ from "lodash";
  import { renderFlow } from "$lib/cash_flow";
  import { ajax, type Graph, type Posting } from "$lib/utils";
  import { dateMin, year, cashflowType } from "../../../store";
  import ZeroState from "$lib/components/ZeroState.svelte";

  let graph: Record<string, Record<string, Graph>>, expenses: Posting[];
  let isEmpty = false;

  $: if (graph) {
    if (graph[$year] == null) {
      isEmpty = true;
    } else {
      renderFlow(graph[$year][$cashflowType], $cashflowType);
      isEmpty = false;
    }
  }

  onMount(async () => {
    ({ expenses, graph } = await ajax("/api/expense"));
    let firstExpense = _.minBy(expenses, (e) => e.date);
    if (firstExpense) {
      dateMin.set(firstExpense.date);
    }
  });
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <ZeroState item={!isEmpty}
            ><strong>Oops!</strong> You have not made any transactions for the selected year.</ZeroState
          >
          <svg
            class:is-not-visible={isEmpty}
            id="d3-expense-flow"
            height={window.innerHeight - 100}
          />
        </div>
      </div>
    </div>
  </div>
</section>
