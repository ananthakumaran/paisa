<script lang="ts">
  import {
    intervalText,
    nextUnpaidSchedule,
    scheduleIcon,
    totalRecurring
  } from "$lib/transaction_sequence";
  import { formatCurrencyCrude, type TransactionSequence } from "$lib/utils";
  import dayjs from "dayjs";

  export let transactionSequece: TransactionSequence;

  let schedule = nextUnpaidSchedule(transactionSequece);
  let n = schedule.scheduled;
  const now = dayjs();
  const icon = scheduleIcon(schedule);
</script>

<div class="has-text-centered mb-3 mr-3 max-w-[200px]">
  <div class="is-size-7 truncate">{transactionSequece.key}</div>
  <div class="my-1">
    <span class="tag">{intervalText(transactionSequece)}</span>
  </div>
  <div class="has-text-grey is-size-7">
    <span class="icon has-text-grey-light">
      <i class="fas fa-calendar" />
    </span>
    {schedule.scheduled.format("DD MMM YYYY")}
  </div>
  <div
    class="m-3 du-radial-progress is-size-7 {icon.color}"
    style="--value: {n.isBefore(now)
      ? '0'
      : (schedule.scheduled.diff(now, 'day') / transactionSequece.interval) *
        100}; --thickness: 3px; --size: 100px;"
  >
    <div class="is-size-6">
      <span class="icon">
        <i class="fas {icon.icon}" />
      </span>
    </div>
    <span>{formatCurrencyCrude(totalRecurring(transactionSequece))}</span>
    <span>due {n.fromNow()}</span>
  </div>
</div>
