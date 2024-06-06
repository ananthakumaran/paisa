İşte svelte bileşeni: <script lang="ts">
  import _ from "lodash";

  export let baslik: string;
  export let deger: string;
  export let kucuk: boolean = false;
  export let dar: boolean = false;
  export let renk: string = null;
  export let altBaslik: string = null;
</script>

<div class="level-item {dar && 'is-narrow'} has-text-left" class:kucuk>
  <div>
    <p class="heading">{baslik}</p>
    {#if renk}
      <p class="title" style="padding: 5px; color: {renk};">{deger}</p>
    {:else}
      <p class="title has-text-grey-dark">{deger}</p>
    {/if}
    {#if !_.isEmpty(altBaslik)}
      <div class="sub-title">{@html altBaslik}</div>
    {/if}
  </div>
</div>

<style lang="scss">
  @import "bulma/sass/utilities/_all.sass";

  .level-item.kucuk {
    .title {
      font-size: 1.25rem !important;
      line-height: 1.5rem !important;
    }
  }

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

  @include widescreen {
    .title {
      font-size: 2.25rem !important;
      line-height: 2.5rem !important;
    }

    .level-item.kucuk {
      .title {
        font-size: 1.5rem !important;
        line-height: 1.75rem !important;
      }
    }
  }

  .sub-title {
    font-weight: normal !important;
    font-size: 0.75rem !important;
  }
</style>