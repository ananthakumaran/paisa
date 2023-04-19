<script lang="ts">
  import { iconify } from "$lib/icon";
  import { formatCurrency, formatFloat, type Posting } from "$lib/utils";

  const unlessINR = (p: Posting) => {
    if (p.commodity == "INR") {
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

<table
  style="margin-bottom: -1px; margin-top: -1px"
  class="table is-fullwidth is-narrow is-striped is-hoverable"
>
  <tbody>
    {#each postings as p}
      <tr>
        <td>{iconify(p.account)}</td>
        <td class="has-text-right">{unlessINR(p)}</td>
        <td class="has-text-right">{formatCurrency(p.amount, 2)}</td>
      </tr>
    {/each}
  </tbody>
</table>
