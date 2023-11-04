<script lang="ts">
  import _ from "lodash";

  export let title: string;
  export let value: string;
  export let narrow: boolean = false;
  export let color: string = null;
  export let subtitle: string = null;
</script>

<div class="level-item {narrow && 'is-narrow'} has-text-left">
  <div>
    <p class="heading">{title}</p>
    {#if color}
      <p class="title" style="padding: 5px; color: {color};">{value}</p>
    {:else}
      <p class="title has-text-grey-dark">{value}</p>
    {/if}
    {#if !_.isEmpty(subtitle)}
      <div class="sub-title">{subtitle}</div>
    {/if}
  </div>
</div>

<style lang="scss">
  @import "bulma/sass/utilities/_all.sass";

  .heading {
    font-weight: 400 !important;
    font-size: 1rem !important;
    text-transform: capitalize !important;
    letter-spacing: normal !important;
    margin-bottom: 0 !important;
  }

  .title {
    font-weight: 800 !important;
    font-size: 1.75rem !important;
    line-height: 2rem !important;
    padding-left: 0 !important;
    padding-bottom: 0 !important;
    margin-bottom: 0 !important;
  }

  @include desktop {
    .title {
      font-size: 2.25rem !important;
      line-height: 2.5rem !important;
    }
  }

  .sub-title {
    font-weight: normal !important;
    font-size: 0.75rem !important;
  }
</style>
