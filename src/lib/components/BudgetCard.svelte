<script lang="ts">
  import { renderBudget } from "$lib/budget";
  import { iconify } from "$lib/icon";
  import type { Action } from "svelte/action";
  import { firstName, formatCurrency, restName, type AccountBudget } from "$lib/utils";

  export let compact = false;
  export let accountBudget: AccountBudget;
  export let selected: boolean;

  function canShow(accountBudget: AccountBudget): boolean {
    return accountBudget.forecast !== 0 || accountBudget.actual !== 0;
  }

  const chart: Action<HTMLElement, { ab: AccountBudget }> = (element, props) => {
    renderBudget(element, props.ab);
    return {};
  };
</script>

<div
  on:click
  class="budget-card box px-2 pt-2 pb-2 my-2 has-background-white"
  class:is-selected={selected}
  class:cursor-pointer={!selected && !compact}
  style={compact ? "" : "width: 800px"}
>
  <div class="is-flex is-justify-content-space-between">
    <div class="has-text-weight-bold ml-2 truncate" title={accountBudget.account}>
      {iconify(restName(accountBudget.account), { group: firstName(accountBudget.account) })}
    </div>
    <div
      class="is-flex is-justify-content-flex-end mr-2 is-align-items-center"
      style="min-width: fit-content"
    >
      {#if !compact}
        <div class="mr-3">
          <span class="budget-label mr-1">Budget</span>
          <span class="budget-amount">{formatCurrency(accountBudget.forecast)}</span>
        </div>
        <div class="mr-3">
          <span class="budget-label mr-1">Used</span>
          <span class="budget-amount">{formatCurrency(accountBudget.actual)}</span>
        </div>
      {/if}
      {#if !compact && accountBudget.rollover != 0}
        <div class="mr-3">
          <span class="budget-label mr-1">Rollover</span>
          <span class="budget-amount warn">{formatCurrency(accountBudget.rollover)}</span>
        </div>
      {/if}
      <div>
        <span class="budget-label mr-1">Available</span>
        <span class="budget-amount {accountBudget.available >= 0 ? 'success' : 'danger'}"
          >{formatCurrency(accountBudget.available)}</span
        >
      </div>
    </div>
  </div>

  {#if canShow(accountBudget)}
    <div use:chart={{ ab: accountBudget }}>
      <svg height="10" width="100%"></svg>
    </div>
  {/if}
</div>
