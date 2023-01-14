<script lang="ts">
  import { ajax } from "$lib/utils";

  let isLoading = false;

  async function onClick() {
    isLoading = true;
    try {
      await ajax("/api/sync", { method: "POST" });
    } finally {
      isLoading = false;
      window.location.reload();
    }
  }
</script>

<button
  title="Sync journal, Update prices"
  class="button is-small"
  class:is-loading={isLoading}
  on:click={onClick}
>
  <span class="icon is-small">
    <i class="fas fa-sync" />
  </span>
</button>
