<script lang="ts">
  import COLORS from "$lib/colors";
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import LegendCard from "$lib/components/LegendCard.svelte";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import {
    renderMonthlyInvestmentTimeline,
    renderYearlyIncomeTimeline,
    renderYearlyTimelineOf
  } from "$lib/income";
  import { ajax, formatCurrency, type Legend } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  import { _ as t } from "$lib/i18n";
  import { get } from "svelte/store";

  let grossIncome = 0;
  let netTax = 0;

  let monthlyInvestmentTimelineLegends: Legend[] = [];
  let yearlyIncomeTimelineLegends: Legend[] = [];
  let yearlyNetIncomeTimelineLegends: Legend[] = [];
  let yearlyNetTaxTimelineLegends: Legend[] = [];

  onMount(async () => {
    const {
      income_timeline: incomes,
      tax_timeline: taxes,
      yearly_cards: yearlyCards
    } = await ajax("/api/income");
    monthlyInvestmentTimelineLegends = renderMonthlyInvestmentTimeline(incomes);
    yearlyIncomeTimelineLegends = renderYearlyIncomeTimeline(yearlyCards);
    const tt = get(t);
    yearlyNetIncomeTimelineLegends = renderYearlyTimelineOf(
      tt("page.income.net_income"),
      "net_income",
      COLORS.gain,
      yearlyCards
    );
    yearlyNetTaxTimelineLegends = renderYearlyTimelineOf(
      tt("page.income.net_tax"),
      "net_tax",
      COLORS.loss,
      yearlyCards
    );

    grossIncome = _.sumBy(incomes, (i) => _.sumBy(i.postings, (p) => -p.amount));
    netTax = _.sumBy(taxes, (t) => _.sumBy(t.postings, (p) => p.amount));
  });
</script>

<section class="section tab-income">
  <div class="container">
    <nav class="level">
      <LevelItem
        title={$t("page.income.gross_income")}
        value={formatCurrency(grossIncome)}
        color={COLORS.gainText}
      />
      <LevelItem
        title={$t("page.income.net_tax")}
        value={formatCurrency(netTax)}
        color={COLORS.lossText}
      />
    </nav>
  </div>
</section>
<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <LegendCard legends={monthlyInvestmentTimelineLegends} clazz="ml-4" />
          <svg id="d3-income-timeline" width="100%" height="500" />
        </div>
      </div>
    </div>
    <BoxLabel text={$t("page.income.monthly_income_timeline")} />
  </div>
</section>
<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-one-third">
        <div class="box px-3">
          <LegendCard legends={yearlyIncomeTimelineLegends} clazz="ml-4" />
          <svg id="d3-yearly-income-timeline" width="100%" />
        </div>
      </div>
      <div class="column is-one-third">
        <div class="box px-3">
          <LegendCard legends={yearlyNetIncomeTimelineLegends} clazz="ml-4" />
          <svg id="d3-yearly-net_income-timeline" width="100%" />
        </div>
      </div>
      <div class="column is-one-third">
        <div class="box px-3">
          <LegendCard legends={yearlyNetTaxTimelineLegends} clazz="ml-4" />
          <svg id="d3-yearly-net_tax-timeline" width="100%" />
        </div>
      </div>
    </div>
    <BoxLabel text={$t("page.income.financial_year_income_timeline")} />
  </div>
</section>
