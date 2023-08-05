<script lang="ts">
  import COLORS from "$lib/colors";
  import { renderAccountOverview, renderLegend } from "$lib/gain";
  import { filterCommodityBreakdowns, renderPortfolioBreakdown } from "$lib/portfolio";
  import {
    ajax,
    type Posting,
    formatCurrency,
    formatFloat,
    type AccountGain,
    type Overview,
    type PortfolioAggregate,
    generateColorScheme
  } from "$lib/utils";
  import _ from "lodash";
  import { onMount, onDestroy } from "svelte";
  import type { PageData } from "./$types";

  let commodities: string[] = [];
  let selectedCommodities: string[] = [];
  let security_type: PortfolioAggregate[] = [];
  let name_and_security_type: PortfolioAggregate[] = [];
  let rating: PortfolioAggregate[] = [];
  let industry: PortfolioAggregate[] = [];
  let color: any;

  let securityTypeEmpty: boolean = false;
  let nameAndSecurityTypeEmpty: boolean = false;
  let ratingEmpty: boolean = false;
  let industryEmpty: boolean = false;

  export let data: PageData;
  let gain: AccountGain;
  let overview: Overview;

  let destroyCallback = () => {};
  let postings: Posting[] = [];

  let securityTypeR: any,
    portfolioR: any,
    industryR: any,
    ratingR: any = null;

  onDestroy(async () => {
    destroyCallback();
  });

  onMount(async () => {
    ({
      gain_timeline_breakdown: gain,
      portfolio_allocation: { name_and_security_type, security_type, rating, industry, commodities }
    } = await ajax("/api/gain/:name", null, data));

    overview = _.last(gain.overview_timeline);
    postings = _.chain(gain.postings)
      .sortBy((p) => p.date)
      .reverse()
      .take(100)
      .value();
    destroyCallback = renderAccountOverview(
      gain.overview_timeline,
      gain.postings,
      "d3-account-timeline-breakdown"
    );
    renderLegend();

    selectedCommodities = [...commodities];
    securityTypeR = renderPortfolioBreakdown("#d3-portfolio-security-type", security_type, {
      showLegend: false,
      small: true
    });
    ratingR = renderPortfolioBreakdown("#d3-portfolio-security-rating", rating, {
      showLegend: false,
      small: true
    });
    industryR = renderPortfolioBreakdown("#d3-portfolio-security-industry", industry, {
      showLegend: false,
      small: true,
      z: ["#D9D9D9"]
    });
    portfolioR = renderPortfolioBreakdown("#d3-portfolio", name_and_security_type, {
      showLegend: true,
      small: true
    });

    if (commodities.length !== 0) {
      color = generateColorScheme(commodities);
    }

    securityTypeEmpty = security_type.length === 0;
    nameAndSecurityTypeEmpty = name_and_security_type.length === 0;
    ratingEmpty = rating.length === 0;
    industryEmpty = industry.length === 0;
  });

  $: if (securityTypeR) {
    securityTypeR(filterCommodityBreakdowns(security_type, selectedCommodities), color);
    ratingR(filterCommodityBreakdowns(rating, selectedCommodities), color);
    industryR(filterCommodityBreakdowns(industry, selectedCommodities), color);
    portfolioR(filterCommodityBreakdowns(name_and_security_type, selectedCommodities), color);
  }
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns is-flex-wrap-wrap">
      <div class="column is-3">
        <div class="columns is-flex-wrap-wrap">
          {#if overview}
            <div class="column is-full">
              <div>
                <nav class="level">
                  <div class="level-item is-narrow">
                    <div>
                      <p class="heading has-text-left">Balance</p>
                      <p class="title" style="background-color: {COLORS.primary};">
                        {formatCurrency(overview.balance_amount)}
                      </p>
                    </div>
                  </div>
                  <div class="level-item is-narrow">
                    <div>
                      <p class="heading has-text-right">Net Investment</p>
                      <p class="title" style="background-color: {COLORS.secondary};">
                        {formatCurrency(overview.net_investment_amount)}
                      </p>
                    </div>
                  </div>
                </nav>
              </div>
            </div>
            <div class="column is-full">
              <div>
                <nav class="level">
                  <div class="level-item is-narrow">
                    <div>
                      <p class="heading has-text-left">Gain / Loss</p>
                      <p
                        class="title"
                        style="background-color: {overview.gain_amount >= 0
                          ? COLORS.gainText
                          : COLORS.lossText};"
                      >
                        {formatCurrency(overview.gain_amount)}
                      </p>
                    </div>
                  </div>
                  <div class="level-item is-narrow">
                    <div>
                      <p class="heading has-text-right">XIRR</p>
                      <p class="title has-text-black pr-0">
                        {formatFloat(gain.xirr)}
                      </p>
                    </div>
                  </div>
                </nav>
              </div>
            </div>
          {/if}

          <div class="column is-full">
            Postings
            {#each postings as posting}
              <div
                class="box p-2 my-2 has-background-white"
                style="border-left: 2px solid {posting.amount >= 0
                  ? COLORS.secondary
                  : COLORS.tertiary}"
              >
                <div class="is-flex is-justify-content-space-between">
                  <div class="has-text-grey is-size-7 truncate">{posting.payee}</div>
                  <div class="has-text-grey min-w-[100px]">
                    <span class="icon is-small has-text-grey-light">
                      <i class="fas fa-calendar" />
                    </span>
                    {posting.date.format("DD MMM YYYY")}
                  </div>
                </div>
                <hr class="m-1" />
                <div class="is-flex is-flex-wrap-wrap is-justify-content-space-between">
                  <div class="has-text-grey">
                    {posting.account}
                  </div>
                  <div>
                    <span
                      class="has-text-grey mr-1"
                      class:is-hidden={posting.quantity == posting.amount}
                      >{formatFloat(posting.quantity, 4)} @ {formatFloat(
                        posting.amount / posting.quantity,
                        3
                      )}</span
                    >
                    <span class="has-text-weight-bold is-size-6"
                      >{formatCurrency(posting.amount)}</span
                    >
                  </div>
                </div>
              </div>
            {/each}
          </div>
        </div>
      </div>
      <div class="column is-9">
        <svg id="d3-gain-legend" width="100%" height="50" />
        <div class="box mb-2">
          <svg id="d3-account-timeline-breakdown" width="100%" height="450" />
        </div>
        <div class="has-text-centered">
          <p class="heading">Timeline</p>
        </div>

        <div class="columns">
          <div class="column is-6">
            <div class="mt-5" class:is-hidden={securityTypeEmpty}>
              <div class="box mb-2">
                <div
                  id="d3-portfolio-security-type-treemap"
                  style="width: 100%; position: relative"
                />
                <svg id="d3-portfolio-security-type" width="100%" />
              </div>
              <div class="has-text-centered">
                <p class="heading">Security Type</p>
              </div>
            </div>

            <div class="mt-5" class:is-hidden={ratingEmpty}>
              <div class="box mb-2">
                <div
                  id="d3-portfolio-security-rating-treemap"
                  style="width: 100%; position: relative"
                />
                <svg id="d3-portfolio-security-rating" width="100%" />
              </div>
              <div class="has-text-centered">
                <p class="heading">Security Rating</p>
              </div>
            </div>

            <div class="mt-5" class:is-hidden={industryEmpty}>
              <div class="box mb-2">
                <div
                  id="d3-portfolio-security-industry-treemap"
                  style="width: 100%; position: relative"
                />
                <svg id="d3-portfolio-security-industry" width="100%" />
              </div>
              <div class="has-text-centered">
                <p class="heading">Industry</p>
              </div>
            </div>
          </div>
          <div class="column is-6 mt-5">
            <div class:is-hidden={nameAndSecurityTypeEmpty}>
              <div class="box mb-2">
                <div id="d3-portfolio-treemap" style="width: 100%; position: relative" />
                <svg id="d3-portfolio" width="100%" />
              </div>
              <div class="has-text-centered">
                <p class="heading">Security</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</section>

<style>
  .level-item p.title {
    padding: 5px;
    color: white;
  }
</style>
