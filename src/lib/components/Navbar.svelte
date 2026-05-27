<script lang="ts">
  import { page } from "$app/stores";
  import Actions from "$lib/components/Actions.svelte";
  import { month, year, dateMax, dateMin, dateRangeOption } from "../../store";
  import {
    cashflowExpenseDepth,
    cashflowExpenseDepthAllowed,
    cashflowIncomeDepth,
    cashflowIncomeDepthAllowed,
    obscure
  } from "../../persisted_store";
  import _ from "lodash";
  // Aliased to `t` to avoid colliding with lodash's `_`. Use `$t('key')` in markup.
  import { _ as t } from "$lib/i18n";
  import { calendarYear, forEachCalendarYear, helpUrl, isMobile, now } from "$lib/utils";
  import { resolveSelectedLinks, type Link } from "$lib/navbar_links";
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import DateRange from "./DateRange.svelte";
  import ThemeSwitcher from "./ThemeSwitcher.svelte";
  import MonthPicker from "./MonthPicker.svelte";
  import Logo from "./Logo.svelte";
  import InputRange from "./InputRange.svelte";
  export let isBurger: boolean = null;
  const readonly = USER_CONFIG.readonly;

  onMount(async () => {
    if (get(year) == "") {
      year.set(calendarYear(now()));
    }
  });

  // `label` here is an i18n key resolved at render time via `$t(label)`.
  // Keeping the tree static lets routing / preload logic stay declarative
  // while the strings still react to locale changes.
  const RecurringIcons = [
    { icon: "fa-circle-check", color: "success", label: "navbar.recurring_icon.cleared" },
    {
      icon: "fa-circle-check",
      color: "warning-dark",
      label: "navbar.recurring_icon.cleared_late"
    },
    {
      icon: "fa-exclamation-triangle",
      color: "danger",
      label: "navbar.recurring_icon.past_due"
    },
    { icon: "fa-circle-check", color: "grey", label: "navbar.recurring_icon.upcoming" }
  ];

  const links: Link[] = [
    { label: "nav.dashboard", href: "/", hide: true },
    {
      label: "nav.cash_flow",
      href: "/cash_flow",
      children: [
        { label: "nav.income_statement", href: "/income_statement", calendarYearPicker: true },
        { label: "nav.monthly", href: "/monthly", dateRangeSelector: true },
        {
          label: "nav.yearly",
          href: "/yearly",
          calendarYearPicker: true,
          maxDepthSelector: true
        },
        {
          label: "nav.recurring",
          href: "/recurring",
          help: "recurring",
          monthPicker: true,
          recurringIcons: true
        }
      ]
    },
    {
      label: "nav.expenses",
      href: "/expense",
      children: [
        { label: "nav.monthly", href: "/monthly", monthPicker: true, dateRangeSelector: true },
        { label: "nav.yearly", href: "/yearly", calendarYearPicker: true },
        { label: "nav.budget", href: "/budget", help: "budget", monthPicker: true }
      ]
    },
    {
      label: "nav.assets",
      href: "/assets",
      children: [
        { label: "nav.balance", href: "/balance" },
        { label: "nav.networth", href: "/networth", dateRangeSelector: true },
        { label: "nav.investment", href: "/investment" },
        { label: "nav.gain", href: "/gain" },
        { label: "nav.allocation", href: "/allocation", help: "allocation-targets" },
        { label: "nav.analysis", href: "/analysis", tag: "alpha", help: "analysis" }
      ]
    },
    {
      label: "nav.liabilities",
      href: "/liabilities",
      children: [
        { label: "nav.balance", href: "/balance" },
        { label: "nav.credit_cards", href: "/credit_cards", help: "credit-cards" },
        { label: "nav.repayment", href: "/repayment" },
        { label: "nav.interest", href: "/interest" }
      ]
    },
    { label: "nav.income", href: "/income" },
    {
      label: "nav.ledger",
      href: "/ledger",
      children: [
        { label: "nav.import", href: "/import", help: "import" },
        { label: "nav.editor", href: "/editor", help: "editor", disablePreload: true },
        { label: "nav.transactions", href: "/transaction", help: "bulk-edit" },
        { label: "nav.postings", href: "/posting" },
        { label: "nav.price", href: "/price" }
      ]
    },
    {
      label: "nav.more",
      href: "/more",
      children: [
        { label: "nav.configuration", href: "/config", help: "config" },
        { label: "nav.sheets", href: "/sheets", help: "sheets", disablePreload: true },
        { label: "nav.goals", href: "/goals", help: "goals" },
        { label: "nav.doctor", href: "/doctor" },
        { label: "nav.logs", href: "/logs" }
      ]
    }
  ];

  const tax = {
    label: "nav.tax",
    href: "/tax",
    help: "tax",
    children: [
      { label: "nav.harvest", href: "/harvest", help: "tax-harvesting" },
      { label: "nav.capital_gains", href: "/capital_gains", help: "capital-gains" },
      {
        label: "nav.schedule_al",
        href: "/schedule_al",
        help: "schedule-al",
        calendarYearPicker: true
      }
    ]
  };

  if (USER_CONFIG.default_currency == "INR") {
    _.last(links).children.push(tax);
  }

  const about = { label: "nav.about", href: "/about" };
  _.last(links).children.push(about);

  let selectedLink: Link = null;
  let selectedSubLink: Link = null;
  let selectedSubSubLink: Link = null;

  $: normalizedPath = $page.url.pathname?.replace(/(.+)\/$/, "");

  $: if (normalizedPath) {
    ({ selectedLink, selectedSubLink, selectedSubSubLink } = resolveSelectedLinks(
      links,
      normalizedPath
    ));
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
              class:is-active={normalizedPath == link.href}>{$t(link.label)}</a
            >
          {/if}
        {:else}
          <div class="navbar-item has-dropdown is-hoverable">
            <a
              class="navbar-link"
              class:is-active={normalizedPath.startsWith(link.href)}
              on:click|preventDefault={(e) =>
                isMobile() && e.currentTarget.parentElement.classList.toggle("is-active")}
              >{$t(link.label)}</a
            >
            <div class="navbar-dropdown {!isMobile() && 'is-boxed'}">
              {#each link.children as sublink}
                {@const href = link.href + sublink.href}
                {#if _.isEmpty(sublink.children)}
                  <a
                    class="navbar-item"
                    {href}
                    data-sveltekit-preload-data={sublink.disablePreload ? "tap" : "hover"}
                    class:is-active={normalizedPath.startsWith(href)}>{$t(sublink.label)}</a
                  >
                {:else}
                  <div class="nested has-dropdown navbar-item">
                    <a
                      class="navbar-link is-arrowless is-flex is-justify-content-space-between is-active"
                      class:is-active={normalizedPath.startsWith(href)}
                    >
                      <span>{$t(sublink.label)}</span>
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
                            >{$t(subsublink.label)}</a
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
                data-tippy-content="<p>{$t('common.readonly_tooltip')}</p>"
                >{$t("common.readonly")}</span
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
          <a class="is-inactive">{$t(selectedLink.label)}</a>
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
            <a class="is-inactive">{$t(selectedSubLink.label)}</a>

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
              <a class="is-inactive">{$t(selectedSubSubLink.label)}</a>
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
          <div data-tippy-content="<p>{$t(icon.label)}</p>">
            <span class="icon is-small has-text-{icon.color}">
              <i class={"fas " + icon.icon} />
            </span>
            <span class="is-hidden-mobile">{$t(icon.label)}</span>
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
              label={$t("navbar.depth_selector.expenses")}
              bind:value={$cashflowExpenseDepth}
              allowed={$cashflowExpenseDepthAllowed}
            />
            <InputRange
              label={$t("navbar.depth_selector.income")}
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

    {#if selectedSubSubLink?.calendarYearPicker || selectedSubLink?.calendarYearPicker || selectedLink?.calendarYearPicker}
      <div class="has-text-centered">
        <div class="select is-small">
          <select bind:value={$year}>
            {#each forEachCalendarYear($dateMin, $dateMax).reverse() as cy}
              <option>{calendarYear(cy)}</option>
            {/each}
          </select>
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
