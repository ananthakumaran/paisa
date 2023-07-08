<script lang="ts">
  import COLORS from "$lib/colors";
  import { renderAccountOverview, renderLegend } from "$lib/gain";
  import {
    ajax,
    type Posting,
    formatCurrency,
    formatFloat,
    type AccountGain,
    type Overview
  } from "$lib/utils";
  import _ from "lodash";
  import { onMount, onDestroy } from "svelte";

  export let data: { name: string };
  let gain: AccountGain;
  let overview: Overview;

  let destroyCallback = () => {};
  let postings: Posting[] = [];

  onDestroy(async () => {
    destroyCallback();
  });

  onMount(async () => {
    ({ gain_timeline_breakdown: gain } = await ajax("/api/gain/:name", null, data));

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
  });
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
                      <p class="title has-text-black">
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
              <div class="box p-2 my-2 has-background-white">
                <div class="is-flex is-flex-wrap-wrap is-justify-content-space-between">
                  <div class="has-text-grey is-size-7">{posting.payee}</div>
                  <div class="has-text-grey">
                    <span
                      class="icon is-small"
                      style="opacity: 0.6; color: {posting.amount >= 0
                        ? COLORS.secondary
                        : COLORS.tertiary} !important"
                    >
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
        <div class="box">
          <svg id="d3-account-timeline-breakdown" width="100%" height="450" />
        </div>
        <div class="has-text-centered">
          <p class="heading">Timeline</p>
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
