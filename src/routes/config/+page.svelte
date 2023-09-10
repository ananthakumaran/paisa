<script lang="ts">
  import { ajax } from "$lib/utils";
  import { onMount } from "svelte";
  import type { JSONSchema7 } from "json-schema";
  import JsonSchemaForm from "$lib/components/JsonSchemaForm.svelte";

  let lastConfig: UserConfig;
  let config: UserConfig;
  let schema: JSONSchema7;
  onMount(async () => {
    ({ config, schema } = await ajax("/api/config"));
    lastConfig = config;
  });
</script>

<div class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        {#if schema}
          <JsonSchemaForm key="config" bind:value={config} {schema} />
        {/if}
      </div>
    </div>
  </div>
</div>
