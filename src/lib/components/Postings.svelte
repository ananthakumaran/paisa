<script lang="ts">
  import { accountColorStyle } from "$lib/colors";
  import { iconText } from "$lib/icon";
  import { firstName, formatCurrency, formatFloat, type Posting } from "$lib/utils";

  const unlessDefaultCurrency = (p: Posting) => {
    if (p.commodity == USER_CONFIG.default_currency) {
      return "";
    } else {
      return `${formatFloat(p.quantity, 3)} ${p.commodity} @ ${formatFloat(
        p.amount / p.quantity,
        4
      )}`;
    }
  };

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
          {formatCurrency(p.amount, 2)}
        </div>
      </div>
    </div>
  {/each}
</div>
