<script lang="ts">
  import { fade } from "svelte/transition";
  import { delayedLoading, delayedUnLoading } from "../../store";
  import Logo from "./Logo.svelte";
  let size = 90;
</script>

<div>
  <div style={$delayedUnLoading || $delayedLoading ? "visibility: hidden" : ""}>
    <slot />
  </div>
  {#if $delayedLoading}
    <div class="circle-container" transition:fade={{ duration: 400 }}>
      <Logo {size} animation />
    </div>
  {/if}
</div>

<style lang="scss">
  .circle-container {
    margin: auto;
    position: absolute;
    top: -45px;
    left: 0;
    bottom: 0;
    right: 0;
    height: 90px;
    width: 90px;
  }
</style>
