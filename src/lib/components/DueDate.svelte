<script lang="ts">
  import { dueDateIcon } from "$lib/utils";
  import dayjs from "dayjs";

  export let dueDate: dayjs.Dayjs;
  export let paidDate: dayjs.Dayjs;
  export let amountDue: number;

  $: icon = dueDateIcon(dueDate, paidDate);
</script>

{#if !paidDate && amountDue < 1}
  <span>No dues</span>
{:else}
  <span title="due on {dueDate.format('DD MMM YYYY')}">
    <span class="icon is-small {icon.color}">
      <i class="fas {icon.icon}" />
    </span>
    {#if paidDate}
      <span>paid on {paidDate.format("DD MMM YYYY")}</span>
    {:else}
      <span>due {dueDate.fromNow()}</span>
    {/if}
  </span>
{/if}
