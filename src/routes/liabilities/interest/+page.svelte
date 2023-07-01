<script lang="ts">
  import {
    renderLegend,
    renderOverview,
    renderPerAccountOverview
  } from "$lib/liabilities/interest";
  import { ajax } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  let isEmpty = false;

  onMount(async () => {
    const { interest_timeline_breakdown: interests } = await ajax("/api/liabilities/interest");

    if (_.isEmpty(interests)) {
      isEmpty = true;
      return;
    }

    renderLegend();
    renderOverview(interests);
    renderPerAccountOverview(interests);
  });
</script>

<section class="section tab-interest" class:is-hidden={!isEmpty}>
  <div class="container is-fluid">
    <div class="columns is-centered">
      <div class="column is-4 has-text-centered">
        <article class="message is-success">
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
        <svg id="d3-interest-legend" width="100%" height="50" />
      </div>
    </div>
  </div>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <svg id="d3-interest-overview" width="100%" />
        </div>
      </div>
    </div>
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div>
          <p class="heading">Interest Overview</p>
        </div>
      </div>
    </div>
  </div>
</section>
<section class="section tab-interest">
  <div class="container is-fluid d3-interest-timeline-breakdown">
    <div class="columns">
      <div id="d3-interest-timeline-breakdown" class="column is-12" />
    </div>
  </div>
</section>
