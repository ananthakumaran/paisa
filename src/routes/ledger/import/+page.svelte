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
  import { parse, asRows, render as renderJournal } from "$lib/sheet";
  import _ from "lodash";
  import type { EditorView } from "codemirror";
  import { onMount } from "svelte";
  import { ajax, type Template } from "$lib/utils";
  import { accountTfIdf } from "../../../store";
  import * as toast from "bulma-toast";

  let templates: Template[] = [];
  let selectedTemplate: Template;
  let saveAsName: string;
  let lastTemplate: any;
  let lastData: any;
  let preview = "";
  let columnCount: number;
  let data: any[][] = [];
  let rows: Array<Record<string, any>> = [];

  let templateEditorDom: Element;
  let templateEditor: EditorView;

  let previewEditorDom: Element;
  let previewEditor: EditorView;

  onMount(async () => {
    ({ templates } = await ajax("/api/templates"));
    selectedTemplate = templates[0];
    saveAsName = selectedTemplate.name;
    accountTfIdf.set(await ajax("/api/account/tf_idf"));
    templateEditor = createTemplateEditor(selectedTemplate.content, templateEditorDom);
    previewEditor = createPreviewEditor(preview, previewEditorDom, { readonly: true });
  });

  async function save() {
    const { id } = await ajax("/api/templates/upsert", {
      method: "POST",
      body: JSON.stringify({
        name: saveAsName,
        content: templateEditor.state.doc.toString()
      })
    });
    ({ templates } = await ajax("/api/templates"));
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
    await ajax("/api/templates/delete", {
      method: "POST",
      body: JSON.stringify({
        id: selectedTemplate.id
      })
    });
    ({ templates } = await ajax("/api/templates"));
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
    if (lastTemplate != $templateEditorState.template || lastData != data) {
      try {
        preview = renderJournal(rows, $templateEditorState.template);
        updatePreviewContent(previewEditor, preview);
        lastTemplate = $templateEditorState.template;
        lastData = data;
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
    data = results.data;
    rows = asRows(results);

    columnCount = _.maxBy(data, (row) => row.length).length;
    _.each(data, (row) => {
      row.length = columnCount;
    });
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
</script>

<section class="section tab-import">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-5">
        <div class="box px-3">
          <div class="field mb-2">
            <p class="control">
              <Select
                bind:value={selectedTemplate}
                showChevron={true}
                items={templates}
                label="name"
                itemId="id"
                searchable={true}
                clearable={false}
                --chevron-color="hsl(229deg, 53%, 53%)"
                --font-size="14px"
                --item-hover-bg="none"
                --item-hover-color="hsl(0deg, 0%, 4%)"
                --item-is-active-bg="none"
                --item-is-active-color="hsl(229deg, 53%, 53%)"
                on:change={(_e) => {
                  saveAsName = selectedTemplate.name;
                }}
              >
                <div slot="selection" let:selection>
                  {selection.name}
                  <span class="tag is-small is-link is-light">{selection.template_type}</span>
                </div>
                <div slot="item" let:item>
                  <span class="name">{item.name}</span>
                  <span class="tag is-small is-link is-light">{item.template_type}</span>
                </div>
              </Select>
            </p>
          </div>

          <div class="is-flex is-align-items-center">
            <div class="field has-addons mb-0">
              <p class="control">
                <button
                  class="button is-small"
                  on:click={(_e) => save()}
                  disabled={$templateEditorState.hasUnsavedChanges == false}
                >
                  <span class="icon is-small">
                    <i class="fas fa-floppy-disk" />
                  </span>
                  <span>Save As</span>
                </button>
              </p>
              <p class="control">
                <input
                  style="width: 250px"
                  class="input is-small"
                  type="text"
                  bind:value={saveAsName}
                />
              </p>
            </div>

            <div class="field has-addons mb-0 ml-2">
              <p class="control">
                <button
                  class="button is-small is-danger"
                  on:click={(_e) => remove()}
                  disabled={$templateEditorState.hasUnsavedChanges == true ||
                    selectedTemplate?.template_type == "builtin"}
                >
                  <span class="icon is-small">
                    <i class="fas fa-floppy-disk" />
                  </span>
                  <span>Delete</span>
                </button>
              </p>
            </div>
          </div>

          <div class="field has-addons" />
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
              <button class="button is-small clipboard" on:click={copyToClipboard}>
                <span class="icon is-small">
                  <i class="fas fa-copy" />
                </span>
              </button>
              <div class="preview-editor" bind:this={previewEditorDom} />
            </div>
          </div>
        </div>
      </div>
      <div class="column is-7">
        <div class="box px-3">
          <Dropzone
            multiple={false}
            inputElement={input}
            accept=".csv,.txt,.xls,.xlsx"
            on:drop={handleFilesSelect}
          >
            Drag 'n' drop CSV, TXT, XLS, XLSX file here or click to select
          </Dropzone>
        </div>
        {#if !_.isEmpty(data)}
          <div class="table-wrapper">
            <table class="mt-2 table is-bordered is-size-7 is-narrow">
              <tr>
                <th />
                {#each _.range(0, columnCount) as ci}
                  <th class="has-background-light">{String.fromCharCode(65 + ci)}</th>
                {/each}
              </tr>
              {#each data as row, ri}
                <tr>
                  <td class="has-background-light"><b>{ri}</b></td>
                  {#each row as cell}
                    <td>{cell || ""}</td>
                  {/each}
                </tr>
              {/each}
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

  .clipboard {
    float: right;
    position: absolute;
    right: 0;
    z-index: 10;
  }

  .table-wrapper {
    overflow-x: auto;
  }
</style>
