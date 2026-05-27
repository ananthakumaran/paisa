<script lang="ts">
  import { accountColorStyle } from "$lib/colors";
  import { iconText } from "$lib/icon";
  import { firstName, formatCurrency, formatFloatUptoPrecision, type Posting } from "$lib/utils";

  const unlessDefaultCurrency = (p: Posting) => {
    if (p.commodity == USER_CONFIG.default_currency) {
      return "";
    } else {
      return `${formatFloatUptoPrecision(p.quantity, 3)} ${
        p.commodity
      } @ ${formatFloatUptoPrecision(p.amount / p.quantity, 4)}`;
    }
  };

  // Render the amount with the posting's own commodity symbol — so a
  // USD line shows `$11.23`, HKD shows `HK$1,663.80`, and the default
  // CNY (or whatever the config says) shows `¥…`.
  const amountWithCurrency = (p: Posting) => formatCurrency(p.amount, p.commodity);

  export let postings: Posting[];
</script>

<div style="margin: 4px 0;">
  {#each postings as p}
    <div class="is-flex is-justify-content-space-between is-hoverable" style="margin: 1px 0;">
      <div class="truncate custom-icon" style="min-width: 100px;" title={p.account}>
        <span style={accountColorStyle(firstName(p.account))}>{iconText(p.account)}</span>
        {p.account}
      </div>
      <div class="is-flex is-align-items-baseline is-justify-content-right">
        <div class="has-text-right has-text-grey is-size-7 mr-2 truncate">
          {unlessDefaultCurrency(p)}
        </div>
        <div class="has-text-right" style="min-width: 50px;">
          {amountWithCurrency(p)}
        </div>
      </div>
    </div>
  {/each}
</div>
