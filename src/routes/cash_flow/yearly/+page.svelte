<script lang="ts">
  import { onMount } from "svelte";
  import _ from "lodash";
  import { renderFlow } from "$lib/cash_flow";
  import { ajax, depth, firstName, type Graph, type Posting } from "$lib/utils";
  import { dateMin, year } from "../../../store";
  import {
    cashflowType,
    setCashflowDepthAllowed,
    cashflowExpenseDepth,
    cashflowIncomeDepth
  } from "../../../persisted_store";
  import ZeroState from "$lib/components/ZeroState.svelte";

  let graph: Record<string, Record<string, Graph>>, expenses: Posting[];
  let isEmpty = false;

  function maxDepth(prefix: string) {
    if (!graph) return 1;
    const max = _.chain(graph)
      .flatMap((g) => g["hierarchy"])
      .flatMap((g) => g.nodes)
      .filter((n) => n.name.startsWith(prefix))
      .map((n) => depth(n.name))
      .max()
      .value();

    return max || 1;
  }

  function filter(graph: Graph, incomeDepth: number, expenseDepth: number, flowType: string) {
    if (!graph || flowType != "hierarchy") return graph;

    const [removed, allowed] = _.partition(graph.nodes, (n) => {
      const account = firstName(n.name);
      if (account === "Income") return depth(n.name) > incomeDepth;
      if (account === "Expenses") return depth(n.name) > expenseDepth;
      return false;
    });

    const removedIds = removed.map((n) => n.id);
    return {
      nodes: allowed,
      links: graph.links.filter(
        (l) => !removedIds.includes(l.source) && !removedIds.includes(l.target)
      )
    };
  }

  $: if (graph) {
    if (graph[$year] == null) {
      isEmpty = true;
    } else {
      renderFlow(
        filter(
          _.cloneDeep(graph[$year][$cashflowType]),
          $cashflowIncomeDepth,
          $cashflowExpenseDepth,
          $cashflowType
        ),
        $cashflowType
      );
      isEmpty = false;
    }
  }

  onMount(async () => {
    ({ expenses, graph } = await ajax("/api/expense"));
    let firstExpense = _.minBy(expenses, (e) => e.date);
    if (firstExpense) {
      dateMin.set(firstExpense.date);
    }

    setCashflowDepthAllowed(maxDepth("Expenses"), maxDepth("Income"));
  });
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box overflow-x-auto">
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
