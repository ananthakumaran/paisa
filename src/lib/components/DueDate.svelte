<script lang="ts">
  import { dueDateIcon } from "$lib/utils";
  import dayjs from "dayjs";

  export let dueDate: dayjs.Dayjs;
  export let paidDate: dayjs.Dayjs;

  $: icon = dueDateIcon(dueDate, paidDate);
</script>

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
