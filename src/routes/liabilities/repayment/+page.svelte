<script lang="ts">
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import { renderMonthlyRepaymentTimeline } from "$lib/repayment";
  import { ajax } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  let isEmpty = false;

  onMount(async () => {
    const { repayments: repayments } = await ajax("/api/liabilities/repayment");
    if (_.isEmpty(repayments)) {
      isEmpty = true;
    } else {
      renderMonthlyRepaymentTimeline(repayments);
    }
  });
</script>

<section class="section" class:is-hidden={!isEmpty}>
  <div class="container is-fluid">
    <div class="columns is-centered">
      <div class="column is-4 has-text-centered">
        <article class="message">
          <div class="message-body">You haven't repaid any liabilities.</div>
        </article>
      </div>
    </div>
  </div>
</section>

<section class="section" class:is-hidden={isEmpty}>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <svg id="d3-repayment-timeline" width="100%" height="500" />
        </div>
      </div>
    </div>
    <BoxLabel text="Monthly Repayment Timeline" />
  </div>
</section>
