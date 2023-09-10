<script lang="ts">
  import type { JSONSchema7 } from "json-schema";
  import _ from "lodash";

  export let key: string;
  export let value: any;
  export let schema: JSONSchema7;
  export let depth: number = 0;
  export let required = false;
  export let deletable: () => void = null;

  let open = depth < 1;
  $: title = _.startCase(key);

  function newItem(schema: any) {
    return _.cloneDeep(schema.default[0]);
  }

  function sortedProperties(schema: JSONSchema7) {
    return _.sortBy(
      Object.entries(schema.properties),
      ([key, subSchema]: [string, JSONSchema7]) => {
        return [
          _.includes(schema.required || [], key) ? 0 : 1,
          subSchema.type == "object" ? 2 : subSchema.type == "array" ? 3 : 1,
          key
        ];
      }
    );
  }

  function documentation(schema: JSONSchema7) {
    if (schema.description) {
      return `<p style="max-width: 300px">${schema.description}</p>`;
    }
    return null;
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
      <label data-tippy-content={documentation(schema)} for="" class="label">{title}</label>
    </div>
    <div class="field-body">
      <div class="field">
        <div class="control">
          {#if schema.enum}
            <div class="select is-small">
              <select bind:value {required}>
                {#each schema.enum as option}
                  <option value={option}>{option}</option>
                {/each}
              </select>
            </div>
          {:else}
            <input
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
{:else if schema.type == "object"}
  <div class="config-header">
    <a
      class="is-link is-light invertable"
      data-tippy-content={documentation(schema)}
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
      {#each sortedProperties(schema) as [key, subSchema]}
        <svelte:self
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
