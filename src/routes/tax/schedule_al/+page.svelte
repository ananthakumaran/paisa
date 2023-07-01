<script lang="ts">
  import { renderBreakdowns } from "$lib/schedule_al";
  import _ from "lodash";
  import { ajax, type ScheduleAL } from "$lib/utils";
  import { onMount } from "svelte";
  import { dateMin, year } from "../../../store";

  let scheduleAls: Record<string, ScheduleAL>;
  let selectedScheduleAl: ScheduleAL;

  $: if (scheduleAls) {
    selectedScheduleAl = scheduleAls[$year];
  }

  $: if (selectedScheduleAl) {
    renderBreakdowns(selectedScheduleAl.entries);
  }

  onMount(async () => {
    ({ schedule_als: scheduleAls } = await ajax("/api/schedule_al"));

    let firstScheduleAl = _.minBy(Object.values(scheduleAls), (e) => e.date);
    if (firstScheduleAl) {
      dateMin.set(firstScheduleAl.date);
    }
  });
</script>

<section class="section tab-schedule_al">
  <div class="container is-fluid">
    {#if selectedScheduleAl}
      <div class="columns">
        <div class="column is-12">
          <p class="subtitle is-12">
            Schedule AL as on <span class="has-text-weight-bold"
              >{selectedScheduleAl.date.format("DD MMM YYYY")}</span
            >
          </p>
        </div>
      </div>
    {/if}

    <div class="columns">
      <div class="column is-12">
        <div class="box px-3" style="max-width: 1024px;">
          <table class="table is-narrow is-fullwidth is-hoverable">
            <thead>
              <tr>
                <th>Code</th>
                <th>Section</th>
                <th>Details</th>
                <th class="has-text-right">Amount</th>
              </tr>
            </thead>
            <tbody class="d3-schedule-al" />
          </table>
        </div>
      </div>
    </div>
  </div>
</section>
