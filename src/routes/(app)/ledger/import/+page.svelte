<script lang="ts">
  import Select from "svelte-select";
  import {
    createEditor as createTemplateEditor,
    editorState as templateEditorState
  } from "$lib/template_editor";
  import {
    createEditor as createPreviewEditor,
    updateContent as updatePreviewContent
  } from "$lib/editor";
  import Dropzone from "svelte-file-dropzone/Dropzone.svelte";
  import { parse, asRows, render as renderJournal } from "$lib/spreadsheet";
  import _ from "lodash";
  import type { EditorView } from "codemirror";
  import { onMount } from "svelte";
  import { ajax, type ImportTemplate } from "$lib/utils";
  import { accountTfIdf } from "../../../../store";
  import * as toast from "bulma-toast";
  import FileModal from "$lib/components/FileModal.svelte";
  import Modal from "$lib/components/Modal.svelte";

  let templates: ImportTemplate[] = [];
  let selectedTemplate: ImportTemplate;
  let saveAsName: string;
  let lastTemplate: any;
  let lastData: any;
  let preview = "";
  let parseErrorMessage: string = null;
  let columnCount: number;
  let data: any[][] = [];
  let rows: Array<Record<string, any>> = [];
  let lastOptions: any;
  let options: { reverse: boolean } = { reverse: false };

  let templateEditorDom: Element;
  let templateEditor: EditorView;

  let previewEditorDom: Element;
  let previewEditor: EditorView;

  onMount(async () => {
    accountTfIdf.set(await ajax("/api/account/tf_idf"));
    ({ templates } = await ajax("/api/templates"));
    selectedTemplate = templates[0];
    saveAsName = selectedTemplate.name;
    templateEditor = createTemplateEditor(selectedTemplate.content, templateEditorDom);
    previewEditor = createPreviewEditor(preview, previewEditorDom, { readonly: true });
  });

  $: saveAsNameDuplicate = !!_.find(templates, { name: saveAsName, template_type: "custom" });

  async function save() {
    const { id } = await ajax("/api/templates/upsert", {
      method: "POST",
      body: JSON.stringify({
        name: saveAsName,
        content: templateEditor.state.doc.toString()
      }),
      background: true
    });
    ({ templates } = await ajax("/api/templates", { background: true }));
    selectedTemplate = _.find(templates, { id });
    saveAsName = selectedTemplate.name;
    toast.toast({
      message: `Saved ${saveAsName}`,
      type: "is-success"
    });

    $templateEditorState = _.assign({}, $templateEditorState, { hasUnsavedChanges: false });
  }

  async function remove() {
    const oldName = selectedTemplate.name;
    const confirmed = confirm(`Are you sure you want to delete ${oldName} template?`);
    if (!confirmed) {
      return;
    }
    await ajax("/api/templates/delete", {
      method: "POST",
      body: JSON.stringify({
        name: selectedTemplate.name
      }),
      background: true
    });
    ({ templates } = await ajax("/api/templates", { background: true }));
    selectedTemplate = templates[0];
    saveAsName = selectedTemplate.name;
    toast.toast({
      message: `Removed ${oldName}`,
      type: "is-success"
    });

    $templateEditorState = _.assign({}, $templateEditorState, { hasUnsavedChanges: false });
  }

  let input: any;

  $: if (!_.isEmpty(data) && $templateEditorState.template) {
    if (
      lastTemplate != $templateEditorState.template ||
      lastData != data ||
      lastOptions != options
    ) {
      try {
        preview = renderJournal(rows, $templateEditorState.template, { reverse: options.reverse });
        updatePreviewContent(previewEditor, preview);
        lastTemplate = $templateEditorState.template;
        lastData = data;
        lastOptions = _.clone(options);
      } catch (e) {
        console.log(e);
      }
    }
  }

  $: if (selectedTemplate && templateEditor) {
    if (templateEditor.state.doc.toString() != selectedTemplate.content) {
      templateEditor.destroy();
      templateEditor = createTemplateEditor(selectedTemplate.content, templateEditorDom);
    }
  }

  async function handleFilesSelect(e: { detail: { acceptedFiles: File[] } }) {
    const { acceptedFiles } = e.detail;

    const results = await parse(acceptedFiles[0]);
    if (results.error) {
      parseErrorMessage = results.error;
    } else {
      parseErrorMessage = null;
      data = results.data;
      rows = asRows(results);

      columnCount = _.maxBy(data, (row) => row.length).length;
      _.each(data, (row) => {
        row.length = columnCount;
      });
    }
  }

  async function copyToClipboard() {
    try {
      await navigator.clipboard.writeText(preview);
      toast.toast({
        message: "Copied to clipboard",
        type: "is-success"
      });
    } catch (e) {
      console.log(e);
      toast.toast({
        message: "Failed to copy to clipboard",
        type: "is-danger"
      });
    }
  }

  let modalOpen = false;
  function openSaveModal() {
    modalOpen = true;
  }

  async function saveToFile(destinationFile: string) {
    const { saved, message } = await ajax("/api/editor/save", {
      method: "POST",
      body: JSON.stringify({ name: destinationFile, content: preview, operation: "overwrite" }),
      background: true
    });

    if (saved) {
      toast.toast({
        message: `Saved <b><a href="/ledger/editor/${encodeURIComponent(
          destinationFile
        )}">${destinationFile}</a></b>`,
        type: "is-success",
        duration: 5000
      });
    } else {
      toast.toast({
        message: `Failed to save ${destinationFile}. reason: ${message}`,
        type: "is-danger",
        duration: 10000
      });
    }
  }

  function builtinNotAllowed(action: string, template: ImportTemplate) {
    if (template?.template_type == "builtin") {
      return `Not allowed to ${action.toLowerCase()} builtin template`;
    }
    return action;
  }

  let templateCreateModalOpen = false;
  function openTemplateCreateModal() {
    templateCreateModalOpen = true;
  }
</script>

<Modal bind:active={templateCreateModalOpen}>
  <svelte:fragment slot="head" let:close>
    <p class="modal-card-title">Create Template</p>
    <button class="delete" aria-label="close" on:click={(e) => close(e)} />
  </svelte:fragment>
  <div class="field" slot="body">
    <label class="label" for="save-filename">Template Name</label>
    <div class="control" id="save-filename">
      <input class="input" type="text" bind:value={saveAsName} />
      {#if saveAsNameDuplicate}
        <p class="help is-danger">Template with the same name already exists</p>
      {/if}
    </div>
  </div>
  <svelte:fragment slot="foot" let:close>
    <button
      class="button is-success"
      disabled={_.isEmpty(saveAsName) || saveAsNameDuplicate}
      on:click={(e) => save() && close(e)}>Create</button
    >
    <button class="button" on:click={(e) => close(e)}>Cancel</button>
  </svelte:fragment>
</Modal>

<FileModal bind:open={modalOpen} on:save={(e) => saveToFile(e.detail)} />

<section class="section tab-import" style="padding-bottom: 0 !important">
  <div class="container is-fluid">
    <div class="columns mb-0">
      <div class="column is-5 py-0">
        <div class="box p-3 mb-3 overflow-x-scroll">
          <div class="field is-grouped mb-0">
            <p class="control">
              <span data-tippy-content="Create" data-tippy-followCursor="false">
                <button class="button" on:click={(_e) => openTemplateCreateModal()}>
                  <span class="icon is-small">
                    <i class="fas fa-file-circle-plus" />
                  </span>
                </button>
              </span>

              <span
                class="du-tooltip ml-4"
                data-tippy-followCursor="false"
                data-tippy-content={$templateEditorState.hasUnsavedChanges == false
                  ? "No Unsaved Chagnes"
                  : builtinNotAllowed("Save", selectedTemplate)}
              >
                <button
                  class="button"
                  on:click={(_e) => save()}
                  disabled={$templateEditorState.hasUnsavedChanges == false ||
                    selectedTemplate?.template_type == "builtin"}
                >
                  <span class="icon is-small">
                    <i class="fas fa-floppy-disk" />
                  </span>
                </button>
              </span>

              <span
                data-tippy-followCursor="false"
                data-tippy-content={builtinNotAllowed("Delete", selectedTemplate)}
              >
                <button
                  class="button"
                  on:click={(_e) => remove()}
                  disabled={selectedTemplate?.template_type == "builtin"}
                >
                  <span class="icon is-small">
                    <i class="fas fa-trash-can" />
                  </span>
                </button>
              </span>
            </p>

            <p class="control is-expanded">
              <Select
                bind:value={selectedTemplate}
                showChevron={true}
                items={templates}
                label="name"
                itemId="id"
                searchable={true}
                clearable={false}
                floatingConfig={{ strategy: "fixed" }}
                on:change={(_e) => {
                  saveAsName = selectedTemplate.name;
                }}
              >
                <div slot="selection" let:selection>
                  {selection.name}
                  <span class="tag is-small is-link invertable is-light"
                    >{selection.template_type}</span
                  >
                </div>
                <div slot="item" let:item>
                  <span class="name">{item.name}</span>
                  <span class="tag is-small is-link invertable is-light">{item.template_type}</span>
                </div>
              </Select>
            </p>
          </div>
        </div>
        <div class="box py-0">
          <div class="field">
            <div class="control">
              <div class="template-editor" bind:this={templateEditorDom} />
            </div>
          </div>
        </div>
        <div class="box py-0">
          <div class="field">
            <div class="control">
              <button
                title="copy to clipboard"
                class="button is-small clipboard"
                disabled={_.isEmpty(preview)}
                on:click={copyToClipboard}
              >
                <span class="icon is-small">
                  <i class="fas fa-clipboard" />
                </span>
              </button>
              <button
                title="save"
                class="button is-small save"
                disabled={_.isEmpty(preview)}
                on:click={openSaveModal}
              >
                <span class="icon is-small">
                  <i class="fas fa-floppy-disk" />
                </span>
              </button>
              <div class="preview-editor" bind:this={previewEditorDom} />
            </div>
          </div>
        </div>
      </div>
      <div class="column is-7 py-0">
        <div class="box p-3 mb-3">
          <Dropzone
            multiple={false}
            inputElement={input}
            accept=".csv,.txt,.xls,.xlsx,.pdf,.CSV,.TXT,.XLS,.XLSX,.PDF"
            on:drop={handleFilesSelect}
          >
            Drag 'n' drop CSV, TXT, XLS, XLSX, PDF file here or click to select
          </Dropzone>
        </div>
        <div class="is-flex justify-end mb-3">
          <div class="field color-switch">
            <input
              id="import-reverse"
              type="checkbox"
              bind:checked={options.reverse}
              class="switch is-rounded is-small"
            />
            <label for="import-reverse">Reverse</label>
          </div>
        </div>
        {#if parseErrorMessage}
          <div class="message invertable is-danger">
            <div class="message-header">Failed to parse document</div>
            <div class="message-body">{parseErrorMessage}</div>
          </div>
        {/if}
        {#if !_.isEmpty(data)}
          <div class="table-wrapper">
            <table
              class="mt-0 table is-bordered is-size-7 is-narrow has-sticky-header has-sticky-column"
            >
              <thead>
                <tr>
                  <th />
                  {#each _.range(0, columnCount) as ci}
                    <th class="has-background-light">{String.fromCharCode(65 + ci)}</th>
                  {/each}
                </tr>
              </thead>
              <tbody>
                {#each data as row, ri}
                  <tr>
                    <th class="has-background-light"><b>{ri}</b></th>
                    {#each row as cell}
                      <td>{cell || ""}</td>
                    {/each}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>
    </div>
    <div />
  </div>
</section>

<style lang="scss">
  @import "bulma/sass/utilities/_all.sass";

  $import-full-height: calc(100vh - 205px);

  .clipboard {
    float: right;
    position: absolute;
    right: 0;
    z-index: 10;
  }

  .save {
    float: right;
    position: absolute !important;
    right: 40px;
    z-index: 10;
  }

  .table-wrapper {
    overflow-x: auto;
    overflow-y: auto;
    max-height: $import-full-height;
  }

  .color-switch {
    .switch[type="checkbox"]:checked + label::before,
    .switch[type="checkbox"]:checked + label:before {
      background: $link;
    }
  }
</style>
