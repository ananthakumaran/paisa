<script lang="ts">
  import type dayjs from "dayjs";
  import BoxedTabs from "./BoxedTabs.svelte";
  import { isMobile } from "$lib/utils";

  export let value: number;
  export let dateMin: dayjs.Dayjs;
  export let dateMax: dayjs.Dayjs;

  let options: { label: string; value: number }[] = [];

  $: {
    options = [{ label: "All", value: -1 }];
    const diff = dateMax.diff(dateMin, "year");
    if (diff >= 10 && !isMobile()) {
      options.push({ label: "10 years", value: 10 });
    }

    if (diff >= 5 && !isMobile()) {
      options.push({ label: "5 years", value: 5 });
    }

    if (diff >= 3) {
      options.push({ label: "3 years", value: 3 });
    }

    if (diff >= 1) {
      options.push({ label: "1 year", value: 1 });
    }
  }
</script>

{#if options.length > 1}
  <BoxedTabs bind:value {options} />
{/if}
