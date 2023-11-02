<script lang="ts">
  export let size = 256;
  export let animation = false;
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  width={size}
  height={size}
  viewBox="0 0 24 24"
  class:animation
  ><g transform="rotate(90 12 12)">
    <path
      fill="currentColor"
      d="M 12 5 C 7.031 5 2 6.546 2 9.5 S 7.031 14 12 14 C 16.97 14 22 12.454 22 9.5 S 16.97 5 12 5 Z"
    />

    <path
      fill="currentColor"
      d="M 7 14.938 V 17.938 C 8.237 18.237 9.605 18.42 11 18.479 V 15.479 A 21.166 21.166 0 0 1 7 14.938 Z"
    />
    <path
      fill="currentColor"
      d="M 13 15.478 V 18.478 A 20.994 20.994 0 0 0 17 17.937 V 14.937 A 20.994 20.994 0 0 1 13 15.478 Z"
    />

    <path
      fill="currentColor"
      d="M 19 14.297 V 17.297 C 20.801 16.542 22 15.44 22 14 V 11 C 22 12.44 20.801 13.542 19 14.297 Z"
    />

    <path
      fill="currentColor"
      d="M 5 17.297 V 14.297 C 3.2 13.542 2 12.439 2 11 V 14 C 2 15.439 3.2 16.542 5 17.297 Z"
    />
  </g>

  {#if animation}
    <g
      transform="translate(2.6,5.5),scale(0.8),rotate(120,12,12),matrix(0.16741001,-0.09665704,0.04832852,0.083705,7.3774358,11.493257)"
    >
      <path d="M 0,70 A 65,70 0 0 0 65,0 5,5 0 0 1 75,0 75,70 0 0 1 0,70 Z">
        <animateTransform
          attributeName="transform"
          type="rotate"
          from="360 0 0"
          to="0 0 0"
          dur="0.6s"
          repeatCount="indefinite"
        />
      </path>
    </g>
  {/if}
</svg>

<style lang="scss">
  @import "bulma/sass/utilities/_all.sass";

  svg.animation {
    filter: drop-shadow(-1px -1px 1px rgba($black, 0.4));
  }
</style>
