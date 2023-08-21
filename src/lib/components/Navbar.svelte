<script lang="ts">
  import { page } from "$app/stores";
  import Sync from "$lib/components/Sync.svelte";
  import { month, year, dateMax, dateMin, dateRangeOption, cashflowType } from "../../store";
  import _ from "lodash";
  import { financialYear, forEachFinancialYear, helpUrl } from "$lib/utils";
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import dayjs from "dayjs";
  import DateRange from "./DateRange.svelte";
  import ThemeSwitcher from "./ThemeSwitcher.svelte";
  import BoxedTabs from "./BoxedTabs.svelte";
  export let isBurger: boolean = null;

  onMount(async () => {
    if (get(year) == "") {
      year.set(financialYear(dayjs()));
    }
  });

  interface Link {
    label: string;
    href: string;
    tag?: string;
    help?: string;
    dateRangeSelector?: boolean;
    monthPicker?: boolean;
    cashflowTypePicker?: boolean;
    financialYearPicker?: boolean;
    children?: Link[];
  }
  const links: Link[] = [
    { label: "Dashboard", href: "/" },
    {
      label: "Cash Flow",
      href: "/cash_flow",
      children: [
        { label: "Monthly", href: "/monthly", dateRangeSelector: true },
        { label: "Yearly", href: "/yearly", cashflowTypePicker: true, financialYearPicker: true },
        { label: "Recurring", href: "/recurring", tag: "alpha", help: "recurring" }
      ]
    },
    {
      label: "Expenses",
      href: "/expense",
      children: [
        { label: "Monthly", href: "/monthly", monthPicker: true, dateRangeSelector: true },
        { label: "Yearly", href: "/yearly", financialYearPicker: true }
      ]
    },
    {
      label: "Assets",
      href: "/assets",
      children: [
        { label: "Balance", href: "/balance" },
        { label: "Networth", href: "/networth", dateRangeSelector: true },
        { label: "Investment", href: "/investment" },
        { label: "Gain", href: "/gain" },
        { label: "Allocation", href: "/allocation", help: "allocation-targets" },
        { label: "Analysis", href: "/analysis", tag: "alpha", help: "analysis" }
      ]
    },
    {
      label: "Liabilities",
      href: "/liabilities",
      children: [
        { label: "Balance", href: "/balance" },
        { label: "Repayment", href: "/repayment" },
        { label: "Interest", href: "/interest" }
      ]
    },
    { label: "Income", href: "/income" },
    {
      label: "Ledger",
      href: "/ledger",
      children: [
        { label: "Import", href: "/import", tag: "alpha", help: "import" },
        { label: "Editor", href: "/editor", tag: "alpha" },
        { label: "Transactions", href: "/transaction" },
        { label: "Postings", href: "/posting" },
        { label: "Price", href: "/price" },
        { label: "Doctor", href: "/doctor" }
      ]
    },
    { label: "Retirement", href: "/retirement/progress", help: "retirement" }
  ];

  const tax = {
    label: "Tax",
    href: "/tax",
    help: "tax",
    children: [
      { label: "Harvest", href: "/harvest", help: "tax-harvesting" },
      { label: "Capital Gains", href: "/capital_gains", help: "capital-gains" },
      {
        label: "Schedule AL",
        href: "/schedule_al",
        help: "schedule-al",
        financialYearPicker: true
      }
    ]
  };

  if (USER_CONFIG.default_currency == "INR") {
    links.push(tax);
  }

  let selectedLink: Link = null;
  let selectedSubLink: Link = null;

  $: if ($page.url.pathname) {
    selectedSubLink = null;
    selectedLink = _.find(links, (l) => $page.url.pathname == l.href);
    if (!selectedLink) {
      selectedLink = _.find(
        links,
        (l) => !_.isEmpty(l.children) && $page.url.pathname.startsWith(l.href)
      );

      selectedSubLink = _.find(
        selectedLink.children,
        (l) => $page.url.pathname == selectedLink.href + l.href
      );

      if (!selectedSubLink) {
        selectedSubLink = _.find(selectedLink.children, (l) =>
          $page.url.pathname.startsWith(selectedLink.href + l.href)
        );
      }
    }
  }
</script>

<nav class="navbar px-3 is-transparent" aria-label="main navigation">
  <div class="navbar-brand">
    <span class="navbar-item is-size-4 has-text-weight-medium">₹ Paisa</span>
    <a
      role="button"
      tabindex="-1"
      class="navbar-burger"
      class:is-active={isBurger === true}
      on:click|preventDefault={(_e) => (isBurger = !isBurger)}
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
      {#each links as link}
        {#if _.isEmpty(link.children)}
          <a class="navbar-item" href={link.href} class:is-active={$page.url.pathname == link.href}
            >{link.label}</a
          >
        {:else}
          <div class="navbar-item has-dropdown is-hoverable">
            <a class="navbar-link" class:is-active={$page.url.pathname.startsWith(link.href)}
              >{link.label}</a
            >
            <div class="navbar-dropdown is-boxed">
              {#each link.children as sublink}
                {@const href = link.href + sublink.href}
                <a class="navbar-item" {href} class:is-active={$page.url.pathname.startsWith(href)}
                  >{sublink.label}</a
                >
              {/each}
            </div>
          </div>
        {/if}
      {/each}
    </div>
    <div class="navbar-end">
      <div class="navbar-item">
        <div class="field is-grouped">
          <p class="control">
            <ThemeSwitcher />
          </p>
          <p class="control">
            <Sync />
          </p>
        </div>
      </div>
    </div>
  </div>
</nav>

<div class="mt-2 px-3 is-flex is-justify-content-space-between">
  {#if selectedLink}
    <nav
      style="margin-left: 12px;"
      class="breadcrumb has-chevron-separator mb-0 is-small"
      aria-label="breadcrumbs"
    >
      <ul>
        <li>
          <a class="is-inactive">{selectedLink.label}</a>
          {#if selectedLink.help}
            <a style="margin-left: -10px;" class="p-0" href={helpUrl(selectedLink.help)}
              ><span class="icon is-small">
                <i class="fas fa-question fa-border" />
              </span></a
            >
          {/if}

          {#if selectedLink.tag}
            <span style="font-size: 0.6rem" class="tag is-rounded is-warning"
              >{selectedLink.tag}</span
            >
          {/if}
        </li>
        {#if selectedSubLink}
          <li>
            <a class="is-inactive">{selectedSubLink.label}</a>

            {#if selectedSubLink.help}
              <a
                style="margin-left: -10px;"
                class="p-0"
                href={`https://ananthakumaran.in/paisa/${selectedSubLink.help}.html`}
                ><span class="icon is-small">
                  <i class="fas fa-question fa-border" />
                </span></a
              >
            {/if}

            {#if selectedSubLink.tag}
              <span style="font-size: 0.6rem" class="tag is-rounded is-warning mr-2"
                >{selectedSubLink.tag}</span
              >
            {/if}
          </li>
        {/if}

        {#if selectedSubLink && selectedLink.href + selectedSubLink.href != $page.url.pathname}
          <li>
            <a class="is-inactive">{decodeURIComponent(_.last($page.url.pathname.split("/")))}</a>
          </li>
        {/if}
      </ul>
    </nav>
  {/if}

  <div class="mr-3 is-flex" style="gap: 12px">
    {#if selectedSubLink?.cashflowTypePicker}
      <BoxedTabs
        options={[
          { label: "Flat", value: "flat" },
          { label: "Hierachy", value: "hierachy" }
        ]}
        bind:value={$cashflowType}
      />
    {/if}

    {#if selectedSubLink?.dateRangeSelector || selectedLink?.dateRangeSelector}
      <div>
        <DateRange bind:value={$dateRangeOption} dateMin={$dateMin} dateMax={$dateMax} />
      </div>
    {/if}

    {#if selectedSubLink?.monthPicker || selectedLink?.monthPicker}
      <div class="has-text-centered">
        <input
          style="width: 125px"
          class="input is-small"
          required
          type="month"
          id="d3-current-month"
          bind:value={$month}
          max={$dateMax.format("YYYY-MM")}
          min={$dateMin.format("YYYY-MM")}
        />
      </div>
    {/if}

    {#if selectedSubLink?.financialYearPicker || selectedLink?.financialYearPicker}
      <div class="has-text-centered">
        <div class="select is-small">
          <select bind:value={$year}>
            {#each forEachFinancialYear($dateMin, $dateMax).reverse() as fy}
              <option>{financialYear(fy)}</option>
            {/each}
          </select>
        </div>
      </div>
    {/if}
  </div>
</div>
