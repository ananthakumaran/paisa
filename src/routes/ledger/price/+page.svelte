<script lang="ts">
  import Toggleable from "$lib/components/Toggleable.svelte";
  import ValueChange from "$lib/components/ValueChange.svelte";
  import { ajax, formatCurrency, type Price } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  import VirtualList from "svelte-tiny-virtual-list";

  let prices: Record<string, Price[]> = {};

  const ITEM_SIZE = 18;

  function change(prices: Price[], days: number, tolerance: number) {
    const first = prices[0];
    if (!first) return null;

    const date = first.date.subtract(days, "day");
    const last = _.find(prices, (p) => p.date.isSameOrBefore(date, "day"));
    if (!last) return null;

    const diffDays = first.date.diff(last.date, "day");
    if (Math.abs(diffDays - days) <= tolerance) {
      return (first.value - last.value) / last.value;
    }
    return null;
  }

  onMount(async () => {
    ({ prices: prices } = await ajax("/api/price"));
    prices = _.omitBy(prices, (v) => v.length === 0);
  });
</script>

<section class="section tab-price">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box overflow-x-auto">
          <table class="table is-narrow is-fullwidth is-light-border is-hoverable">
            <thead>
              <tr>
                <th />
                <th>Commodity Name</th>
                <th>Last Date</th>
                <th class="has-text-right">Last Price</th>
                <th class="has-text-right">1 Day</th>
                <th class="has-text-right">1 Week</th>
                <th class="has-text-right">4 Weeks</th>
                <th class="has-text-right">1 Year</th>
                <th class="has-text-right">3 Years</th>
                <th class="has-text-right">5 Years</th>
                <th>Commodity Type</th>
                <th>Commodity ID</th>
              </tr>
            </thead>
            <tbody class="has-text-grey-dark">
              {#each Object.keys(prices) as commodity}
                {@const p = prices[commodity][0]}
                <Toggleable>
                  <tr
                    class={active ? "is-active" : ""}
                    style="cursor: pointer;"
                    slot="toggle"
                    let:active
                    let:onclick
                    on:click={(e) => onclick(e)}
                  >
                    <td>
                      <span class="icon has-text-link">
                        <i
                          class="fas {active ? 'fa-chevron-up' : 'fa-chevron-down'}"
                          aria-hidden="true"
                        />
                      </span>
                    </td>

                    <td>{p.commodity_name}</td>
                    <td class="whitespace-nowrap">{p.date.format("DD MMM YYYY")}</td>
                    <td class="has-text-right">{formatCurrency(p.value, 4)}</td>
                    <td class="has-text-right"
                      ><ValueChange value={change(prices[commodity], 1, 0)} /></td
                    >
                    <td class="has-text-right"
                      ><ValueChange value={change(prices[commodity], 7, 2)} /></td
                    >
                    <td class="has-text-right"
                      ><ValueChange value={change(prices[commodity], 28, 4)} />
                    </td>
                    <td class="has-text-right"
                      ><ValueChange value={change(prices[commodity], 365, 7)} />
                    </td>
                    <td class="has-text-right"
                      ><ValueChange value={change(prices[commodity], 365 * 3, 7)} /></td
                    >
                    <td class="has-text-right"
                      ><ValueChange value={change(prices[commodity], 365 * 5, 7)} /></td
                    >
                    <td>{p.commodity_type}</td>
                    <td>{p.commodity_id}</td>
                  </tr>
                  <tr slot="content">
                    <td colspan="10" />
                    <td colspan="2" class="p-0">
                      <div>
                        <VirtualList
                          width="100%"
                          height={_.min([ITEM_SIZE * prices[commodity].length, ITEM_SIZE * 20])}
                          itemCount={prices[commodity].length}
                          itemSize={ITEM_SIZE}
                        >
                          <div
                            slot="item"
                            let:index
                            let:style
                            {style}
                            class="small-box is-flex is-flex-wrap-wrap is-justify-content-space-between is-size-7"
                          >
                            {@const p = prices[commodity][index]}
                            <div class="pl-1">{p.date.format("DD MMM YYYY")}</div>
                            <div class="pr-1 has-text-right">
                              {formatCurrency(p.value, 4)}
                            </div>
                          </div>
                        </VirtualList>
                      </div>
                    </td>
                  </tr>
                </Toggleable>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</section>
