<script lang="ts">
  import { filterTransactions, renderTransactions } from "$lib/journal";
  import { ajax, type Posting } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  let rows: { date: string; posting: Posting; markup: string }[], clusterTable: any;

  const handleInput = _.debounce((event) => {
    const filter = event.srcElement.value;
    const filtered = filterTransactions(rows, filter);
    clusterTable.update(_.map(filtered, (r) => r.markup));
  }, 100);

  onMount(async () => {
    const { postings: postings } = await ajax("/api/ledger");
    ({ rows, clusterTable } = renderTransactions(postings));
  });
</script>

<section class="section tab-journal">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <nav class="level">
          <div class="level-left">
            <div class="level-item">
              <p class="subtitle is-5">Postings</p>
            </div>
            <div class="level-item">
              <div class="field">
                <p class="control">
                  <input
                    class="d3-posting-filter input"
                    style="width: 440px"
                    type="text"
                    placeholder="filter by account or description or date"
                    on:input={handleInput}
                  />
                </p>
              </div>
            </div>
          </div>
        </nav>
      </div>
    </div>
    <div class="columns">
      <div
        class="column is-12 clusterize-scroll"
        id="d3-postings-container"
        style="max-height: calc(100vh - 200px); min-height: calc(100vh - 200px)"
      >
        <table class="table is-narrow is-fullwidth is-hoverable">
          <thead>
            <tr>
              <th>Date</th>
              <th>Description</th>
              <th>Account</th>
              <th class="has-text-right">Amount</th>
              <th class="has-text-right">Units</th>
              <th class="has-text-right">Unit Price</th>
              <th class="has-text-right">Market Value</th>
              <th class="has-text-right">Change</th>
              <th class="has-text-right">CAGR</th>
            </tr>
          </thead>
          <tbody class="d3-postings clusterize-content" id="d3-postings" />
        </table>
      </div>
    </div>
  </div>
</section>
