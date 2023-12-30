<script lang="ts">
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import LegendCard from "$lib/components/LegendCard.svelte";
  import {
    buildLegends,
    renderOverview,
    renderPerAccountOverview
  } from "$lib/liabilities/interest";
  import { ajax, type Legend } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  let isEmpty = false;
  let legends: Legend[] = [];

  onMount(async () => {
    const { interest_timeline_breakdown: interests } = await ajax("/api/liabilities/interest");

    if (_.isEmpty(interests)) {
      isEmpty = true;
      return;
    }

    legends = buildLegends();
    renderOverview(interests);
    renderPerAccountOverview(interests);
  });
</script>

<section class="section tab-interest" class:is-hidden={!isEmpty}>
  <div class="container is-fluid">
    <div class="columns is-centered">
      <div class="column is-4 has-text-centered">
        <article class="message">
          <div class="message-body">
            <strong>Hurray!</strong> You have no liabilities.
          </div>
        </article>
      </div>
    </div>
  </div>
</section>

<section class="section tab-interest" class:is-hidden={isEmpty}>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <LegendCard {legends} />
      </div>
    </div>
  </div>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box overflow-x-auto">
          <svg id="d3-interest-overview" />
        </div>
      </div>
    </div>
    <BoxLabel text="Interest Overview" />
  </div>
</section>
<section class="section tab-interest">
  <div class="container is-fluid d3-interest-timeline-breakdown">
    <div class="columns">
      <div id="d3-interest-timeline-breakdown" class="column is-12" />
    </div>
  </div>
</section>
