<script lang="ts">
  import type { TransactionSchedule } from "$lib/utils";
  import type { Dayjs } from "dayjs";
  import RecurringSchedule from "./RecurringSchedule.svelte";
  import dayjs from "dayjs";

  export let month: string;
  export let day: Dayjs;
  export let schedules: TransactionSchedule[];
  const isToday = day.isSame(dayjs(), "day");
</script>

<div class="box m-0 p-0" class:is-invisible={day.format("YYYY-MM") != month}>
  <div class="has-text-centered has-text-weight-bold mt-1 mb-1">
    <span
      class="is-size-6 px-2 py-1 {isToday
        ? 'rounded-full is-bordered is-link has-text-link'
        : 'has-text-grey'}">{day.format("D")}</span
    >
  </div>

  {#each schedules as schedule (schedule)}
    <RecurringSchedule {schedule} />
  {/each}
</div>
