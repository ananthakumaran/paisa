<script lang="ts">
  import { loading } from "../../store";
</script>

<div>
  {#if $loading}
    <div class="circle-container">
      <div class="circle" />
      <svg xmlns="http://www.w3.org/2000/svg" version="1.0" viewBox="40 -1 170 250" height="40"
        ><path
          class="svg-text-grey-dark"
          d="M153 23h41l15-23H55L40 23h26c27 0 52 2 62 25H55L40 71h91v1c0 17-14 43-60 43H48v22l90 113h41L85 133c39-2 75-24 80-62h29l15-23h-45c-1-9-5-18-11-25z"
        /></svg
      >
    </div>
  {/if}
  <div style={$loading ? "display: none" : ""}>
    <slot />
  </div>
</div>

<style lang="scss">
  @import "bulma/sass/utilities/_all.sass";

  svg {
    position: absolute;
    left: 15px;
    top: 10px;
  }

  .circle-container {
    margin: auto;
    position: absolute;
    top: 0;
    left: 0;
    bottom: 0;
    right: 0;
    height: 100px;
    width: 100px;
  }

  .circle {
    height: 60px;
    width: 60px;
    border-color: $grey-dark $grey-dark transparent transparent;
    border-width: 3px;
    border-style: solid;
    border-image: initial;
    border-radius: 50%;
    animation: 0.75s linear 0s infinite normal none running rotate;
  }

  @keyframes rotate {
    0% {
      transform: rotate(0);
    }
    100% {
      transform: rotate(360deg);
    }
  }
</style>
