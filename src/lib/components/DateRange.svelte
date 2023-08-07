<script lang="ts">
  import type dayjs from "dayjs";

  export let value: number;
  export let dateMin: dayjs.Dayjs;
  export let dateMax: dayjs.Dayjs;

  let options: { label: string; value: number }[] = [];

  $: {
    options = [{ label: "All", value: -1 }];
    const diff = dateMax.diff(dateMin, "year");
    if (diff >= 10) {
      options.push({ label: "10 years", value: 10 });
    }

    if (diff >= 5) {
      options.push({ label: "5 years", value: 5 });
    }

    if (diff >= 3) {
      options.push({ label: "3 years", value: 3 });
    }
  }
</script>

{#if options.length > 1}
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
{/if}
