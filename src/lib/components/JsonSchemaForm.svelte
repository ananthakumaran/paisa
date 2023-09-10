<script lang="ts">
  import type { JSONSchema7 } from "json-schema";
  import _ from "lodash";

  export let key: string;
  export let value: any;
  export let schema: JSONSchema7;
  export let depth: number = 0;
  export let deletable: () => void = null;

  let open = depth < 1;
  $: title = _.startCase(key);

  function newItem(schema: any) {
    return _.cloneDeep(schema.default[0]);
  }
</script>

{#if deletable}
  <a on:click={(_e) => deletable()} class="config-delete">
    <span class="icon is-small">
      <i class="fas fa-circle-minus"></i>
    </span>
  </a>
{/if}

{#if schema.type === "string" || _.isEqual(schema.type, ["string", "integer"])}
  <div class="field is-horizontal">
    <div class="field-label is-small">
      <label for="" class="label">{title}</label>
    </div>
    <div class="field-body">
      <div class="field">
        <div class="control">
          <input
            class="input is-small"
            style="max-width: 300px;"
            type="text"
            placeholder="Text input"
            bind:value
          />
        </div>
      </div>
    </div>
  </div>
{:else if schema.type === "integer" || schema.type === "number"}
  <div class="field is-horizontal">
    <div class="field-label is-small">
      <label for="" class="label">{title}</label>
    </div>
    <div class="field-body">
      <div class="field">
        <div class="control">
          <input
            class="input is-small"
            style="max-width: 300px;"
            type="number"
            placeholder="Text input"
            bind:value
          />
        </div>
      </div>
    </div>
  </div>
{:else if schema.type == "object"}
  <div class="config-header">
    <a
      class="is-link is-light invertable"
      data-tippy-content="<div>This is a even more <b>important config</a></div>"
      on:click={(_e) => (open = !open)}
    >
      <span>{value.name || value.code || title}</span>
      <span class="icon is-small">
        <i class="fas {open ? 'fa-angle-up' : 'fa-angle-down'}"></i>
      </span>
    </a>
  </div>

  {#if open}
    <div class="config-body {depth % 2 == 1 ? 'odd' : 'even'}">
      {#each Object.entries(schema.properties) as [key, subSchema]}
        <svelte:self depth={depth + 1} {key} bind:value={value[key]} schema={subSchema} />
      {/each}
    </div>
  {/if}
{:else if schema.type == "array"}
  <div class="config-header">
    <a
      class="is-link is-light invertable"
      data-tippy-content="This is a very <b>important config</a>"
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
          deletable={() => {
            console.log(i, value);
            value.splice(i, 1);
            value = [...value];
            console.log(value);
          }}
          depth={depth + 1}
          {key}
          bind:value={value[i]}
          schema={schema.items}
        />
      {/each}
    </div>
  {/if}
{:else}
  <div>{JSON.stringify(schema)}</div>
{/if}
