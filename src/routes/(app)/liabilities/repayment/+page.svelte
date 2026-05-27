<script lang="ts">
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import LegendCard from "$lib/components/LegendCard.svelte";
  import {
    renderMonthlyRepaymentTimeline,
    renderAmortizationChart,
    type Amortization,
    type RepaymentResponse
  } from "$lib/repayment";
  import { ajax, formatCurrency, formatPercentage, type Legend } from "$lib/utils";
  import _ from "lodash";
  import { onMount, tick } from "svelte";
  import { _ as t } from "$lib/i18n";
  import { get } from "svelte/store";

  let isEmpty = false;
  let legends: Legend[] = [];
  let amortizations: Amortization[] = [];
  let amortizationLegends: Record<string, Legend[]> = {};

  function chartId(name: string) {
    return "d3-amortization-" + name.replace(/[^A-Za-z0-9]/g, "_");
  }

  function scheduleLabel(s: string) {
    const tt = get(t);
    return s === "equal_payment"
      ? tt("page.liabilities.schedule_equal_payment")
      : tt("page.liabilities.schedule_equal_principal");
  }

  onMount(async () => {
    const data = (await ajax("/api/liabilities/repayment")) as unknown as RepaymentResponse;
    const repayments = data.repayments || [];
    amortizations = data.amortizations || [];

    if (_.isEmpty(repayments) && _.isEmpty(amortizations)) {
      isEmpty = true;
      return;
    }

    if (!_.isEmpty(repayments)) {
      legends = renderMonthlyRepaymentTimeline(repayments);
    }

    if (!_.isEmpty(amortizations)) {
      // Wait for DOM to render the per-amortization svg containers.
      await tick();
      for (const a of amortizations) {
        amortizationLegends[a.name] = renderAmortizationChart("#" + chartId(a.name), a);
      }
    }
  });
</script>

<section class="section" class:is-hidden={!isEmpty}>
  <div class="container is-fluid">
    <div class="columns is-centered">
      <div class="column is-4 has-text-centered">
        <article class="message">
          <div class="message-body">{$t("page.liabilities.no_repayment")}</div>
        </article>
      </div>
    </div>
  </div>
</section>

<section class="section" class:is-hidden={isEmpty || legends.length === 0}>
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <LegendCard {legends} clazz="ml-4" />
          <svg id="d3-repayment-timeline" width="100%" height="500" />
        </div>
      </div>
    </div>
    <BoxLabel text={$t("page.liabilities.monthly_repayment_timeline")} />
  </div>
</section>

{#each amortizations as a (a.name)}
  <section class="section">
    <div class="container is-fluid">
      <div class="columns">
        <div class="column is-3">
          <div class="box">
            <p class="heading">{a.name}</p>
            <p class="title is-6">{scheduleLabel(a.schedule)}</p>
            <table class="table is-narrow is-fullwidth">
              <tbody>
                <tr>
                  <td>{$t("page.liabilities.principal")}</td>
                  <td class="has-text-right">{formatCurrency(a.principal)}</td>
                </tr>
                <tr>
                  <td>{$t("page.liabilities.apr")}</td>
                  <td class="has-text-right">{formatPercentage(a.apr / 100, 2)}</td>
                </tr>
                <tr>
                  <td>{$t("page.liabilities.term")}</td>
                  <td class="has-text-right"
                    >{a.term_months} {$t("page.liabilities.term_months_suffix")}</td
                  >
                </tr>
                {#if a.start_date}
                  <tr>
                    <td>{$t("page.liabilities.start")}</td>
                    <td class="has-text-right">{a.start_date}</td>
                  </tr>
                {/if}
                <tr>
                  <td>
                    {a.schedule === "equal_payment"
                      ? $t("page.liabilities.monthly_payment")
                      : $t("page.liabilities.first_payment")}
                  </td>
                  <td class="has-text-right">{formatCurrency(a.monthly_payment)}</td>
                </tr>
                <tr>
                  <td>{$t("page.liabilities.total_interest")}</td>
                  <td class="has-text-right">{formatCurrency(a.total_interest)}</td>
                </tr>
                <tr>
                  <td>{$t("page.liabilities.total_payment")}</td>
                  <td class="has-text-right">{formatCurrency(a.total_payment)}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <div class="column is-9">
          <div class="box">
            <LegendCard legends={amortizationLegends[a.name] || []} clazz="ml-4" />
            <svg id={chartId(a.name)} width="100%" height="500" />
          </div>
        </div>
      </div>
      <BoxLabel text={$t("page.liabilities.amortization_schedule") + " – " + a.name} />
    </div>
  </section>
{/each}
