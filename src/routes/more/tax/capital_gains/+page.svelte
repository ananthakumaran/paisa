<script lang="ts">
  import type { CapitalGain } from "$lib/utils";
  import CapitalGainCard from "$lib/components/CapitalGainCard.svelte";
  import { ajax } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  let years: string[] = [];
  let capitalGains: CapitalGain[] = [];

  onMount(async () => {
    const { capital_gains: capital_gains } = await ajax("/api/capital_gains");

    years = _.chain(capital_gains)
      .values()
      .flatMap((c) => _.keys(c.fy))
      .uniq()
      .sort()
      .reverse()
      .value();

    capitalGains = _.values(capital_gains);
  });
</script>

<section class="section tab-capital-gains">
  <div class="container is-fluid">
    <div class="columns is-flex-wrap-wrap">
      {#each years as year}
        <CapitalGainCard financialYear={year} {capitalGains} />
      {/each}
    </div>
  </div>
</section>
