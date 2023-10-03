<script lang="ts">
  import { sync } from "$lib/sync";
  import { refresh } from "../../store";

  let isLoading = false;

  async function syncWithLoader(request: Record<string, any>) {
    isLoading = true;
    try {
      await sync(request);
    } finally {
      isLoading = false;
      refresh();
    }
  }

  const obscureId = "obscure";
  let obscure = localStorage.getItem("obscure") === "true";

  $: {
    if (localStorage.getItem("obscure") !== obscure.toString()) {
      refresh();
      localStorage.setItem(obscureId, obscure.toString());
    }
  }
</script>

<div class="dropdown is-right" class:is-hoverable={!isLoading}>
  <div class="dropdown-trigger">
    <button class:is-loading={isLoading} class="button is-small" aria-haspopup="true">
      <span class="icon is-small">
        <i class="fas fa-caret-down" />
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
          <input bind:checked={obscure} id={obscureId} type="checkbox" class="is-hidden" />
          <span class="ml-0 icon is-small">
            <i class="fas {obscure ? 'fa-eye-slash' : 'fa-eye'}" />
          </span>
          <span>Hide numbers</span>
        </label>
      </a>
    </div>
  </div>
</div>
