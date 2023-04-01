<script lang="ts">
  import { filterCommodityBreakdowns, renderPortfolioBreakdown } from "$lib/portfolio";
  import { ajax, generateColorScheme, type PortfolioAggregate } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  let commodities: string[] = [];
  let selectedCommodities: string[] = [];
  let security_type: PortfolioAggregate[] = [];
  let name_and_security_type: PortfolioAggregate[] = [];
  let rating: PortfolioAggregate[] = [];
  let industry: PortfolioAggregate[] = [];

  let color: any;

  let securityTypeR: any,
    portfolioR: any,
    industryR: any,
    ratingR: any = null;

  onMount(async () => {
    ({ name_and_security_type, security_type, rating, industry, commodities } = await ajax(
      "/api/portfolio_allocation"
    ));

    selectedCommodities = [...commodities];
    securityTypeR = renderPortfolioBreakdown("#d3-portfolio-security-type", security_type);
    ratingR = renderPortfolioBreakdown("#d3-portfolio-security-rating", rating);
    industryR = renderPortfolioBreakdown("#d3-portfolio-security-industry", industry);
    portfolioR = renderPortfolioBreakdown("#d3-portfolio", name_and_security_type, true);
    color = generateColorScheme(commodities);
  });

  $: if (securityTypeR) {
    securityTypeR(filterCommodityBreakdowns(security_type, selectedCommodities), color);
    ratingR(filterCommodityBreakdowns(rating, selectedCommodities), color);
    industryR(filterCommodityBreakdowns(industry, selectedCommodities), color);
    portfolioR(filterCommodityBreakdowns(name_and_security_type, selectedCommodities), color);
  }
</script>

<section class="section tab-portfolio">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 is-flex">
        {#each commodities as commodity}
          {@const name = `switch-${commodity}`}
          <div class="field mr-5 color-switch" style="--color: {color(commodity)}">
            <input
              id={name}
              type="checkbox"
              bind:group={selectedCommodities}
              name="commodities"
              class="switch is-rounded"
              value={commodity}
            />
            <label for={name}>{commodity}</label>
          </div>
        {/each}
      </div>
    </div>
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

<style lang="scss">
  .color-switch {
    .switch[type="checkbox"]:checked + label::before,
    .switch[type="checkbox"]:checked + label:before {
      background: var(--color);
    }
  }
</style>
