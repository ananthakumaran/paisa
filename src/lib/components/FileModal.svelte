<script lang="ts">
  import Modal from "$lib/components/Modal.svelte";
  import _ from "lodash";
  import { createEventDispatcher } from "svelte";

  export let label = "Save As";
  export let help = "Create or overwrite existing file";
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
    <label class="label" for="save-filename">File Name</label>
    <div class="control" id="save-filename">
      <input class="input" type="text" placeholder="expense.ledger" bind:value={destinationFile} />
      <p class="help">{help}</p>
    </div>
  </div>
  <svelte:fragment slot="foot" let:close>
    <button
      class="button is-success"
      disabled={_.isEmpty(destinationFile)}
      on:click={(e) => dispatch("save", destinationFile) && close(e)}>{label}</button
    >
    <button class="button" on:click={(e) => close(e)}>Cancel</button>
  </svelte:fragment>
</Modal>
