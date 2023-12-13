<script lang="ts">
  import sha256 from "crypto-js/sha256";
  import type { JSONSchema7 } from "json-schema";
  import Select from "svelte-select";
  import _ from "lodash";
  import PriceCodeSearchModal from "./PriceCodeSearchModal.svelte";
  import { iconGlyph, iconsList } from "$lib/icon";
  import AccountSelect from "./AccountsSelect.svelte";

  interface Schema extends JSONSchema7 {
    "ui:header"?: string;
    "ui:widget"?: string;
    "ui:order"?: number;
  }

  const ICON_MAX_RESULTS = 200;

  export let key: string;
  export let value: any;
  export let rawValue: string = "";
  export let schema: Schema;
  export let depth: number = 0;
  export let required = false;
  export let deletable: () => void = null;
  export let disabled: boolean = false;
  export let allAccounts: string[];

  export let modalOpen = false;

  let open = depth < 1;
  $: title = _.startCase(key);

  function newItem(schema: any) {
    return _.cloneDeep(schema.default[0]);
  }

  function sortedProperties(schema: Schema) {
    return _.sortBy(Object.entries(schema.properties), ([key, subSchema]: [string, Schema]) => {
      return [
        subSchema["ui:order"] || 999,
        _.includes(schema.required || [], key) ? 0 : 1,
        subSchema.type == "object" ? 2 : subSchema.type == "array" ? 3 : 1,
        key
      ];
    });
  }

  function documentation(schema: Schema) {
    if (schema.description) {
      return `<p style="max-width: 300px">${schema.description}</p>`;
    }
    return null;
  }

  async function searchIcons(text: string) {
    text = text.toLowerCase();
    if (_.isEmpty(text)) {
      return _.take(iconsList, ICON_MAX_RESULTS);
    }
    return _.take(
      iconsList.filter((icon) => icon.includes(text)),
      ICON_MAX_RESULTS
    );
  }
</script>

{#if deletable}
  <a on:click={(_e) => deletable()} class="config-delete">
    <span class="icon is-small">
      <i class="fas fa-circle-minus"></i>
    </span>
  </a>
{/if}

{#if schema["ui:widget"] == "hidden"}
  <div></div>
{:else if schema["ui:widget"] == "password"}
  <div class="field is-horizontal">
    <div class="field-label is-small">
      <label data-tippy-content={documentation(schema)} for="" class="label">{title}</label>
    </div>
    <div class="field-body">
      <div class="field">
        <div class="control">
          <input
            {disabled}
            {required}
            class="input is-small"
            style="max-width: 350px;"
            type="password"
            bind:value={rawValue}
            on:change={() => {
              if (!_.isEmpty(rawValue)) {
                value = "sha256:" + sha256(sha256(rawValue).toString()).toString();
              }
            }}
          />
        </div>
      </div>
    </div>
  </div>
{:else if schema["ui:widget"] == "icon"}
  <div class="field is-horizontal">
    <div class="field-label is-small">
      <label for="" data-tippy-content={documentation(schema)} class="label">{title}</label>
    </div>
    <div class="field-body">
      <div class="field">
        <div class="control" style="max-width: 350px">
          <Select
            bind:justValue={value}
            class="icon-select is-small"
            {value}
            showChevron={true}
            loadOptions={searchIcons}
            searchable={true}
            clearable={!required}
          >
            <div class="custom-icon" slot="selection" let:selection>
              <span>{iconGlyph(selection.value)} {selection.value}</span>
            </div>
            <div class="custom-icon" slot="item" let:item>
              <span class="name">{iconGlyph(item.value)} {item.value}</span>
            </div>
          </Select>
        </div>
      </div>
    </div>
  </div>
{:else if schema["ui:widget"] == "boolean"}
  <div class="field is-horizontal">
    <div class="field-label is-small">
      <label for="" data-tippy-content={documentation(schema)} class="label">{title}</label>
    </div>
    <div class="field-body">
      <div class="field">
        <div class="control">
          <label class="radio">
            <input value="yes" bind:group={value} type="radio" name="yes" />
            Yes
          </label>
          <label class="radio">
            <input value="no" bind:group={value} type="radio" name="no" />
            No
          </label>
        </div>
      </div>
    </div>
  </div>
{:else if schema.type === "string" || _.isEqual(schema.type, ["string", "integer"])}
  <div class="field is-horizontal">
    <div class="field-label is-small">
      <label data-tippy-content={documentation(schema)} for="" class="label">{title}</label>
    </div>
    <div class="field-body">
      <div class="field">
        <div class="control">
          {#if schema.enum}
            <div class="select is-small">
              <select {disabled} bind:value {required}>
                {#each schema.enum as option}
                  <option value={option}>{option}</option>
                {/each}
              </select>
            </div>
          {:else if schema["ui:widget"] == "textarea"}
            <textarea
              {disabled}
              {required}
              class="textarea is-small"
              style="min-width: 350px;max-width: 350px; width: 350px;"
              rows="5"
              bind:value
              spellcheck="false"
              data-enable-grammarly="false"
            />
          {:else}
            <input
              {disabled}
              {required}
              pattern={schema.pattern}
              class="input is-small"
              style="max-width: 350px;"
              type="text"
              bind:value
            />
          {/if}
        </div>
      </div>
    </div>
  </div>
{:else if schema.type === "integer" || schema.type === "number"}
  <div class="field is-horizontal">
    <div class="field-label is-small">
      <label for="" data-tippy-content={documentation(schema)} class="label">{title}</label>
    </div>
    <div class="field-body">
      <div class="field">
        <div class="control">
          <input
            {required}
            class="input is-small"
            style="max-width: 350px;"
            type="number"
            min={schema.minimum}
            max={schema.maximum}
            step={schema.type == "integer" ? 1 : 0.01}
            bind:value
          />
        </div>
      </div>
    </div>
  </div>
{:else if schema["ui:widget"] == "accounts"}
  <div class="field is-horizontal">
    <div class="field-label is-small">
      <label for="" data-tippy-content={documentation(schema)} class="label">{title}</label>
    </div>
    <div class="field-body">
      <div class="field">
        <div class="control pr-5">
          <AccountSelect {allAccounts} bind:accounts={value} />
        </div>
      </div>
    </div>
  </div>
{:else if schema["ui:widget"] == "price"}
  <div class="config-header">
    <a class="is-link" data-tippy-content={documentation(schema)}>
      <span>{title}</span>
    </a>

    <a on:click={(_e) => (modalOpen = true)} class="is-link">
      <span class="icon is-small">
        <i class="fas fa-pen-to-square"></i>
      </span>
    </a>
  </div>

  <PriceCodeSearchModal
    bind:open={modalOpen}
    on:select={(e) => {
      value["code"] = e.detail.code;
      value["provider"] = e.detail.provider;
    }}
  />

  <div class="config-body {depth % 2 == 1 ? 'odd' : 'even'}">
    {#each sortedProperties(schema) as [key, subSchema]}
      <svelte:self
        {allAccounts}
        required={_.includes(schema.required || [], key)}
        depth={depth + 1}
        {key}
        bind:value={value[key]}
        schema={subSchema}
        disabled={true}
      />
    {/each}
  </div>
{:else if schema.type == "object"}
  <div class="config-header">
    <a
      class="is-link is-light invertable"
      data-tippy-content={documentation(schema)}
      on:click={(_e) => (open = !open)}
    >
      <span>{schema["ui:header"] ? value[schema["ui:header"]] || title : title}</span>
      <span class="icon is-small">
        <i class="fas {open ? 'fa-angle-up' : 'fa-angle-down'}"></i>
      </span>
    </a>
  </div>

  {#if open}
    <div class="config-body {depth % 2 == 1 ? 'odd' : 'even'}">
      {#each sortedProperties(schema) as [key, subSchema]}
        <svelte:self
          {allAccounts}
          required={_.includes(schema.required || [], key)}
          depth={depth + 1}
          {key}
          bind:value={value[key]}
          schema={subSchema}
        />
      {/each}
    </div>
  {/if}
{:else if schema.type == "array"}
  <div class="config-header">
    <a
      class="is-link is-light invertable"
      data-tippy-content={documentation(schema)}
      on:click={(_e) => (open = !open)}
    >
      <span>{title}</span>
      <span class="icon is-small">
        <i class="fas {open ? 'fa-angle-up' : 'fa-angle-down'}"></i>
      </span>
    </a>
    {#if open}
      <a on:click={(_e) => (value = [newItem(schema), ...value])} class="config-add">
        <span class="icon is-small">
          <i class="fas fa-circle-plus"></i>
        </span>
      </a>
    {/if}
  </div>

  {#if open}
    <div class="config-body {depth % 2 == 1 ? 'odd' : 'even'}">
      {#each value as _item, i}
        <svelte:self
          {allAccounts}
          deletable={() => {
            value.splice(i, 1);
            value = [...value];
          }}
          depth={depth + 1}
          key=""
          bind:value={value[i]}
          schema={schema.items}
        />
      {/each}
    </div>
  {/if}
{:else}
  <div>{JSON.stringify(schema)}</div>
{/if}
