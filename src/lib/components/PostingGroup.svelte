<script lang="ts">
  import { formatCurrency, type Posting } from "$lib/utils";

  export let postings: Posting[];
  export let groupFormat: string;

  interface GroupedPosting {
    key: string;
    postings: Posting[];
    total: number;
  }

  let groupedPostings: GroupedPosting[] = [];
  $: groupedPostings = group(postings);

  function group(ps: Posting[]) {
    let groupedPostings: GroupedPosting[] = [];
    let lastGroup: string;
    for (const posting of ps) {
      const group = posting.date.format(groupFormat);
      if (group !== lastGroup) {
        groupedPostings.push({
          key: group,
          postings: [],
          total: 0
        });
        lastGroup = group;
      }

      groupedPostings[groupedPostings.length - 1].postings.push(posting);
      let amount = posting.amount;
      if (posting.account.startsWith("Income:CapitalGains")) {
        amount = -amount;
      }
      groupedPostings[groupedPostings.length - 1].total += amount;
    }

    if (ps.length == 100) {
      groupedPostings.pop();
    }

    return groupedPostings;
  }
</script>

<div>
  {#each groupedPostings as groupedPosting}
    <div class="mb-3">
      <div class="flex justify-between -mb-1 has-text-weight-bold has-text-grey-light">
        <div>{groupedPosting.key}</div>
        <div>{formatCurrency(groupedPosting.total)}</div>
      </div>
      <slot groupedPostings={groupedPosting.postings} />
    </div>
  {/each}
</div>
