<script lang="ts">
  import { ajax, formatCurrency, formatFloat, type Networth } from "$lib/utils";
  import COLORS from "$lib/colors";
  import { renderNetworth } from "$lib/networth";
  import _ from "lodash";
  import { onDestroy, onMount } from "svelte";
  import { dateRange, setAllowedDateRange } from "../../../store";

  let networth = 0;
  let investment = 0;
  let gain = 0;
  let xirr = 0;
  let svg: Element;
  let destroyCallback: () => void;
  let points: Networth[] = [];

  $: if (!_.isEmpty(points)) {
    if (destroyCallback) {
      destroyCallback();
    }

    destroyCallback = renderNetworth(
      _.filter(
        points,
        (p) => p.date.isSameOrBefore($dateRange.to) && p.date.isAfter($dateRange.from)
      ),
      svg
    );
  }

  onDestroy(async () => {
    destroyCallback();
  });

  onMount(async () => {
    const result = await ajax("/api/networth");
    points = result.networthTimeline;
    setAllowedDateRange(_.map(points, (p) => p.date));

    const current = _.last(points);
    networth = current.investmentAmount + current.gainAmount - current.withdrawalAmount;
    investment = current.investmentAmount - current.withdrawalAmount;
    gain = current.gainAmount;
    xirr = result.xirr;
  });
</script>

<section class="section tab-networth">
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

<section class="section tab-networth">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <svg id="d3-networth-timeline" width="100%" height="500" bind:this={svg} />
        </div>
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
      <div id="d3-networth-timeline-breakdown" class="column is-12" />
    </div>
  </div>
</section>
