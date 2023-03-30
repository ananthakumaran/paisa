<script lang="ts">
  import { page } from "$app/stores";
  import Sync from "$lib/components/Sync.svelte";
  import _ from "lodash";
  export let isBurger: boolean = null;

  interface Link {
    label: string;
    href: string;
    tag?: string;
    help?: string;
    children?: Link[];
  }
  const links: Link[] = [
    { label: "Overview", href: "/" },
    { label: "Expenses", href: "/expense" },
    {
      label: "Assets",
      href: "/assets",
      children: [
        { label: "Balance", href: "/balance" },
        { label: "Investment", href: "/investment" },
        { label: "Gain", href: "/gain" },
        { label: "Allocation", href: "/allocation", help: "allocation-targets" },
        { label: "Analysis", href: "/analysis", tag: "alpha" }
      ]
    },
    {
      label: "Liabilities",
      href: "/liabilities",
      children: [
        { label: "Balance", href: "/balance" },
        { label: "Interest", href: "/interest" }
      ]
    },
    { label: "Income", href: "/income" },
    {
      label: "Ledger",
      href: "/ledger",
      children: [
        { label: "Transactions", href: "/transaction" },
        { label: "Journal", href: "/journal" },
        { label: "Price", href: "/price" }
      ]
    },
    {
      label: "Tax",
      href: "/tax",
      help: "tax",
      children: [
        { label: "Harvest", href: "/harvest", help: "tax-harvesting" },
        { label: "Capital Gains", href: "/capital_gains", help: "capital-gains" },
        { label: "Schedule AL", href: "/schedule_al", help: "schedule-al" }
      ]
    },
    { label: "Retirement", href: "/retirement/progress", help: "retirement" },
    { label: "Doctor", href: "/doctor" }
  ];

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
    }
  }
</script>

<nav class="navbar is-transparent" aria-label="main navigation">
  <div class="navbar-brand">
    <span class="navbar-item is-size-4 has-text-weight-medium">₹ Paisa</span>
    <a
      role="button"
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
                <a class="navbar-item" {href} class:is-active={$page.url.pathname === href}
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
            <Sync />
          </p>
        </div>
      </div>
    </div>
  </div>
</nav>

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
          <a
            style="margin-left: -10px;"
            class="p-0"
            href={`https://ananthakumaran.in/paisa/${selectedLink.help}.html`}
            ><span class="icon is-small">
              <i class="fas fa-question fa-border" />
            </span></a
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
            <span style="font-size: 0.6rem" class="tag is-rounded is-warning"
              >{selectedSubLink.tag}</span
            >
          {/if}
        </li>
      {/if}
    </ul>
  </nav>
{/if}

<style lang="scss">
  .breadcrumb.has-chevron-separator li + li::before {
    content: "";
    color: hsl(229deg, 53%, 53%);
  }

  a.is-inactive {
    color: hsl(0deg, 0%, 21%);
    cursor: default;
    pointer-events: none;
  }
</style>
