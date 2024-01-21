<script lang="ts">
  import CreditCardNetwork from "$lib/components/CreditCardNetwork.svelte";
  import DueDate from "$lib/components/DueDate.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import { iconText } from "$lib/icon";
  import {
    helpUrl,
    type CreditCardSummary,
    ajax,
    restName,
    formatCurrency,
    now,
    type CreditCardBill,
    formatPercentage
  } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";

  let isEmpty = false;
  let creditCards: CreditCardSummary[] = [];

  onMount(async () => {
    ({ creditCards } = await ajax("/api/credit_cards"));
    if (_.isEmpty(creditCards)) {
      isEmpty = true;
    }
  });

  function lastBill(creditCard: CreditCardSummary): CreditCardBill {
    return _.find(_.reverse(_.clone(creditCard.bills)), (b) => {
      return b.statementEndDate.isSameOrBefore(now());
    });
  }
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns flex-wrap">
      <div class="column is-12">
        <div class="credit-card-container">
          {#each creditCards as creditCard}
            {@const bill = lastBill(creditCard)}
            <div class="credit-card box p-3 m-0 flex-col justify-between">
              <div class="is-flex justify-between uppercase has-text-weight-bold is-size-5">
                <div style="margin: 35px 0 0 15px;" class="opacity-20 chip">
                  <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24"
                    ><path
                      fill="currentColor"
                      d="M10 4h10c1.11 0 2 .89 2 2v2h-3.41L16 10.59v4l-2 2V20h-4v-3.41l-2-2V9.41l2-2zm8 7.41V14h4v-4h-2.59zM6.59 8L8 6.59V4H4c-1.11 0-2 .89-2 2v2zM6 14v-4H2v4zm2 3.41L6.59 16H2v2c0 1.11.89 2 2 2h4zM17.41 16L16 17.41V20h4c1.11 0 2-.89 2-2v-2z"
                    /></svg
                  >
                </div>
                <div class="has-text-grey">
                  <span class="custom-icon">{iconText(creditCard.account)}</span>
                  <span class="ml-1">{restName(restName(creditCard.account))}</span>
                </div>
              </div>
              <div class="flex justify-between">
                <div class="flex flex-col">
                  {#if bill}
                    <div class="is-size-7">
                      <span class="has-text-grey">Amount Due</span>
                    </div>
                    <div>
                      <span class="is-size-4 has-text-grey-dark"
                        >{formatCurrency(bill.closingBalance)}</span
                      >
                    </div>
                    <div class="is-size-7 has-text-grey">
                      <DueDate dueDate={bill.dueDate} paidDate={bill.paidDate} />
                    </div>
                  {/if}
                </div>
                <div class="flex flex-col">
                  <div class="is-size-7">
                    <span class="has-text-grey">Balance</span>
                  </div>
                  <div class="flex flex-col">
                    <span class="is-size-4 has-text-grey-dark"
                      >{formatCurrency(creditCard.balance)}</span
                    >
                    <span class="is-size-7 has-text-grey"
                      >{formatPercentage(creditCard.balance / creditCard.creditLimit)} of {formatCurrency(
                        creditCard.creditLimit
                      )}
                    </span>
                  </div>
                </div>
              </div>
              <div class="is-flex justify-between items-end">
                <div class="opacity-25 has-text-weight-bold is-size-5">
                  * * * * &nbsp; {creditCard.number}
                </div>
                <div class="opacity-15">
                  <CreditCardNetwork size={48} name={creditCard.network} />
                </div>
              </div>
            </div>
          {/each}
        </div>
      </div>
    </div>
    <div class="columns flex-wrap">
      <div class="column is-12">
        <ZeroState item={!isEmpty}>
          <strong>Oops!</strong> You haven't configured any credit cards yet. Checkout the
          <a href={helpUrl("credit-card")}>docs</a> page to get started.
        </ZeroState>
      </div>
    </div>
  </div>
</section>
