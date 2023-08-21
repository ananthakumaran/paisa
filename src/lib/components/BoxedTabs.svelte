<script lang="ts">
  export let options: { label: string; value: any }[];
  export let value: any;
</script>

<div class="du-tabs du-tabs-boxed">
  {#each options as option}
    <a
      class="du-tab du-tab-sm {option.value === value ? 'du-tab-active' : ''}"
      on:click={() => (value = option.value)}
    >
      {option.label}
    </a>
  {/each}
</div>
