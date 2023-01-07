<script lang="ts">
  import { formatCurrency, type FYCapitalGain } from "$lib/utils";
  const DATE_FORMAT = "DD MMM YYYY";

  export let fyCapitalGain: FYCapitalGain;
</script>

<div>
  <table class="table is-narrow is-fullwidth is-bordered">
    <thead>
      <tr>
        <th>Purchase Date</th>
        <th class="has-text-right">Purchase Price</th>
        <th>Sell Date</th>
        <th class="has-text-right">Sell Price</th>
        <th class="has-text-right">Gain</th>
        <th class="has-text-right">Taxable Gain</th>
        <th class="has-text-right">Short Term Tax</th>
        <th class="has-text-right">Long Term Tax</th>
      </tr>
    </thead>
    <tbody>
      {#each fyCapitalGain.posting_pairs as pp}
        <tr class="is-size-7">
          <td>{pp.purchase.date.format(DATE_FORMAT)}</td>
          <td class="has-text-right">{formatCurrency(pp.purchase.amount)}</td>
          <td>{pp.sell.date.format(DATE_FORMAT)}</td>
          <td class="has-text-right">{formatCurrency(-pp.sell.amount)}</td>
          <td class="has-text-right has-text-weight-bold">{formatCurrency(pp.gain)}</td>
          <td class="has-text-right has-text-weight-bold">{formatCurrency(pp.taxable_gain)}</td>
          <td class="has-text-right has-text-weight-bold">{formatCurrency(pp.short_term_tax)}</td>
          <td class="has-text-right has-text-weight-bold">{formatCurrency(pp.long_term_tax)}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
