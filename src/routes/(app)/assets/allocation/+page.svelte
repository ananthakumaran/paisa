<script lang="ts">
  import {
    renderAllocation,
    renderAllocationTarget,
    renderAllocationTimeline
  } from "$lib/allocation";
  import { generateColorScheme } from "$lib/colors";
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import { ajax } from "$lib/utils";
  import _ from "lodash";
  import { onMount, tick } from "svelte";

  let showAllocation = false;
  let depth = 2;

  onMount(async () => {
    const {
      aggregates: aggregates,
      aggregates_timeline: aggregatesTimeline,
      allocation_targets: allocationTargets
    } = await ajax("/api/allocation");
    const accounts = _.keys(aggregates);
    const color = generateColorScheme(accounts);
    depth = _.max(_.map(accounts, (account) => account.split(":").length));

    if (!_.isEmpty(allocationTargets)) {
      showAllocation = true;
    }
    await tick();

    renderAllocationTarget(allocationTargets, color);
    renderAllocation(aggregates, color);
    renderAllocationTimeline(aggregatesTimeline);
  });
</script>

<section class="section tab-allocation" style={showAllocation ? "" : "display: none"}>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div class="box overflow-x-auto">
          <div id="d3-allocation-target-treemap" style="width: 100%; position: relative" />
          <svg id="d3-allocation-target" />
        </div>
      </div>
    </div>
    <BoxLabel text="Allocation Targets" />
  </div>
</section>
<section class="section tab-allocation">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div id="d3-allocation-category" style="width: 100%; height: {depth * 100}px" />
      </div>
    </div>
    <BoxLabel text="Allocation by category" />
  </div>
</section>
<section class="section tab-allocation">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div id="d3-allocation-value" style="width: 100%; height: 300px" />
      </div>
    </div>
    <BoxLabel text="Allocation by value" />
  </div>
</section>
<section class="section tab-allocation">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <svg id="d3-allocation-timeline" width="100%" height="300" />
        </div>
      </div>
    </div>
    <BoxLabel text="Allocation Timeline" />
  </div>
</section>
