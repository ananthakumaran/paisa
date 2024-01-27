<script lang="ts">
  import { iconText } from "$lib/icon";
  import {
    formatCurrency,
    formatPercentage,
    restName,
    type CreditCardSummary,
    now,
    type CreditCardBill
  } from "$lib/utils";
  import _ from "lodash";
  import CreditCardNetwork from "./CreditCardNetwork.svelte";
  import DueDate from "./DueDate.svelte";

  export let creditCard: CreditCardSummary;

  function lastBill(creditCard: CreditCardSummary): CreditCardBill {
    return _.find(_.reverse(_.clone(creditCard.bills)), (b) => {
      return b.statementEndDate.isSameOrBefore(now());
    });
  }

  $: bill = lastBill(creditCard);
</script>

<div class="credit-card box p-3 m-0 flex-col justify-between">
  <div class="is-flex justify-between has-text-weight-bold is-size-5">
    <div style="margin: 35px 0 0 15px;" class="flex items-center opacity-20">
      <svg
        class="chip"
        xmlns="http://www.w3.org/2000/svg"
        width="36"
        height="36"
        viewBox="0 0 24 24"
        ><path
          fill="currentColor"
          d="M10 4h10c1.11 0 2 .89 2 2v2h-3.41L16 10.59v4l-2 2V20h-4v-3.41l-2-2V9.41l2-2zm8 7.41V14h4v-4h-2.59zM6.59 8L8 6.59V4H4c-1.11 0-2 .89-2 2v2zM6 14v-4H2v4zm2 3.41L6.59 16H2v2c0 1.11.89 2 2 2h4zM17.41 16L16 17.41V20h4c1.11 0 2-.89 2-2v-2z"
        /></svg
      >
      <svg
        class="nfc ml-1"
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="24"
        viewBox="0 0 24 24"
        ><path
          fill="none"
          stroke="currentColor"
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M6 8.32a7.43 7.43 0 0 1 0 7.36m3.46-9.47a11.76 11.76 0 0 1 0 11.58M12.91 4.1a15.91 15.91 0 0 1 .01 15.8M16.37 2a20.16 20.16 0 0 1 0 20"
        /></svg
      >
    </div>
    <div>
      <a
        class="secondary-link has-text-grey"
        href="/liabilities/credit_cards/{encodeURIComponent(creditCard.account)}"
      >
        <span class="custom-icon">{iconText(creditCard.account)}</span>
        <span class="ml-1">{restName(restName(creditCard.account))}</span>
      </a>
    </div>
  </div>
  <div class="flex justify-between">
    <div class="flex flex-col">
      {#if bill}
        <div class="is-size-7">
          <span class="has-text-grey">Amount Due</span>
        </div>
        <div>
          <span class="is-size-4 has-text-grey-dark">{formatCurrency(bill.closingBalance)}</span>
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
        <span class="is-size-4 has-text-grey-dark">{formatCurrency(creditCard.balance)}</span>
        <span class="is-size-7 has-text-grey"
          >{formatPercentage(creditCard.balance / creditCard.creditLimit)} of {formatCurrency(
            creditCard.creditLimit
          )}
        </span>
      </div>
    </div>
  </div>
  <div class="is-flex justify-between items-end">
    <div class="has-text-weight-bold is-size-5 inline-flex items-center">
      <span class="opacity-40 inline-flex flex-col mr-2" style="font-size: 0.5rem; line-height: 1;">
        <span>VALID</span>
        <span>THRU</span>
      </span>
      <span class="opacity-30"
        >{creditCard.expirationDate.format("MM / YY")} &nbsp; &nbsp; &nbsp; * * * * &nbsp; {creditCard.number}</span
      >
    </div>
    <div class="opacity-15">
      <CreditCardNetwork size={48} name={creditCard.network} />
    </div>
  </div>
</div>
