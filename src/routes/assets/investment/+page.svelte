<script lang="ts">
  import {
    renderMonthlyInvestmentTimeline,
    renderYearlyCards,
    renderYearlyInvestmentTimeline
  } from "$lib/investment";
  import { ajax } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  onMount(async () => {
    const { assets: assets, yearly_cards: yearlyCards } = await ajax("/api/investment");
    renderMonthlyInvestmentTimeline(assets);
    renderYearlyInvestmentTimeline(yearlyCards);
    renderYearlyCards(yearlyCards);
  });
</script>

<section class="section tab-investment">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <svg id="d3-investment-timeline" width="100%" height="500" />
      </div>
    </div>
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div>
          <p class="heading">Monthly Investment Timeline</p>
        </div>
      </div>
    </div>
  </div>
</section>
<section class="section tab-investment">
  <div class="container is-fluid">
    <div class="columns is-flex-wrap-wrap">
      <div class="column is-full-tablet is-half-fullhd">
        <div class="py-3">
          <svg id="d3-yearly-investment-timeline" width="100%" />
        </div>
        <div class="py-3 has-text-centered">
          <p class="heading">Financial Year Investment Timeline</p>
        </div>
      </div>
      <div class="column is-full-tablet is-half-fullhd">
        <div class="columns is-flex-wrap-wrap" id="d3-yearly-investment-cards" />
      </div>
    </div>
  </div>
</section>
