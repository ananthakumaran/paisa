<script lang="ts">
  import { renderBreakdowns } from "$lib/holding";
  import { ajax } from "$lib/utils";
  import { onMount } from "svelte";

  onMount(async () => {
    const { breakdowns: breakdowns } = await ajax("/api/ledger");
    renderBreakdowns(breakdowns);
  });
</script>

<section class="section tab-holding">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <p class="subtitle is-12">Holdings</p>
      </div>
    </div>
    <div class="columns">
      <div class="column is-12">
        <table class="table is-narrow is-fullwidth is-hoverable">
          <thead>
            <tr>
              <th>Account</th>
              <th class="has-text-right">Investment Amount</th>
              <th class="has-text-right">Withdrawal Amount</th>
              <th class="has-text-right">Balance Units</th>
              <th class="has-text-right">Market Value</th>
              <th class="has-text-right">Change</th>
              <th class="has-text-right">XIRR</th>
            </tr>
          </thead>
          <tbody class="d3-postings-breakdown" />
        </table>
      </div>
    </div>
  </div>
</section>
