<script lang="ts">
  import { sync } from "$lib/sync";
  import { isLoggedIn, isMobile, logout } from "$lib/utils";
  import { refresh } from "../../store";
  import { obscure } from "../../persisted_store";
  import { goto } from "$app/navigation";

  async function syncWithLoader(request: Record<string, any>) {
    try {
      await sync(request);
    } finally {
      refresh();
    }
  }

  const obscureId = "obscure";
  let last = $obscure;
  obscure.subscribe(() => {
    if ($obscure === last) return;

    refresh();
  });

  function doLogout() {
    logout();
    goto("/login");
  }

  let showLogout = isLoggedIn();
</script>

<div class="dropdown ml-2 is-hoverable {isMobile() ? 'is-left' : 'is-right'}">
  <div class="dropdown-trigger dropdown-icon">
    <button class="button is-large" aria-haspopup="true">
      <span class="icon">
        <i class="fas fa-ellipsis-vertical" />
      </span>
    </button>
  </div>
  <div class="dropdown-menu" id="dropdown-menu4" role="menu">
    <div class="dropdown-content">
      <a on:click={(_e) => syncWithLoader({ journal: true })} class="dropdown-item icon-text">
        <span class="icon is-small">
          <i class="fa-regular fa-file-lines" />
        </span>
        <span>Sync Journal</span>
      </a>
      <a on:click={(_e) => syncWithLoader({ prices: true })} class="dropdown-item icon-text">
        <span class="icon is-small">
          <i class="fas fa-dollar-sign" />
        </span>
        <span>Update Prices</span></a
      >
      <a on:click={(_e) => syncWithLoader({ portfolios: true })} class="dropdown-item icon-text">
        <span class="icon is-small">
          <i class="fas fa-layer-group" />
        </span>
        <span>Update Mutual Fund Portfolios</span></a
      >
      <hr class="dropdown-divider" />
      <a class="dropdown-item icon-text">
        <label for={obscureId} class="cursor-pointer w-full inline-block">
          <input bind:checked={$obscure} id={obscureId} type="checkbox" class="is-hidden" />
          <span class="ml-0 icon is-small">
            <i class="fas {$obscure ? 'fa-eye-slash' : 'fa-eye'}" />
          </span>
          <span>{$obscure ? "Show" : "Hide"} numbers</span>
        </label>
      </a>
      {#if showLogout}
        <hr class="dropdown-divider" />
        <a on:click={(_e) => doLogout()} class="dropdown-item icon-text">
          <span class="icon is-small">
            <i class="fas fa-arrow-right-from-bracket" />
          </span>
          <span>Logout</span>
        </a>
      {/if}
    </div>
  </div>
</div>
