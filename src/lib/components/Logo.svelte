<script lang="ts">
  import COLORS from "$lib/colors";
  import _ from "lodash";

  export let size = 256;
  export let thickness = 4;
  export let animation = false;

  let colors = [COLORS.income, COLORS.expenses, COLORS.equity, COLORS.liabilities, COLORS.assets];
  let dotColor = COLORS.primary;
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  width={size}
  height={size}
  viewBox="0 0 24 24"
  class:animation
>
  <g>
    <g>
      <g transform="rotate(360, 12, 12)"
        ><path
          d="M12 12Z"
          fill="none"
          stroke={dotColor}
          stroke-width={thickness}
          stroke-linecap="round"
        /></g
      >
    </g>
    <g>
      {#each colors as color, i}
        <g transform="rotate({i * 72}, 12, 12)"
          ><path
            d="M 21.510565162951536 8.909830056250526 A 10 10 0 0 0 12 6"
            fill="none"
            stroke={color}
            stroke-width={thickness}
          /></g
        >
      {/each}

      {#each colors as color, i}
        <g transform="rotate({(i + 1) * 72}, 12, 12)"
          ><path
            d="M12.59 5.88Z"
            fill="none"
            stroke={color}
            stroke-width={thickness * 1.05}
            stroke-linecap="round"
          /></g
        >
      {/each}

      {#if animation}
        <animateTransform
          attributeName="transform"
          attributeType="XML"
          type="rotate"
          from="0 12 12"
          to="360 12 12"
          dur="1s"
          repeatCount="indefinite"
        />
      {/if}
    </g>
  </g>
</svg>
