<script lang="ts">
  import { onMount } from "svelte";
  import { initializeTheme } from "$lib/theme";
  import { theme } from "../store";

  onMount(() => initializeTheme(theme.set));
</script>

<slot />
