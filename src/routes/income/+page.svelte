<script lang="ts">
  import COLORS from "$lib/colors";
  import {
    renderMonthlyInvestmentTimeline,
    renderYearlyIncomeTimeline,
    renderYearlyTimelineOf
  } from "$lib/income";
  import { ajax, formatCurrency, setHtml } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  onMount(async () => {
    const {
      income_timeline: incomes,
      tax_timeline: taxes,
      yearly_cards: yearlyCards
    } = await ajax("/api/income");
    renderMonthlyInvestmentTimeline(incomes);
    renderYearlyIncomeTimeline(yearlyCards);
    renderYearlyTimelineOf("Net Income", "net_income", COLORS.gain, yearlyCards);
    renderYearlyTimelineOf("Net Tax", "net_tax", COLORS.loss, yearlyCards);

    const grossIncome = _.sumBy(incomes, (i) => _.sumBy(i.postings, (p) => -p.amount));

    const netTax = _.sumBy(taxes, (t) => _.sumBy(t.postings, (p) => p.amount));

    setHtml("gross-income", formatCurrency(grossIncome), COLORS.gainText);
    setHtml("net-tax", formatCurrency(netTax), COLORS.lossText);
  });
</script>

<section class="section tab-income">
  <div class="container">
    <nav class="level">
      <div class="level-item has-text-centered">
        <div>
          <p class="heading">Gross Income</p>
          <p class="d3-gross-income title" />
        </div>
      </div>
      <div class="level-item has-text-centered">
        <div>
          <p class="heading">Net Tax</p>
          <p class="d3-net-tax title" />
        </div>
      </div>
    </nav>
  </div>
</section>
<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <svg id="d3-income-timeline" width="100%" height="500" />
        </div>
      </div>
    </div>
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div>
          <p class="heading">Monthly Income Timeline</p>
        </div>
      </div>
    </div>
  </div>
</section>
<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-one-third">
        <div class="box px-3">
          <svg id="d3-yearly-income-timeline" width="100%" />
        </div>
      </div>
      <div class="column is-one-third">
        <div class="box px-3">
          <svg id="d3-yearly-net_income-timeline" width="100%" />
        </div>
      </div>
      <div class="column is-one-third">
        <div class="box px-3">
          <svg id="d3-yearly-net_tax-timeline" width="100%" />
        </div>
      </div>
    </div>
    <div class="columns">
      <div class="column is-12 has-text-centered">
        <div>
          <p class="heading">Financial Year Income Timeline</p>
        </div>
      </div>
    </div>
  </div>
</section>
