<script lang="ts">
  import { filterPostings, renderPostings } from "$lib/posting";
  import { ajax, type Posting } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  let rows: { date: string; posting: Posting; markup: string }[], clusterTable: any;

  const handleInput = _.debounce((event) => {
    const filter = event.srcElement.value;
    const filtered = filterPostings(rows, filter);
    clusterTable.update(_.map(filtered, (r) => r.markup));
  }, 100);

  onMount(async () => {
    const { postings: postings } = await ajax("/api/ledger");
    ({ rows, clusterTable } = renderPostings(postings));
  });
</script>

<section class="section tab-journal">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <nav class="level">
          <div class="level-left">
            <div class="level-item">
              <div class="field">
                <p class="control">
                  <input
                    class="d3-posting-filter input"
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
      <div class="column is-12">
        <div class="box">
          <div
            class="clusterize-scroll"
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
              <tbody class="d3-postings clusterize-content has-text-grey-dark" id="d3-postings" />
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</section>

<style lang="scss">
  @import "bulma/sass/utilities/_all.sass";
  .d3-posting-filter {
    width: calc(100vw - 35px);

    @include desktop {
      width: 440px;
    }
  }
</style>
