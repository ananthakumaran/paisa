<script lang="ts">
  import { scheduleIcon } from "$lib/transaction_sequence";
  import {
    formatCurrency,
    formatCurrencyCrude,
    tooltip,
    type TransactionSchedule
  } from "$lib/utils";
  export let schedule: TransactionSchedule;

  const icon = scheduleIcon(schedule);

  const tooltipHtml = tooltip(
    [
      [
        "Due Date",
        [schedule.scheduled.format("DD MMM YYYY"), "has-text-weight-bold has-text-right"]
      ],
      [
        "Cleared On",
        [schedule.actual?.format("DD MMM YYYY") || "", "has-text-weight-bold has-text-right"]
      ],
      ["Amount", [formatCurrency(schedule.amount), "has-text-weight-bold has-text-right"]]
    ],
    { header: schedule.key }
  );
</script>

<div class="px-2 flex is-size-6 justify-between gap-2" data-tippy-content={tooltipHtml}>
  <div class="truncate" title={schedule.key}>
    <span class="icon is-small {icon.color}">
      <i class="fas {icon.icon}" />
    </span>
    <span class="ml-1">
      {schedule.key}
    </span>
  </div>
  <div>{formatCurrencyCrude(schedule.amount)}</div>
</div>
