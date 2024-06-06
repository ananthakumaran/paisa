<script lang="ts">
  import { renderBudget } from "$lib/budget";
  import { iconify } from "$lib/icon";
  import type { Action } from "svelte/action";
  import { firstName, formatCurrency, restName, type AccountBudget, tooltip } from "$lib/utils";
  import _ from "lodash";

  export let compact = false;
  export let accountBudget: AccountBudget;

  function canShow(accountBudget: AccountBudget): boolean {
    return accountBudget.forecast !== 0 || accountBudget.actual !== 0;
  }

  function color(accountBudget: AccountBudget): string {
    if (accountBudget.available == 0) {
      return "info";
    } else if (accountBudget.available > 0) {
      return "success";
    } else {
      return "danger";
    }
  }

  const chart: Action<HTMLElement, { ab: AccountBudget }> = (element, props) => {
    renderBudget(element, props.ab);
    return {};
  };

  const tooltipContent = tooltip(
    accountBudget.expenses.map((e) => {
      return [
        e.date.format("DD MMM YYYY"),
        [e.payee, "is-clipped"],
        [formatCurrency(e.amount), "has-text-weight-bold has-text-right"]
      ];
    })
  );
</script>

<div
  class="budget-card box px-2 pt-2 pb-2 my-3 has-background-white"
  data-tippy-content={_.isEmpty(accountBudget.expenses) ? null : tooltipContent}
>
  <div class="md:flex is-justify-content-space-between">
    <div
      class="has-text-weight-bold has-text-grey ml-2 truncate custom-icon"
      title={accountBudget.account}
    >
      {iconify(restName(accountBudget.account), { group: firstName(accountBudget.account) })}
    </div>
    <div
      class="is-flex is-justify-content-flex-end mr-2 is-align-items-center"
      style="min-width: fit-content"
    >
      {#if !compact}
        <div class="mr-3">
          <span class="budget-label mr-1">Bütçe</span>
          <span class="budget-amount">{formatCurrency(accountBudget.forecast)}</span>
        </div>
        <div class="mr-3">
          <span class="budget-label mr-1">Harcanan</span>
          <span class="budget-amount">{formatCurrency(accountBudget.actual)}</span>
        </div>
      {/if}
      {#if !compact && accountBudget.rollover != 0}
        <div class="mr-3">
          <span class="budget-label mr-1">Devir</span>
          <span class="budget-amount warn">{formatCurrency(accountBudget.rollover)}</span>
        </div>
      {/if}
      <div>
        <span class="budget-label mr-1"
          >{accountBudget.available >= 0 ? "Kullanılabilir" : "Aşılım"}</span
        >
        <span class="budget-amount {color(accountBudget)}"
          >{formatCurrency(Math.abs(accountBudget.available))}</span
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