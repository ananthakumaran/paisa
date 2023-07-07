<script lang="ts">
  import { onMount } from "svelte";
  import _ from "lodash";
  import { renderFlow } from "$lib/expense/flow";
  import { ajax, type Graph, type Posting } from "$lib/utils";
  import { dateMin, year } from "../../../store";

  let graph: Record<string, Graph>, expenses: Posting[];
  let isEmpty = false;

  $: if (graph) {
    renderFlow(graph[$year]);
    if (graph[$year] == null) {
      isEmpty = true;
    } else {
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

<section class="section" class:is-hidden={!isEmpty}>
  <div class="container is-fluid">
    <div class="columns is-centered">
      <div class="column is-4 has-text-centered">
        <article class="message is-warning">
          <div class="message-body">No transactions found for the selected year.</div>
        </article>
      </div>
    </div>
  </div>
</section>

<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <svg id="d3-expense-flow" height={window.innerHeight - 100} />
        </div>
      </div>
    </div>
  </div>
</section>
