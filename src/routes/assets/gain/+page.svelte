<script lang="ts">
  import { renderLegend, renderOverview, renderPerAccountOverview } from "$lib/gain";
  import { ajax } from "$lib/utils";
  import dayjs from "dayjs";
  import _ from "lodash";
  import { onMount } from "svelte";

  onMount(async () => {
    const { gain_timeline_breakdown: gains } = await ajax("/api/gain");
    _.each(gains, (g) => _.each(g.overview_timeline, (o) => (o.timestamp = dayjs(o.date))));

    renderLegend();
    renderOverview(gains);
    renderPerAccountOverview(gains);
  });
</script>

<section class="section tab-gain">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <svg id="d3-gain-legend" width="100%" height="50" />
      </div>
    </div>
  </div>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <svg id="d3-gain-overview" width="100%" />
      </div>
    </div>
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div>
          <p class="heading">Gain Overview</p>
        </div>
      </div>
    </div>
  </div>
</section>
<section class="section tab-gain">
  <div class="container is-fluid d3-gain-timeline-breakdown">
    <div class="columns">
      <div id="d3-gain-timeline-breakdown" class="column is-12" />
    </div>
  </div>
</section>
