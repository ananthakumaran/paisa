<script lang="ts">
  import { goto } from "$app/navigation";
  import COLORS from "$lib/colors";
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import CreditCardCard from "$lib/components/CreditCardCard.svelte";
  import DueDate from "$lib/components/DueDate.svelte";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import TransactionCard from "$lib/components/TransactionCard.svelte";
  import { renderYearlySpends } from "$lib/credit_cards";
  import { iconify } from "$lib/icon";
  import {
    ajax,
    formatCurrency,
    formatPercentage,
    type CreditCardBill,
    type CreditCardSummary
  } from "$lib/utils";
  import { MasonryGrid } from "@egjs/svelte-grid";
  import _, { now } from "lodash";
  import { onMount } from "svelte";
  import type { PageData } from "./$types";
  let UntypedMasonryGrid = MasonryGrid as any;

  export let data: PageData;
  let svg: SVGElement;

  let creditCard: CreditCardSummary;
  let currentBill: CreditCardBill;
  let found = false;
  let small = true;
  let rendered = false;

  function lastBill(creditCard: CreditCardSummary): CreditCardBill {
    return _.find(_.reverse(_.clone(creditCard.bills)), (b) => {
      return b.statementEndDate.isSameOrBefore(now());
    });
  }

  $: if (creditCard && svg && !rendered) {
    renderYearlySpends(svg, creditCard.yearlySpends);
    rendered = true;
  }

  onMount(async () => {
    ({ creditCard, found } = await ajax("/api/credit_cards/:account", null, data));
    if (!found) {
      return goto("/liabilities/credit_cards");
    }

    currentBill = lastBill(creditCard);
  });
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns flex-wrap">
      <div class="column is-3-widescreen is-4">
        {#if creditCard}
          <div class="flex mb-12">
            <CreditCardCard {creditCard} />
          </div>

          <nav class="level grid-2">
            <LevelItem
              narrow
              small
              title="Available Credit"
              color={COLORS.neutral}
              value={formatCurrency(Math.max(creditCard.creditLimit - creditCard.balance, 0))}
            />

            <LevelItem
              narrow
              small
              title="Credit Usage"
              color={COLORS.neutral}
              value={formatPercentage(creditCard.balance / creditCard.creditLimit, 2)}
            />
          </nav>

          <nav class="level grid-2">
            <LevelItem
              narrow
              small
              title="Statement Count"
              color={COLORS.neutral}
              value={creditCard.bills.length.toString()}
            />
            <LevelItem
              narrow
              small
              title="Transaction Count"
              color={COLORS.neutral}
              value={_.sumBy(creditCard.bills, (b) => b.transactions.length).toString()}
            />
          </nav>

          <div class="box px-3 py-0">
            <svg bind:this={svg} width="100%" />
          </div>
          <BoxLabel text="Year wise spends" />
        {/if}
      </div>
      <div class="column is-9-widescreen is-8">
        {#if currentBill}
          <div class="flex flex-wrap gap-4 mb-4">
            <div
              class="box py-2 m-0 flex-grow overflow-x-scroll"
              style="border: 1px solid transparent"
            >
              <div class="is-flex mr-2 is-align-items-baseline" style="min-width: fit-content">
                <div class="ml-3 custom-icon is-size-5 whitespace-nowrap">
                  <span>{iconify(creditCard.account)}</span>
                </div>
                <div class="ml-3 whitespace-nowrap">
                  <span class="mr-1 is-size-7 has-text-grey">Payment</span>
                  <span
                    ><DueDate dueDate={currentBill.dueDate} paidDate={currentBill.paidDate} /></span
                  >
                </div>
              </div>
            </div>
            <div class="has-text-right">
              <div class="select is-medium">
                <select bind:value={currentBill}>
                  {#each _.reverse(_.clone(creditCard.bills)) as bill}
                    <option value={bill}
                      >{bill.statementStartDate.format("DD MMM YYYY")} — {bill.statementEndDate.format(
                        "DD MMM YYYY"
                      )}</option
                    >
                  {/each}
                </select>
              </div>
            </div>
          </div>
          <nav class="level flex gap-4 overflow-x-scroll" style="justify-content: start;">
            <LevelItem
              {small}
              narrow
              title="Opening Balance"
              color={COLORS.neutral}
              value={formatCurrency(currentBill.openingBalance)}
            />
            <div class="level-item is-narrow">
              <span class="icon is-size-3">
                <i class="fas fa-plus" />
              </span>
            </div>
            <LevelItem
              {small}
              narrow
              title="Debits"
              color={COLORS.expenses}
              value={formatCurrency(currentBill.debits)}
            />
            <div class="level-item is-narrow">
              <span class="icon is-size-3">
                <i class="fas fa-minus" />
              </span>
            </div>
            <LevelItem
              {small}
              narrow
              title="Credits"
              color={COLORS.income}
              value={formatCurrency(currentBill.credits)}
            />
            <div class="level-item is-narrow">
              <span class="icon is-size-3">
                <i class="fas fa-equals" />
              </span>
            </div>
            <LevelItem
              {small}
              narrow
              title="Amount Due"
              color={COLORS.liabilities}
              value={formatCurrency(currentBill.closingBalance)}
            />
          </nav>

          <div>
            <UntypedMasonryGrid gap={10} maxStretchColumnSize={500} align="stretch">
              {#each currentBill.transactions as t}
                <div class="mr-3 is-flex-grow-1">
                  <TransactionCard {t} />
                </div>
              {/each}
            </UntypedMasonryGrid>
          </div>
        {/if}
      </div>
    </div>
  </div>
</section>
