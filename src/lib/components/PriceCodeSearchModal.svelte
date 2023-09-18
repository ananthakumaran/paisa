<script lang="ts">
  import Select from "svelte-select";
  import Modal from "$lib/components/Modal.svelte";
  import _ from "lodash";
  import { createEventDispatcher, onMount } from "svelte";
  import { ajax, type AutoCompleteItem, type PriceProvider } from "$lib/utils";

  let label = "Choose Price Provider";
  export let open = false;
  let code = "";

  let providers: PriceProvider[] = [];
  let selectedProvider: PriceProvider = null;

  let filters: Record<string, AutoCompleteItem> = {};

  onMount(async () => {
    ({ providers } = await ajax("/api/price/providers"));
    selectedProvider = providers[0];
  });

  let isLoading = false;
  async function clearProviderCache() {
    isLoading = true;
    try {
      await ajax(
        "/api/price/providers/delete/:provider",
        { method: "POST" },
        { provider: selectedProvider.code }
      );
    } finally {
      isLoading = false;
    }
  }

  let autocompleteCache: number[] = [];
  function clearCache(i: number) {
    autocompleteCache[i] = (autocompleteCache[i] || 0) + 1;
  }

  function reset() {
    code = "";
    filters = {};
    for (let i = 0; i < _.max(_.map(providers, (p) => p.fields.length)); i++) {
      clearCache(i);
    }
  }

  function makeAutoComplete(
    field: string,
    filters: Record<string, AutoCompleteItem>,
    i: number,
    provider: PriceProvider
  ) {
    return async function autocomplete(filterText: string): Promise<AutoCompleteItem[]> {
      for (let j = 0; j < i; j++) {
        if (_.isEmpty(filters[provider.fields[j].id])) {
          return [];
        }
      }

      const queryFilters = _.mapValues(filters, (v) => v?.id);
      queryFilters[field] = filterText;
      const { completions } = await ajax("/api/price/autocomplete", {
        method: "POST",
        body: JSON.stringify({
          field,
          provider: selectedProvider.code,
          filters: queryFilters
        })
      });
      return completions;
    };
  }

  const dispatch = createEventDispatcher();
</script>

<Modal bind:active={open} footerClass="is-justify-content-space-between">
  <svelte:fragment slot="head" let:close>
    <p class="modal-card-title">{label}</p>
    <button class="delete" aria-label="close" on:click={(e) => close(e)} />
  </svelte:fragment>
  <div style="min-height: 500px;" slot="body">
    {#if selectedProvider}
      <div class="field">
        <label class="label" for="price-provider">Provider</label>
        <div class="control" id="price-provider">
          <div class="select">
            <select bind:value={selectedProvider} required on:change={(_e) => reset()}>
              {#each providers as provider}
                <option value={provider}>{provider.label}</option>
              {/each}
            </select>
          </div>
        </div>
      </div>
      <div class="field">
        {#each selectedProvider.fields as field, i}
          <div class="field">
            <label class="label" for="">{field.label}</label>
            <div class="control">
              {#if field.inputType == "text"}
                <input class="input" type="text" bind:value={code} required />
              {:else}
                {#key autocompleteCache[i]}
                  <Select
                    bind:value={filters[field.id]}
                    --list-z-index="5"
                    showChevron={true}
                    loadOptions={makeAutoComplete(field.id, filters, i, selectedProvider)}
                    label="label"
                    itemId="id"
                    searchable={true}
                    clearable={false}
                    on:change={() => {
                      _.each(selectedProvider.fields, (f, j) => {
                        if (j > i) {
                          clearCache(j);
                          filters[f.id] = null;
                        }
                      });

                      if (i === selectedProvider.fields.length - 1) {
                        code = filters[field.id].id;
                      } else {
                        code = "";
                      }
                    }}
                  ></Select>
                {/key}
              {/if}
              <p class="help">{field.help}</p>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
  <svelte:fragment slot="foot" let:close>
    <div>
      <button
        class="button is-success"
        disabled={_.isEmpty(code)}
        on:click={(e) => {
          dispatch("select", { code: code, provider: selectedProvider.code });
          reset();
          close(e);
        }}>Select</button
      >
      <button class="button" on:click={(e) => close(e)}>Cancel</button>
    </div>

    <div>
      <button
        on:click={(_e) => clearProviderCache()}
        class="button is-danger {isLoading && 'is-loading'}"
        disabled={!selectedProvider}>Clear Provider Cache</button
      >
    </div>
  </svelte:fragment>
</Modal>
