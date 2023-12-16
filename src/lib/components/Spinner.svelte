<script lang="ts">
  import { delayedLoading, loading } from "../../store";
  import Logo from "./Logo.svelte";
  let size = 90;
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
