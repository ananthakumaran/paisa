<script lang="ts">
  import _ from "lodash";
  import BoxedTabs from "./BoxedTabs.svelte";
  import { now } from "$lib/utils";

  export let n = 2;

  let currentMonth = now();
  export let value: string = currentMonth.format("YYYY-MM");

  let options: { label: string; value: string }[] = _.reverse(
    _.map(_.range(0, n), (i) => {
      let month = currentMonth.subtract(i, "month");
      return {
        label: month.format("MMMM"),
        value: month.format("YYYY-MM")
      };
    })
  );
</script>

<BoxedTabs bind:value {options} />
