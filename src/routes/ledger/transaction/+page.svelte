<script lang="ts">
  import { ajax, type Transaction } from "$lib/utils";
  import { filterTransactions } from "$lib/transaction";
  import _ from "lodash";
  import { onMount } from "svelte";
  import VirtualList from "svelte-tiny-virtual-list";
  import Postings from "$lib/components/Postings.svelte";

  let transactions: Transaction[] = [];
  let filtered: Transaction[] = [];

  const debits = (t: Transaction) => {
    return _.filter(t.postings, (p) => p.amount < 0);
  };

  const credits = (t: Transaction) => {
    return _.filter(t.postings, (p) => p.amount >= 0);
  };

  const handleInput = _.debounce((event) => {
    const filter = event.srcElement.value;
    filtered = filterTransactions(transactions, filter);
  }, 100);

  const itemSize = (i: number) => {
    const t = filtered[i];
    return 50 + Math.max(credits(t).length, debits(t).length) * 30;
  };

  onMount(async () => {
    ({ transactions } = await ajax("/api/transaction"));
    filtered = transactions;
  });
</script>

<section class="section tab-journal">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <nav class="level">
          <div class="level-left">
            <div class="level-item">
              <p class="subtitle is-5">Transactions</p>
            </div>
            <div class="level-item">
              <div class="field">
                <p class="control">
                  <input
                    class="d3-transaction-filter input"
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
      <VirtualList
        width="100%"
        height={window.innerHeight - 150}
        itemCount={filtered.length}
        {itemSize}
      >
        <div slot="item" let:index let:style {style}>
          {@const t = filtered[index]}
          <div class="column is-12">
            <div class="columns is-flex-wrap-wrap">
              <div class="column is-full pb-0 px-0">
                <span
                  style="display: inline-block"
                  class="is-bordered-left is-bordered-top is-bordered-right p-2"
                >
                  <b>{t.date.format("DD MMM YYYY")}</b>
                  <span>{t.payee}</span>
                </span>
              </div>
              <div class="column is-half py-0 is-bordered-bottom is-bordered-top is-bordered-left">
                <Postings postings={debits(t)} />
              </div>
              <div class="column is-half py-0 is-bordered-bottom is-bordered-top is-bordered-right">
                <Postings postings={credits(t)} />
              </div>
            </div>
          </div>
        </div>
      </VirtualList>
    </div>
  </div>
</section>
