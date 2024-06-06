<script lang="ts">
  import { dueDateIcon } from "$lib/utils";
  import dayjs from "dayjs";

  export let dueDate: dayjs.Dayjs;
  export let paidDate: dayjs.Dayjs;

  $: icon = dueDateIcon(dueDate, paidDate);
</script>

<span title="{dueDate.format('DD MMM YYYY')} tarihinde ödeme yapılması gerekiyor">
  <span class="icon is-small {icon.color}">
    <i class="fas {icon.icon}" />
  </span>
  {#if paidDate}
    <span>{paidDate.format("DD MMM YYYY")} tarihinde ödendi</span>
  {:else}
    <span>{dueDate.fromNow()} içerisinde ödeme yapılmalı</span>
  {/if}
</span>