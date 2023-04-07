<script lang="ts">
  import { renderPrices } from "$lib/price";
  import { ajax } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  onMount(async () => {
    const { prices: prices } = await ajax("/api/price");
    renderPrices(prices);
  });
</script>

<section class="section tab-price">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <p class="subtitle is-12">Prices</p>
      </div>
    </div>
    <div class="columns">
      <div class="column is-12">
        <table class="table is-narrow is-fullwidth is-hoverable">
          <thead>
            <tr>
              <th>Commodity Name</th>
              <th>Commodity Type</th>
              <th>Commodity ID</th>
              <th>Date</th>
              <th class="has-text-right">Value</th>
            </tr>
          </thead>
          <tbody class="d3-prices" />
        </table>
      </div>
    </div>
  </div>
</section>
