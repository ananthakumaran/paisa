<script lang="ts">
  import dayjs from "dayjs";
  import _ from "lodash";

  export let n = 2;

  let currentMonth = dayjs();
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
