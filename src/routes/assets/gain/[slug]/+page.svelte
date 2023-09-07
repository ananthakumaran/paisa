<script lang="ts">
  import COLORS, { generateColorScheme, genericBarColor } from "$lib/colors";
  import { renderAccountOverview, renderLegend } from "$lib/gain";
  import { filterCommodityBreakdowns, renderPortfolioBreakdown } from "$lib/portfolio";
  import {
    ajax,
    type Posting,
    formatCurrency,
    formatFloat,
    type AccountGain,
    type Networth,
    type PortfolioAggregate
  } from "$lib/utils";
  import _ from "lodash";
  import { onMount, onDestroy } from "svelte";
  import type { PageData } from "./$types";
  import PostingCard from "$lib/components/PostingCard.svelte";
  import LevelItem from "$lib/components/LevelItem.svelte";

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
  let overview: Networth;

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

    overview = _.last(gain.networthTimeline);
    postings = _.chain(gain.postings)
      .sortBy((p) => p.date)
      .reverse()
      .take(100)
      .value();
    destroyCallback = renderAccountOverview(
      gain.networthTimeline,
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
      z: [genericBarColor()]
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
                  <LevelItem
                    narrow
                    title="Balance"
                    color={COLORS.primary}
                    value={formatCurrency(overview.balanceAmount)}
                  />
                  <LevelItem
                    narrow
                    title="Net Investment"
                    color={COLORS.secondary}
                    value={formatCurrency(overview.netInvestmentAmount)}
                  />
                </nav>
              </div>
            </div>
            <div class="column is-full">
              <div>
                <nav class="level">
                  <LevelItem
                    narrow
                    title="Gain / Loss"
                    color={overview.gainAmount >= 0 ? COLORS.gainText : COLORS.lossText}
                    value={formatCurrency(overview.gainAmount)}
                  />

                  <LevelItem narrow title="XIRR" value={formatFloat(gain.xirr)} />
                </nav>
              </div>
            </div>
          {/if}

          <div class="column is-full">
            Postings
            {#each postings as posting}
              <PostingCard
                {posting}
                color={posting.amount >= 0 ? COLORS.secondary : COLORS.tertiary}
              />
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
