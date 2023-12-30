<script lang="ts">
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import LegendCard from "$lib/components/LegendCard.svelte";
  import {
    renderMonthlyInvestmentTimeline,
    renderYearlyCards,
    renderYearlyInvestmentTimeline
  } from "$lib/investment";
  import { ajax, type Legend } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  let monthlyInvestmentTimelineLegends: Legend[] = [];
  let yearlyInvestmentTimelineLegends: Legend[] = [];

  onMount(async () => {
    const { assets: assets, yearly_cards: yearlyCards } = await ajax("/api/investment");
    monthlyInvestmentTimelineLegends = renderMonthlyInvestmentTimeline(assets);
    yearlyInvestmentTimelineLegends = renderYearlyInvestmentTimeline(yearlyCards);
    renderYearlyCards(yearlyCards);
  });
</script>

<section class="section tab-investment">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <LegendCard legends={monthlyInvestmentTimelineLegends} clazz="ml-4" />
          <svg id="d3-investment-timeline" width="100%" height="500" />
        </div>
      </div>
    </div>
    <BoxLabel text="Monthly Investment Timeline" />
  </div>
</section>
<section class="section tab-investment">
  <div class="container is-fluid">
    <div class="columns is-flex-wrap-wrap">
      <div class="column is-full-tablet is-half-fullhd">
        <div class="box px-2">
          <LegendCard legends={yearlyInvestmentTimelineLegends} clazz="ml-4" />
          <svg id="d3-yearly-investment-timeline" width="100%" />
        </div>
        <BoxLabel text="Financial Year Investment Timeline" />
      </div>
      <div class="column is-full-tablet is-half-fullhd">
        <div class="columns is-flex-wrap-wrap" id="d3-yearly-investment-cards" />
      </div>
    </div>
  </div>
</section>
