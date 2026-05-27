<script lang="ts">
  import { ajax, configUpdated } from "$lib/utils";
  import { onMount } from "svelte";
  import type { JSONSchema7 } from "json-schema";
  import JsonSchemaForm from "$lib/components/JsonSchemaForm.svelte";
  import _ from "lodash";
  import * as toast from "bulma-toast";
  import { refresh } from "../../../../store";
  import { sync } from "$lib/sync";
  import { _ as t } from "$lib/i18n";
  import { get } from "svelte/store";

  let lastConfig: UserConfig;
  let config: UserConfig;
  let schema: JSONSchema7;
  let hasChanges = true;
  let isLoading = false;
  let error: string = null;
  let accounts: string[] = [];
  onMount(async () => {
    ({ config, schema, accounts } = await ajax("/api/config"));
    lastConfig = _.cloneDeep(config);
  });

  async function resetToDefault() {
    if (confirm(get(t)("page.more.config.reset_confirm"))) {
      save({
        journal_path: lastConfig.journal_path,
        db_path: lastConfig.db_path
      } as any);
    }
  }

  async function save(newConfig: UserConfig) {
    isLoading = true;
    try {
      let success = false;
      ({ success, error } = await ajax("/api/config", {
        method: "POST",
        body: JSON.stringify(newConfig),
        background: true
      }));

      if (success) {
        lastConfig = _.cloneDeep(newConfig);
        config = _.cloneDeep(newConfig);
        globalThis.USER_CONFIG = _.cloneDeep(newConfig);
        configUpdated();
        refresh();
        toast.toast({
          message: `Saved config`,
          type: "is-success"
        });

        await sync({ journal: true });
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
            <article class="message">
              <div class="message-body">
                {@html $t("page.more.config.intro")}
              </div>
            </article>

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
                  on:click={(_e) => save(config)}
                  class="button is-success {isLoading && 'is-loading'}"
                  disabled={!hasChanges}>{$t("page.more.config.save")}</button
                >
              </div>
              <div class="control">
                <button
                  on:click={(_e) => (config = _.cloneDeep(lastConfig))}
                  class="button is-light">{$t("page.more.config.cancel")}</button
                >
              </div>
              <div class="control">
                <button on:click={(_e) => resetToDefault()} class="button is-danger"
                  >{$t("page.more.config.reset_defaults")}</button
                >
              </div>
            </div>
            <JsonSchemaForm
              allAccounts={accounts}
              key="configuration"
              bind:value={config}
              {schema}
            />
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>
