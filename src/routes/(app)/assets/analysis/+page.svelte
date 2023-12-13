<script lang="ts">
  import { generateColorScheme, genericBarColor } from "$lib/colors";
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import { filterCommodityBreakdowns, renderPortfolioBreakdown } from "$lib/portfolio";
  import { ajax, type PortfolioAggregate } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  let commodities: string[] = [];
  let selectedCommodities: string[] = [];
  let security_type: PortfolioAggregate[] = [];
  let name_and_security_type: PortfolioAggregate[] = [];
  let rating: PortfolioAggregate[] = [];
  let industry: PortfolioAggregate[] = [];
  let isEmpty = false;
  let color: any;

  let securityTypeR: any,
    portfolioR: any,
    industryR: any,
    ratingR: any = null;

  onMount(async () => {
    ({ name_and_security_type, security_type, rating, industry, commodities } = await ajax(
      "/api/portfolio_allocation"
    ));

    if (_.isEmpty(commodities)) {
      isEmpty = true;
      return;
    } else {
      isEmpty = false;
    }

    selectedCommodities = [...commodities];
    securityTypeR = renderPortfolioBreakdown("#d3-portfolio-security-type", security_type);
    ratingR = renderPortfolioBreakdown("#d3-portfolio-security-rating", rating);
    industryR = renderPortfolioBreakdown("#d3-portfolio-security-industry", industry, {
      z: [genericBarColor()]
    });
    portfolioR = renderPortfolioBreakdown("#d3-portfolio", name_and_security_type, {
      showLegend: true
    });
    color = generateColorScheme(commodities);
  });

  $: if (securityTypeR) {
    securityTypeR(filterCommodityBreakdowns(security_type, selectedCommodities), color);
    ratingR(filterCommodityBreakdowns(rating, selectedCommodities), color);
    industryR(filterCommodityBreakdowns(industry, selectedCommodities), color);
    portfolioR(filterCommodityBreakdowns(name_and_security_type, selectedCommodities), color);
  }
</script>

<section class="section tab-interest" class:is-hidden={!isEmpty}>
  <div class="container is-fluid">
    <div class="columns is-centered">
      <div class="column is-4 has-text-centered">
        <article class="message">
          <div class="message-body">
            <strong>Oops!</strong> Looks like mutual fund portfolio data is not available<br /><br
            />
            Use the <strong>Update Mutual Fund Portfolios</strong> menu option at the right corner to
            update the data.
          </div>
        </article>
      </div>
    </div>
  </div>
</section>

<section class="section tab-portfolio" class:is-hidden={isEmpty}>
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
        <div class="box overflow-x-auto">
          <div id="d3-portfolio-security-type-treemap" style="width: 100%; position: relative" />
          <svg id="d3-portfolio-security-type" />
        </div>
      </div>
    </div>
    <BoxLabel text="Security Type" />

    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div class="box overflow-x-auto">
          <div id="d3-portfolio-security-rating-treemap" style="width: 100%; position: relative" />
          <svg id="d3-portfolio-security-rating" />
        </div>
      </div>
    </div>
    <BoxLabel text="Security Rating" />

    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div class="box overflow-x-auto">
          <div
            id="d3-portfolio-security-industry-treemap"
            style="width: 100%; position: relative"
          />
          <svg id="d3-portfolio-security-industry" />
        </div>
      </div>
    </div>
    <BoxLabel text="Industry" />

    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div class="box overflow-x-auto">
          <div id="d3-portfolio-treemap" style="width: 100%; position: relative" />
          <svg id="d3-portfolio" />
        </div>
      </div>
    </div>
    <BoxLabel text="Security" />
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
