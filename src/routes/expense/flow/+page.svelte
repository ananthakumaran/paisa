<script lang="ts">
  import { onMount } from "svelte";
  import _ from "lodash";
  import { renderFlow } from "$lib/expense/flow";
  import { ajax, type Graph, type Posting } from "$lib/utils";
  import { dateMin, year } from "../../../store";

  let graph: Record<string, Graph>, expenses: Posting[];

  $: if (graph) {
    renderFlow(graph[$year]);
  }

  onMount(async () => {
    ({ expenses, graph } = await ajax("/api/expense"));
    let firstExpense = _.minBy(expenses, (e) => e.date);
    if (firstExpense) {
      dateMin.set(firstExpense.date);
    }
  });
</script>

<section class="section tab-flow">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <svg id="d3-expense-flow" height={window.innerHeight - 100} />
      </div>
    </div>
  </div>
</section>
