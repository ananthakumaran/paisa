<script lang="ts">
  import dayjs from "dayjs";

  let currentMonth = dayjs();
  let lastMonth = dayjs().subtract(1, "month");

  export let value: string = currentMonth.format("YYYY-MM");

  let options: { label: string; value: string }[] = [
    { label: lastMonth.format("MMMM"), value: lastMonth.format("YYYY-MM") },
    { label: currentMonth.format("MMMM"), value: currentMonth.format("YYYY-MM") }
  ];
</script>

<div class="du-tabs du-tabs-boxed">
  {#each options as option}
    <a
      class="du-tab du-tab-sm {option.value === value ? 'du-tab-active' : ''}"
      on:click={() => (value = option.value)}
    >
      {option.label}
    </a>
  {/each}
</div>
