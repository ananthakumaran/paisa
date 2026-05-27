<script lang="ts">
  import Modal from "$lib/components/Modal.svelte";
  import _ from "lodash";
  import { createEventDispatcher } from "svelte";
  import { _ as t } from "$lib/i18n";
  import { get } from "svelte/store";

  export let label: string = get(t)("component.file_modal.save_as");
  export let help: string = get(t)("component.file_modal.create_or_overwrite");
  export let placeholder = "expense.ledger";
  export let open = false;
  let destinationFile = "";

  const dispatch = createEventDispatcher();
</script>

<Modal bind:active={open}>
  <svelte:fragment slot="head" let:close>
    <p class="modal-card-title">{label}</p>
    <button class="delete" aria-label="close" on:click={(e) => close(e)} />
  </svelte:fragment>
  <div class="field" slot="body">
    <label class="label" for="save-filename">{$t("component.file_modal.file_name")}</label>
    <div class="control" id="save-filename">
      <input class="input" type="text" {placeholder} bind:value={destinationFile} />
      <p class="help">{help}</p>
    </div>
  </div>
  <svelte:fragment slot="foot" let:close>
    <button
      class="button is-success"
      disabled={_.isEmpty(destinationFile)}
      on:click={(e) => dispatch("save", destinationFile) && close(e)}>{label}</button
    >
    <button class="button" on:click={(e) => close(e)}>{$t("component.file_modal.cancel")}</button>
  </svelte:fragment>
</Modal>
