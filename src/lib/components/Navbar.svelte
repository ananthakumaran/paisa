<script lang="ts">
  import { page } from "$app/stores";
  import Actions from "$lib/components/Actions.svelte";
  import {
    month,
    year,
    viewMode,
    selectedMonths,
    dateMax,
    dateMin,
    dateRangeOption
  } from "../../store";
  import {
    cashflowExpenseDepth,
    cashflowExpenseDepthAllowed,
    cashflowIncomeDepth,
    cashflowIncomeDepthAllowed,
    obscure
  } from "../../persisted_store";
  import _ from "lodash";
  import { financialYear, forEachFinancialYear, helpUrl, isMobile, now } from "$lib/utils";
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import DateRange from "./DateRange.svelte";
  import ThemeSwitcher from "./ThemeSwitcher.svelte";
  import MonthPicker from "./MonthPicker.svelte";
  import MultiSelectMonthPicker from "./MultiSelectMonthPicker.svelte";
  import Logo from "./Logo.svelte";
  import InputRange from "./InputRange.svelte";
  export let isBurger: boolean = null;
  const readonly = USER_CONFIG.readonly;

  onMount(async () => {
    if (get(year) == "") {
      year.set(financialYear(now()));
    }
  });

  // Auto-select all months when switching to monthly view
  $: if ($viewMode === "monthly" && $selectedMonths.length === 0) {
    selectedMonths.set(["01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"]);
  }

  const RecurringIcons = [
    { icon: "fa-circle-check", color: "success", label: "Cleared" },
    { icon: "fa-circle-check", color: "warning-dark", label: "Cleared late" },
    { icon: "fa-exclamation-triangle", color: "danger", label: "Past due" },
    { icon: "fa-circle-check", color: "grey", label: "Upcoming" }
  ];

  interface Link {
    label: string;
    href: string;
    tag?: string;
    help?: string;
    hide?: boolean;
    dateRangeSelector?: boolean;
    monthPicker?: boolean;
    financialYearPicker?: boolean;
    customYearMonthPicker?: boolean;
    maxDepthSelector?: boolean;
    recurringIcons?: boolean;
    children?: Link[];
    disablePreload?: boolean;
  }
  const links: Link[] = [
    { label: "Dashboard", href: "/", hide: true },
    {
      label: "Cash Flow",
      href: "/cash_flow",
      children: [
        { label: "Income Statement", href: "/income_statement", customYearMonthPicker: true },
        { label: "Monthly", href: "/monthly", dateRangeSelector: true },
        {
          label: "Yearly",
          href: "/yearly",
          financialYearPicker: true,
          maxDepthSelector: true
        },
        {
          label: "Recurring",
          href: "/recurring",
          help: "recurring",
          monthPicker: true,
          recurringIcons: true
        }
      ]
    },
    {
      label: "Expenses",
      href: "/expense",
      children: [
        { label: "Monthly", href: "/monthly", monthPicker: true, dateRangeSelector: true },
        { label: "Yearly", href: "/yearly", financialYearPicker: true },
        { label: "Budget", href: "/budget", help: "budget", monthPicker: true }
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
        { label: "Credit Cards", href: "/credit_cards", help: "credit-cards" },
        { label: "Repayment", href: "/repayment" },
        { label: "Interest", href: "/interest" }
      ]
    },
    { label: "Income", href: "/income" },
    {
      label: "Ledger",
      href: "/ledger",
      children: [
        { label: "Import", href: "/import", help: "import" },
        { label: "Editor", href: "/editor", help: "editor", disablePreload: true },
        { label: "Transactions", href: "/transaction", help: "bulk-edit" },
        { label: "Postings", href: "/posting" },
        { label: "Price", href: "/price" }
      ]
    },
    {
      label: "More",
      href: "/more",
      children: [
        { label: "Configuration", href: "/config", help: "config" },
        { label: "Sheets", href: "/sheets", help: "sheets", disablePreload: true },
        { label: "Goals", href: "/goals", help: "goals" },
        { label: "Doctor", href: "/doctor" },
        { label: "Logs", href: "/logs" }
      ]
    }
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
    _.last(links).children.push(tax);
  }

  const about = { label: "About", href: "/about" };
  _.last(links).children.push(about);

  let selectedLink: Link = null;
  let selectedSubLink: Link = null;
  let selectedSubSubLink: Link = null;

  $: normalizedPath = $page.url.pathname?.replace(/(.+)\/$/, "");

  $: if (normalizedPath) {
    selectedSubLink = null;
    selectedSubSubLink = null;
    selectedLink = _.find(links, (l) => normalizedPath == l.href);
    if (!selectedLink) {
      selectedLink = _.find(
        links,
        (l) => !_.isEmpty(l.children) && normalizedPath.startsWith(l.href)
      );

      selectedSubLink = _.find(
        selectedLink.children,
        (l) => normalizedPath == selectedLink.href + l.href
      );

      if (!selectedSubLink) {
        selectedSubLink = _.find(selectedLink.children, (l) =>
          normalizedPath.startsWith(selectedLink.href + l.href)
        );

        if (!_.isEmpty(selectedSubLink.children)) {
          selectedSubSubLink = _.find(selectedSubLink.children, (l) =>
            normalizedPath.startsWith(selectedLink.href + selectedSubLink.href + l.href)
          );
        }
      }
    }
  }
</script>

<nav class="navbar px-2 is-transparent" aria-label="main navigation">
  <div class="navbar-brand">
    <a
      href="/"
      class:is-active={normalizedPath == "/"}
      class="navbar-item is-size-4 has-text-weight-medium"
    >
      {#if $obscure}
        <span class="icon is-small is-size-5">
          <i class="fas fa-user-secret" />
        </span><span class="ml-2 is-primary-color">Paisa</span>
      {:else}
        <Logo size={22} /><span class="ml-1 is-primary-color">Paisa</span>
      {/if}
    </a>
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
          {#if !link.hide}
            <a
              class="navbar-item"
              href={link.href}
              data-sveltekit-preload-data={link.disablePreload ? "tap" : "hover"}
              class:is-active={normalizedPath == link.href}>{link.label}</a
            >
          {/if}
        {:else}
          <div class="navbar-item has-dropdown is-hoverable">
            <a
              class="navbar-link"
              class:is-active={normalizedPath.startsWith(link.href)}
              on:click|preventDefault={(e) =>
                isMobile() && e.currentTarget.parentElement.classList.toggle("is-active")}
              >{link.label}</a
            >
            <div class="navbar-dropdown {!isMobile() && 'is-boxed'}">
              {#each link.children as sublink}
                {@const href = link.href + sublink.href}
                {#if _.isEmpty(sublink.children)}
                  <a
                    class="navbar-item"
                    {href}
                    data-sveltekit-preload-data={sublink.disablePreload ? "tap" : "hover"}
                    class:is-active={normalizedPath.startsWith(href)}>{sublink.label}</a
                  >
                {:else}
                  <div class="nested has-dropdown navbar-item">
                    <a
                      class="navbar-link is-arrowless is-flex is-justify-content-space-between is-active"
                      class:is-active={normalizedPath.startsWith(href)}
                    >
                      <span>{sublink.label}</span>
                      <span class="icon is-small">
                        <i
                          class="fas {isMobile() ? 'fa-angle-down' : 'fa-angle-right'}"
                          aria-hidden="true"
                        ></i>
                      </span>
                    </a>

                    <div class="dropdown-menu">
                      <div class="dropdown-content">
                        {#each sublink.children as subsublink}
                          <a
                            href={href + subsublink.href}
                            class="navbar-item"
                            data-sveltekit-preload-data={subsublink.disablePreload
                              ? "tap"
                              : "hover"}
                            class:is-active={normalizedPath == href + subsublink.href}
                            >{subsublink.label}</a
                          >
                        {/each}
                      </div>
                    </div>
                  </div>
                {/if}
              {/each}
            </div>
          </div>
        {/if}
      {/each}
    </div>
    <div class="navbar-end" style="margin-right: 0.3em">
      <div class="navbar-item">
        <div class="field is-grouped">
          {#if readonly}
            <p class="control">
              <span
                class="mt-1 tag is-rounded is-danger is-light invertable"
                data-tippy-content="<p>Paisa is in readonly mode</p>">readonly</span
              >
            </p>
          {/if}

          <p class="control">
            <ThemeSwitcher />
          </p>
          <p class="control">
            <Actions />
          </p>
        </div>
      </div>
    </div>
  </div>
</nav>

<div class="mt-3 px-3 is-flex is-justify-content-space-between">
  {#if selectedLink}
    <nav
      style="margin-left: 0.73rem;"
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
              <a style="margin-left: -10px;" class="p-0" href={helpUrl(selectedSubLink.help)}
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

        {#if selectedSubLink}
          {#if selectedSubSubLink}
            <li>
              <a class="is-inactive">{selectedSubSubLink.label}</a>
            </li>
          {:else if selectedLink.href + selectedSubLink.href != normalizedPath}
            <li>
              <a class="is-inactive">{decodeURIComponent(_.last(normalizedPath.split("/")))}</a>
            </li>
          {/if}
        {/if}
      </ul>
    </nav>
  {/if}

  <div class="mr-3 is-flex" style="gap: 12px">
    {#if selectedSubLink?.recurringIcons}
      <div class="flex gap-5 items-center has-text-grey">
        {#each RecurringIcons as icon}
          <div data-tippy-content="<p>{icon.label}</p>">
            <span class="icon is-small has-text-{icon.color}">
              <i class={"fas " + icon.icon} />
            </span>
            <span class="is-hidden-mobile">{icon.label}</span>
          </div>
        {/each}
      </div>
    {/if}

    {#if selectedSubLink?.maxDepthSelector && ($cashflowExpenseDepthAllowed.max > 1 || $cashflowIncomeDepthAllowed.max > 1)}
      <div class="dropdown is-right is-hoverable">
        <div class="dropdown-trigger">
          <button class="button is-small" aria-haspopup="true">
            <span class="icon is-small">
              <i class="fas fa-sliders" />
            </span>
          </button>
        </div>
        <div class="dropdown-menu" role="menu">
          <div class="dropdown-content px-2 py-2">
            <InputRange
              label="Expenses"
              bind:value={$cashflowExpenseDepth}
              allowed={$cashflowExpenseDepthAllowed}
            />
            <InputRange
              label="Income"
              bind:value={$cashflowIncomeDepth}
              allowed={$cashflowIncomeDepthAllowed}
            />
          </div>
        </div>
      </div>
    {/if}

    {#if selectedSubLink?.dateRangeSelector || selectedLink?.dateRangeSelector}
      <div>
        <DateRange bind:value={$dateRangeOption} dateMin={$dateMin} dateMax={$dateMax} />
      </div>
    {/if}

    {#if selectedSubLink?.monthPicker || selectedLink?.monthPicker}
      <MonthPicker bind:value={$month} max={$dateMax} min={$dateMin} />
    {/if}

    {#if selectedSubSubLink?.financialYearPicker || selectedSubLink?.financialYearPicker || selectedLink?.financialYearPicker}
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

    {#if selectedSubSubLink?.customYearMonthPicker || selectedSubLink?.customYearMonthPicker || selectedLink?.customYearMonthPicker}
      <div class="has-text-centered">
        <div class="field has-addons">
          <!-- Multi-select Month Dropdown (positioned on the left, only shown when monthly is selected) -->
          {#if $viewMode === "monthly"}
            <div class="control">
              <MultiSelectMonthPicker
                bind:selectedMonths={$selectedMonths}
                on:change={(e) => selectedMonths.set(e.detail)}
              />
            </div>
          {/if}

          <!-- View Mode Dropdown (Yearly/Monthly) -->
          <div class="control">
            <div class="select is-small">
              <select bind:value={$viewMode}>
                <option value="yearly">Yearly</option>
                <option value="monthly">Monthly</option>
              </select>
            </div>
          </div>

          <!-- Year Dropdown -->
          <div class="control">
            <div class="select is-small">
              <select bind:value={$year}>
                {#each forEachFinancialYear($dateMin, $dateMax).reverse() as fy}
                  <option>{financialYear(fy)}</option>
                {/each}
              </select>
            </div>
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<style lang="scss">
  li a span.icon {
    margin-top: -5px;
  }
</style>
