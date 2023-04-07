<script lang="ts">
  import { ajax, formatCurrency, formatFloat } from "$lib/utils";
  import COLORS from "$lib/colors";
  import { renderOverview } from "$lib/overview";
  import _ from "lodash";
  import { onDestroy, onMount } from "svelte";

  let networth = 0;
  let investment = 0;
  let gain = 0;
  let xirr = 0;
  let svg: Element;
  let destroyCallback: () => void;

  onDestroy(async () => {
    destroyCallback();
  });
  onMount(async () => {
    const result = await ajax("/api/overview");
    const points = result.overview_timeline;

    const current = _.last(points);
    networth = current.investment_amount + current.gain_amount - current.withdrawal_amount;
    investment = current.investment_amount - current.withdrawal_amount;
    gain = current.gain_amount;
    xirr = result.xirr;

    destroyCallback = renderOverview(points, svg);
  });
</script>

<section class="section tab-overview">
  <div class="container">
    <nav class="level">
      <div class="level-item has-text-centered">
        <div>
          <p class="heading">Net worth</p>
          <p class="title" style="background-color: {COLORS.primary};">
            {formatCurrency(networth)}
          </p>
        </div>
      </div>
      <div class="level-item has-text-centered">
        <div>
          <p class="heading">Net Investment</p>
          <p class="title" style="background-color: {COLORS.secondary};">
            {formatCurrency(investment)}
          </p>
        </div>
      </div>
      <div class="level-item has-text-centered">
        <div>
          <p class="heading">Gain / Loss</p>
          <p
            class="title"
            style="background-color: {gain >= 0 ? COLORS.gainText : COLORS.lossText};"
          >
            {formatCurrency(gain)}
          </p>
        </div>
      </div>
      <div class="level-item has-text-centered">
        <div>
          <p class="heading">XIRR</p>
          <p class="title has-text-black">{formatFloat(xirr)}</p>
        </div>
      </div>
    </nav>
  </div>
</section>

<section class="section tab-overview">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <svg id="d3-overview-timeline" width="100%" height="500" bind:this={svg} />
      </div>
    </div>
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div>
          <p class="heading">Networth Timeline</p>
        </div>
      </div>
    </div>
  </div>
  <div class="container is-fluid">
    <div class="columns">
      <div id="d3-overview-timeline-breakdown" class="column is-12" />
    </div>
  </div>
</section>

<style>
  .level-item p.title {
    padding: 5px;
    color: white;
  }
</style>
