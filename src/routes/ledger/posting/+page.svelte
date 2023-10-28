<script lang="ts">
  import { accountColorStyle } from "$lib/colors";
  import { iconText } from "$lib/icon";
  import { change, filterPostings } from "$lib/posting";
  import {
    ajax,
    postingUrl,
    type Posting,
    formatCurrency,
    formatFloat,
    firstName
  } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  import VirtualList from "svelte-tiny-virtual-list";

  let filteredPostings: Posting[] = [];
  let rows: { date: string; posting: Posting }[];

  const handleInput = _.debounce((event) => {
    const filter = event.srcElement.value;
    const filtered = filterPostings(rows, filter);
    filteredPostings = filtered.map((r) => r.posting);
  }, 100);

  onMount(async () => {
    const { postings: postings } = await ajax("/api/ledger");
    filteredPostings = postings;
    rows = _.map(postings, (p) => ({ date: p.date.format("YYYY-MM-DD"), posting: p }));
  });

  function unlessDefault(p: Posting, text: string) {
    if (p.commodity !== USER_CONFIG.default_currency) {
      return text;
    }
    return "";
  }

  function unlessZero(value: number, text: string) {
    if (value > 0) {
      return text;
    }
    return "";
  }
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
        <div class="box overflow-x-auto" style="max-width: 98rem;">
          <div style="width: 98rem;">
            <div
              class="px-3 pt-1 grid grid-cols-7x gap-1 posting-row items-baseline has-text-weight-bold"
            >
              <div>Date</div>
              <div>Description</div>
              <div>Account</div>
              <div class="has-text-right">Amount</div>
              <div class="has-text-right">Balance</div>
              <div class="has-text-right">Units</div>
              <div class="has-text-right">Unit Price</div>
              <div class="has-text-right">Market Value</div>
              <div class="has-text-right">Change</div>
              <div class="has-text-right">CAGR</div>
            </div>
            <VirtualList
              height={window.innerHeight - 245}
              itemCount={filteredPostings.length}
              itemSize={27}
            >
              <div
                slot="item"
                class="px-3 pt-1 grid grid-cols-7x gap-1 posting-row items-baseline is-hoverable"
                let:index
                let:style
                {style}
              >
                {@const p = filteredPostings[index]}
                {@const c = change(p)}
                <div>{p.date.format("DD MMM YYYY")}</div>
                <div class="is-size-7 truncate" title={p.payee}>
                  {#if p.status == "cleared"}
                    <span class="icon is-small">
                      <i class="fas fa-check"></i>
                    </span>
                  {:else if p.status == "pending"}
                    <span class="icon is-small">
                      <i class="fas fa-exclamation"></i>
                    </span>
                  {/if}<a class="secondary-link" href={postingUrl(p)}>{p.payee}</a>
                </div>
                <div class="custom-icon truncate" title={p.account}>
                  <div class="flex">
                    <span class="mr-1" style={accountColorStyle(firstName(p.account))}
                      >{iconText(p.account)}</span
                    >
                    {p.account}
                  </div>
                </div>
                <div class="has-text-right">{formatCurrency(p.amount, 2)}</div>
                <div class="has-text-right">{formatCurrency(p.balance, 2)}</div>
                <div class="has-text-right">{unlessDefault(p, formatFloat(p.quantity, 4))}</div>
                <div class="has-text-right">
                  {unlessDefault(p, formatCurrency(Math.abs(p.amount / p.quantity), 4))}
                </div>
                <div class="has-text-right">
                  {unlessDefault(p, unlessZero(c.days, formatCurrency(p.market_amount)))}
                </div>
                <div class="has-text-right {c.class}">
                  {unlessZero(c.value, formatCurrency(c.value))}
                </div>
                <div class="has-text-right {c.class}">
                  {unlessZero(c.percentage, formatFloat(c.percentage))}
                </div>
              </div>
            </VirtualList>
          </div>
        </div>
      </div>
    </div>
  </div>
</section>

<style lang="scss">
  @import "bulma/sass/utilities/_all.sass";

  input.d3-posting-filter {
    width: calc(100vw - 35px);

    @include desktop {
      width: 440px;
    }
  }
</style>
