<script lang="ts">
  import { page } from "$app/stores";
  import { afterNavigate, beforeNavigate } from "$app/navigation";
  import { followCursor, type Instance, delegate } from "tippy.js";
  import _ from "lodash";
  import Sync from "$lib/components/Sync.svelte";
  import Spinner from "$lib/components/Spinner.svelte";

  let isBurger: boolean = null;
  let tippyInstances: Instance[] = [];

  beforeNavigate(() => {
    tippyInstances.forEach((t) => t.destroy());
  });

  afterNavigate(() => {
    isBurger = null;
    tippyInstances = delegate("section,nav", {
      target: "[data-tippy-content]",
      theme: "light",
      onShow: (instance) => {
        const content = instance.reference.getAttribute("data-tippy-content");
        if (!_.isEmpty(content)) {
          instance.setContent(content);
        } else {
          return false;
        }
      },
      maxWidth: "none",
      delay: 0,
      allowHTML: true,
      followCursor: true,
      plugins: [followCursor]
    });
  });
</script>

<nav class="navbar is-transparent" aria-label="main navigation">
  <div class="navbar-brand">
    <span class="navbar-item is-size-4 has-text-weight-medium">₹ Paisa</span>
    <a
      role="button"
      class="navbar-burger"
      class:is-active={isBurger === true}
      on:click={(_e) => (isBurger = !isBurger)}
      aria-label="menu"
      aria-expanded="false"
      data-target="navbarBasicExample"
    >
      <span aria-hidden="true" />
      <span aria-hidden="true" />
      <span aria-hidden="true" />
    </a>
  </div>

  <div class="navbar-menu" class:is-active={isBurger === true}>
    <div class="navbar-start">
      <a id="overview" class="navbar-item" href="/" class:is-active={$page.url.pathname === "/"}
        >Overview</a
      >
      <a
        id="expense"
        class="navbar-item"
        href="/expense"
        class:is-active={$page.url.pathname === "/expense"}>Expense</a
      >
      <div class="navbar-item has-dropdown is-hoverable">
        <a class="navbar-link" class:is-active={$page.url.pathname.startsWith("/assets/")}>Assets</a
        >
        <div class="navbar-dropdown is-boxed">
          <a
            id="investment"
            class="navbar-item"
            href="/assets/investment"
            class:is-active={$page.url.pathname === "/assets/investment"}>Investment</a
          >
          <a
            id="gain"
            class="navbar-item"
            href="/assets/gain"
            class:is-active={$page.url.pathname === "/assets/gain"}>Gain</a
          >
          <a
            id="allocation"
            class="navbar-item"
            href="/assets/allocation"
            class:is-active={$page.url.pathname === "/assets/allocation"}>Allocation</a
          >
        </div>
      </div>
      <div class="navbar-item has-dropdown is-hoverable">
        <a class="navbar-link" class:is-active={$page.url.pathname.startsWith("/liabilities/")}
          >Liabilities</a
        >
        <div class="navbar-dropdown is-boxed">
          <a
            class="navbar-item"
            href="/liabilities/interest"
            class:is-active={$page.url.pathname === "/liabilities/interest"}>Interest</a
          >
        </div>
      </div>
      <a
        id="income"
        class="navbar-item"
        href="/income"
        class:is-active={$page.url.pathname === "/income"}>Income</a
      >
      <div class="navbar-item has-dropdown is-hoverable">
        <a class="navbar-link" class:is-active={$page.url.pathname.startsWith("/ledger/")}>Ledger</a
        >
        <div class="navbar-dropdown is-boxed">
          <a
            id="holding"
            class="navbar-item"
            href="/ledger/holding"
            class:is-active={$page.url.pathname === "/ledger/holding"}>Holding</a
          >
          <a
            id="journal"
            class="navbar-item"
            href="/ledger/journal"
            class:is-active={$page.url.pathname === "/ledger/journal"}>Journal</a
          >
          <a
            id="price"
            class="navbar-item"
            href="/ledger/price"
            class:is-active={$page.url.pathname === "/ledger/price"}>Price</a
          >
        </div>
      </div>
      <div class="navbar-item has-dropdown is-hoverable">
        <a class="navbar-link" class:is-active={$page.url.pathname.startsWith("/tax/")}>Tax</a>
        <div class="navbar-dropdown is-boxed">
          <a
            id="harvest"
            class="navbar-item"
            href="/tax/harvest"
            class:is-active={$page.url.pathname === "/tax/harvest"}>Harvest</a
          >
          <a
            id="capital_gains"
            class="navbar-item"
            href="/tax/capital_gains"
            class:is-active={$page.url.pathname === "/tax/capital_gains"}>Capital Gains</a
          >
          <a
            id="schedule_al"
            class="navbar-item"
            href="/tax/schedule_al"
            class:is-active={$page.url.pathname === "/tax/schedule_al"}>Schedule AL</a
          >
        </div>
      </div>
      <div class="navbar-item has-dropdown is-hoverable">
        <a class="navbar-link" class:is-active={$page.url.pathname.startsWith("/retirement/")}
          >Retirement</a
        >
        <div class="navbar-dropdown is-boxed">
          <a
            id="retirement_progress"
            class="navbar-item"
            href="/retirement/progress"
            class:is-active={$page.url.pathname === "/retirement/progress"}>Progress</a
          >
        </div>
      </div>
      <a
        id="doctor"
        class="navbar-item"
        href="/doctor"
        class:is-active={$page.url.pathname === "/doctor"}>Doctor</a
      >
    </div>
    <div class="navbar-end">
      <div class="navbar-item">
        <div class="field is-grouped">
          <p class="control">
            <Sync />
          </p>
        </div>
      </div>
    </div>
  </div>
</nav>

<Spinner>
  <slot />
</Spinner>
