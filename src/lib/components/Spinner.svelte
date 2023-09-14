<script lang="ts">
  import { delayedLoading, loading } from "../../store";
  let size = 48;
</script>

<div>
  {#if $delayedLoading}
    <div class="circle-container">
      <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24">
        <g transform="rotate(90 12 12)">
          <path
            class="coin"
            d="M12 5C7.031 5 2 6.546 2 9.5S7.031 14 12 14c4.97 0 10-1.546 10-4.5S16.97 5 12 5zm-5 9.938v3c1.237.299 2.605.482 4 .541v-3a21.166 21.166 0 0 1-4-.541zm6 .54v3a20.994 20.994 0 0 0 4-.541v-3a20.994 20.994 0 0 1-4 .541zm6-1.181v3c1.801-.755 3-1.857 3-3.297v-3c0 1.44-1.199 2.542-3 3.297zm-14 3v-3C3.2 13.542 2 12.439 2 11v3c0 1.439 1.2 2.542 3 3.297z"
          />
        </g>

        <g
          transform="translate(2.6,5.5),scale(0.8),rotate(120,12,12),matrix(0.16741001,-0.09665704,0.04832852,0.083705,7.3774358,11.493257)"
        >
          <path class="comet" d="M 0,70 A 65,70 0 0 0 65,0 5,5 0 0 1 75,0 75,70 0 0 1 0,70 Z">
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
      </svg>
    </div>
  {/if}
  <div style={$loading ? "display: none" : ""}>
    <slot />
  </div>
</div>
