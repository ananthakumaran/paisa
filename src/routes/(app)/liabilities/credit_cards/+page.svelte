<script lang="ts">
  import CreditCardCard from "$lib/components/CreditCardCard.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import { ajax, helpUrl, type CreditCardSummary } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  import { _ as t } from "$lib/i18n";

  let isEmpty = false;
  let creditCards: CreditCardSummary[] = [];

  onMount(async () => {
    ({ creditCards } = await ajax("/api/credit_cards"));
    if (_.isEmpty(creditCards)) {
      isEmpty = true;
    }
  });
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns flex-wrap">
      <div class="column is-12">
        <div class="credit-card-container">
          {#each creditCards as creditCard}
            <CreditCardCard {creditCard} />
          {/each}
        </div>
      </div>
    </div>
    <div class="columns flex-wrap">
      <div class="column is-12">
        <ZeroState item={!isEmpty}>
          <strong>{$t("common.oops")}</strong>
          {$t("page.liabilities.no_credit_cards_pre")}<a href={helpUrl("credit-card")}
            >{$t("page.liabilities.docs")}</a
          >{$t("page.liabilities.no_credit_cards_post")}
        </ZeroState>
      </div>
    </div>
  </div>
</section>
