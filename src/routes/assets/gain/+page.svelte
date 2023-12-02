<script lang="ts">
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import { renderLegend, renderOverview } from "$lib/gain";
  import { ajax } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  onMount(async () => {
    const { gain_breakdown: gains } = await ajax("/api/gain");

    renderLegend();
    renderOverview(gains);
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
        <div class="box overflow-x-auto">
          <svg id="d3-gain-overview" />
        </div>
      </div>
    </div>
    <BoxLabel text="Gain Overview" />
  </div>
</section>
<section class="section tab-gain">
  <div class="container is-fluid d3-gain-timeline-breakdown">
    <div class="columns">
      <div id="d3-gain-timeline-breakdown" class="column is-12" />
    </div>
  </div>
</section>
