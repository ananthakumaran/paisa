<script lang="ts">
  import { ajax } from "$lib/utils";
  import { onMount } from "svelte";
  import type { JSONSchema7 } from "json-schema";
  import JsonSchemaForm from "$lib/components/JsonSchemaForm.svelte";
  import _ from "lodash";
  import * as toast from "bulma-toast";

  let lastConfig: UserConfig;
  let config: UserConfig;
  let schema: JSONSchema7;
  let hasChanges = true;
  let isLoading = false;
  let error: string = null;
  onMount(async () => {
    ({ config, schema } = await ajax("/api/config"));
    lastConfig = _.cloneDeep(config);
  });

  async function save() {
    isLoading = true;
    try {
      let success = false;
      ({ success, error } = await ajax("/api/config", {
        method: "POST",
        body: JSON.stringify(config)
      }));

      if (success) {
        lastConfig = _.cloneDeep(config);
        globalThis.USER_CONFIG = _.cloneDeep(config);
        toast.toast({
          message: `Saved config`,
          type: "is-success"
        });
      }
    } finally {
      isLoading = false;
    }
  }

  $: hasChanges = !_.isEqual(config, lastConfig);
</script>

<div class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        {#if schema}
          <div class="box px-3" style="max-width: 1024px;">
            {#if error}
              <article class="message is-danger">
                <div class="message-body" style="overflow: auto; white-space: pre;">
                  {error}
                </div>
              </article>
            {/if}
            <div class="field is-grouped is-grouped-right">
              <div class="control">
                <button
                  on:click={(_e) => save()}
                  class="button is-success {isLoading && 'is-loading'}"
                  disabled={!hasChanges}>Submit</button
                >
              </div>
              <div class="control">
                <button
                  on:click={(_e) => (config = _.cloneDeep(lastConfig))}
                  class="button is-light">Cancel</button
                >
              </div>
            </div>
            <JsonSchemaForm key="config" bind:value={config} {schema} />
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>
