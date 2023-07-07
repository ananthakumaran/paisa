<script lang="ts">
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
        <article class="message is-success">
          <div class="message-body">
            <strong>Hurray!</strong> You have repaid any liabilities.
          </div>
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
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div>
          <p class="heading">Monthly Repayment Timeline</p>
        </div>
      </div>
    </div>
  </div>
</section>
