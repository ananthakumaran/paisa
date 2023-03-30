<script lang="ts">
  import { renderPortfolioBreakdown } from "$lib/portfolio";
  import { ajax } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  onMount(async () => {
    const {
      name_and_security_type: nas,
      security_type: st,
      rating: r,
      industry: i
    } = await ajax("/api/portfolio_allocation");

    renderPortfolioBreakdown("#d3-portfolio-security-type", st);
    renderPortfolioBreakdown("#d3-portfolio-security-rating", r);
    renderPortfolioBreakdown("#d3-portfolio-security-industry", i);
    renderPortfolioBreakdown("#d3-portfolio", nas, true);
  });
</script>

<section class="section tab-portfolio">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div id="d3-portfolio-security-type-treemap" style="width: 100%; position: relative" />
        <svg id="d3-portfolio-security-type" width="100%" />
      </div>
    </div>
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div id="d3-portfolio-security-rating-treemap" style="width: 100%; position: relative" />
        <svg id="d3-portfolio-security-rating" width="100%" />
      </div>
    </div>
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div id="d3-portfolio-security-industry-treemap" style="width: 100%; position: relative" />
        <svg id="d3-portfolio-security-industry" width="100%" />
      </div>
    </div>
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div id="d3-portfolio-treemap" style="width: 100%; position: relative" />
        <svg id="d3-portfolio" width="100%" />
      </div>
    </div>
  </div>
</section>
