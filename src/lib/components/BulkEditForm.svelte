<script lang="ts">
  import _ from "lodash";
  import { createEventDispatcher } from "svelte";
  import Select from "svelte-select";
  import { _ as t } from "$lib/i18n";
  import { get } from "svelte/store";

  export let accounts: string[];

  $: selectItems = _.map(accounts, (account) => {
    return { id: account, name: account };
  });

  let selectedItem: { id: string; name: string };

  const OPERATIONS = [
    { id: "rename_account", label: get(t)("component.bulk_edit.rename_account") }
  ];
  let selectedOperation = OPERATIONS[0].id;

  let args = { oldAccountName: "", newAccountName: "" };

  const dispatch = createEventDispatcher();
</script>

<div class="field is-grouped">
  <div class="control">
    <div class="select">
      <select bind:value={selectedOperation}>
        {#each OPERATIONS as operation}
          <option value={operation.id}>{operation.label}</option>
        {/each}
      </select>
    </div>
  </div>
  {#if selectedOperation === "rename_account"}
    <div class="control is-expanded">
      <Select
        bind:value={selectedItem}
        showChevron={true}
        items={selectItems}
        label="name"
        itemId="id"
        placeholder={$t("component.bulk_edit.old_account")}
        searchable={true}
        clearable={false}
        on:change={(_e) => {
          args.oldAccountName = selectedItem.name;
        }}
      ></Select>
    </div>
    <div class="control is-expanded">
      <input
        bind:value={args.newAccountName}
        class="input"
        type="text"
        placeholder={$t("component.bulk_edit.new_account")}
      />
    </div>
  {/if}
  <p class="control">
    <a
      class="button is-link"
      on:click={(_e) => dispatch("preview", { operation: selectedOperation, args: args })}
      >{$t("component.bulk_edit.preview")}</a
    >
  </p>
</div>
