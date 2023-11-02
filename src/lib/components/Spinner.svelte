<script lang="ts">
  import { delayedLoading, loading } from "../../store";
  import Logo from "./Logo.svelte";
  let size = 48;
</script>

<div>
  {#if $delayedLoading}
    <div class="circle-container">
      <Logo {size} animation />
    </div>
  {/if}
  <div style={$loading ? "display: none" : ""}>
    <slot />
  </div>
</div>
