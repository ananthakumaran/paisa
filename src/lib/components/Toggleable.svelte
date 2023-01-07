<script lang="ts">
  let active = false;
  let onclick = function (_e: any) {
    active = !active;
  };
</script>

<slot name="toggle" {active} {onclick} />
{#if active}
  <slot name="content" />
{/if}
