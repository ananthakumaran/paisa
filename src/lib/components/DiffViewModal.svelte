<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import Modal from "./Modal.svelte";
  import type { MergeView } from "@codemirror/merge";
  import { createDiffEditor } from "$lib/editor";
  import type { LedgerFile } from "$lib/utils";
  let editorDom: Element;
  let editor: MergeView;
  export let oldFiles: LedgerFile[] = [];
  export let newFiles: LedgerFile[] = [];
  export let updatedTransactionsCount = 0;
  let changedOldFiles: LedgerFile[] = [];
  let changedNewFiles: LedgerFile[] = [];
  let selectedFileIndex = 0;

  const dispatch = createEventDispatcher();
  export let open = false;

  $: {
    changedOldFiles = [];
    changedNewFiles = [];
    for (let i = 0; i < newFiles.length; i++) {
      if (oldFiles[i].content !== newFiles[i].content) {
        changedOldFiles.push(oldFiles[i]);
        changedNewFiles.push(newFiles[i]);
      }
    }
  }

  $: if (open) {
    if (editor) {
      editor.destroy();
    }

    if (changedOldFiles.length > 0) {
      editor = createDiffEditor(
        changedOldFiles[selectedFileIndex].content,
        changedNewFiles[selectedFileIndex].content,
        editorDom
      );
    }
  }
</script>

<Modal
  bind:active={open}
  width="min(1300px, 100vw)"
  bodyClass="p-0 min-h-[500px]"
  headerClass="pt-1 pb-1"
  footerClass="is-justify-content-right"
>
  <svelte:fragment slot="head" let:close>
    <p class="modal-card-title">
      {#if changedOldFiles.length > 0}
        {changedOldFiles[selectedFileIndex]?.name}
        [{selectedFileIndex + 1}/{changedNewFiles.length}]
      {:else}
        No Changes
      {/if}
    </p>
    <div class="field has-addons mt-3 mr-3">
      {#if changedOldFiles.length > 0}
        <div class="mr-3 mt-2"><b>{updatedTransactionsCount}</b> transaction(s) changed</div>
      {/if}
      <p class="control">
        <button
          class="button"
          disabled={selectedFileIndex <= 0}
          on:click={(_e) => selectedFileIndex--}
        >
          <span class="icon is-small">
            <i class="fas fa-chevron-left"></i>
          </span>
          <span>Prev</span>
        </button>
      </p>
      <p class="control">
        <button
          class="button"
          disabled={selectedFileIndex >= changedNewFiles.length - 1}
          on:click={(_e) => selectedFileIndex++}
        >
          <span>Next</span>
          <span class="icon is-small">
            <i class="fas fa-chevron-right"></i>
          </span>
        </button>
      </p>
    </div>
    <button class="delete" aria-label="close" on:click={(e) => close(e)} />
  </svelte:fragment>
  <div class="field" slot="body">
    <div class="box py-0">
      <div class="diff-editor" bind:this={editorDom} />
      {#if changedOldFiles.length === 0}
        <div class="has-text-centered mt-6">
          <strong>Oops!</strong> No changes has been made. Make sure the bulk edit arguments are correct.
        </div>
      {/if}
    </div>
  </div>
  <svelte:fragment slot="foot" let:close>
    <button class="button" on:click={(e) => close(e)}>Cancel</button>
    {#if changedOldFiles.length > 0}
      <button
        class="button is-success"
        on:click={(e) => dispatch("save", changedNewFiles) && close(e)}>Save All</button
      >
    {/if}
  </svelte:fragment>
</Modal>
