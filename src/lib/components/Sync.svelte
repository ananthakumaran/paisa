<script lang="ts">
  import { ajax } from "$lib/utils";
  import { refresh } from "../../store";

  let isLoading = false;

  async function sync(request: Record<string, any>) {
    isLoading = true;
    try {
      await ajax("/api/sync", { method: "POST", body: JSON.stringify(request) });
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
        <i class="fas fa-gear" />
      </span>
    </button>
  </div>
  <div class="dropdown-menu" id="dropdown-menu4" role="menu">
    <div class="dropdown-content">
      <a on:click={(_e) => sync({ journal: true })} class="dropdown-item">Sync Journal</a>
      <a on:click={(_e) => sync({ prices: true })} class="dropdown-item">Update Prices</a>
      <a on:click={(_e) => sync({ portfolios: true })} class="dropdown-item"
        >Update Mutual Fund Portfolios</a
      >
      <hr class="dropdown-divider" />
      <span class="dropdown-item">
        <input
          bind:checked={obscure}
          id={obscureId}
          type="checkbox"
          class="switch is-rounded is-link is-small"
        />
        <label for={obscureId} class="is-size-6" style="padding-bottom: 2px;">
          <span class="icon is-small mx-2">
            <i class="fas {obscure ? 'fa-eye-slash' : 'fa-eye'}" />
          </span>
          Hide numbers
        </label>
      </span>
    </div>
  </div>
</div>
