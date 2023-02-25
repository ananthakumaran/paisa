<script lang="ts">
  import { ajax } from "$lib/utils";

  let isLoading = false;

  async function sync(request: Record<string, any>) {
    isLoading = true;
    try {
      await ajax("/api/sync", { method: "POST", body: JSON.stringify(request) });
    } finally {
      isLoading = false;
      window.location.reload();
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
    </div>
  </div>
</div>
