Svelte komponenti şu şekildedir: <script lang="ts">
  import Modal from "$lib/components/Modal.svelte";
  import _ from "lodash";
  import { createEventDispatcher } from "svelte";

  export let label = "Farklı Kaydet";
  export let help = "Yeni dosya oluştur veya mevcut dosyanın üzerine yaz";
  export let placeholder = "gider.ledger";
  export let open = false;
  let destinationFile = "";

  const dispatch = createEventDispatcher();
</script>

<Modal bind:active={open}>
  <svelte:fragment slot="head" let:close>
    <p class="modal-card-title">{label}</p>
    <button class="delete" aria-label="kapat" on:click={(e) => close(e)} />
  </svelte:fragment>
  <div class="field" slot="body">
    <label class="label" for="save-filename">Dosya Adı</label>
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
    <button class="button" on:click={(e) => close(e)}>İptal</button>
  </svelte:fragment>
</Modal>