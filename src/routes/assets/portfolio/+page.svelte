<script lang="ts">
  import { renderPortfolioBreakdown } from "$lib/portfolio";
  import { ajax } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  onMount(async () => {
    const { portfolio_aggregates: pas } = await ajax("/api/portfolio_allocation");
    renderPortfolioBreakdown(pas);
  });
</script>

<section class="section tab-portfolio">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div id="d3-portfolio-treemap" style="width: 100%; position: relative" />
        <svg id="d3-portfolio" width="100%" />
      </div>
    </div>
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div>
          <p class="heading">Portfolio Breakdown</p>
        </div>
      </div>
    </div>
  </div>
</section>
