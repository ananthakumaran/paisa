<script lang="ts">
  import { ajax, type Transaction as T } from "$lib/utils";
  import { filterTransactions } from "$lib/transaction";
  import _ from "lodash";
  import { onMount } from "svelte";
  import VirtualList from "svelte-tiny-virtual-list";
  import Transaction from "$lib/components/Transaction.svelte";

  let transactions: T[] = [];
  let filtered: T[] = [];

  const debits = (t: T) => {
    return _.filter(t.postings, (p) => p.amount < 0);
  };

  const credits = (t: T) => {
    return _.filter(t.postings, (p) => p.amount >= 0);
  };

  const handleInput = _.debounce((event) => {
    const filter = event.srcElement.value;
    filtered = filterTransactions(transactions, filter);
  }, 100);

  const itemSize = (i: number) => {
    const t = filtered[i];
    return 2 + Math.max(credits(t).length, debits(t).length) * 30;
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
      <div class="column is-12">
        <div class="box">
          <VirtualList
            width="100%"
            height={window.innerHeight - 150}
            itemCount={filtered.length}
            {itemSize}
          >
            <div slot="item" let:index let:style {style}>
              {@const t = filtered[index]}
              <Transaction {t} />
            </div>
          </VirtualList>
        </div>
      </div>
    </div>
  </div>
</section>
