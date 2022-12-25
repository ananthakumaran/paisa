<script lang="ts">
  import { renderBreakdowns } from "$lib/schedule_al";
  import { ajax, setHtml } from "$lib/utils";
  import dayjs from "dayjs";
  import { onMount } from "svelte";

  onMount(async () => {
    const { schedule_al_entries: scheduleALEntries, date: date } = await ajax("/api/schedule_al");

    setHtml("schedule-al-date", dayjs(date).format("DD MMM YYYY"));
    renderBreakdowns(scheduleALEntries);
  });
</script>

<section class="section tab-schedule_al">
  <div class="container is-fluid">
    <div class="container is-fluid">
      <div class="columns">
        <div class="column is-12">
          <p class="subtitle is-12">
            Schedule AL as on <span class="d3-schedule-al-date has-text-weight-bold" />
          </p>
        </div>
      </div>
      <div class="columns">
        <div class="column is-12">
          <table class="table is-narrow is-fullwidth">
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
