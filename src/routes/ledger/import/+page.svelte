<script lang="ts">
  import {
    createEditor as createTemplateEditor,
    editorState as templateEditorState
  } from "$lib/template_editor";
  import {
    createEditor as createPreviewEditor,
    updateContent as updatePreviewContent
  } from "$lib/editor";
  import Dropzone from "svelte-file-dropzone/Dropzone.svelte";
  import Papa from "papaparse";
  import _ from "lodash";
  import type { EditorView } from "codemirror";
  import { onMount } from "svelte";
  import { ajax } from "$lib/utils";
  import { accountTfIdf } from "../../../store";
  import * as toast from "bulma-toast";

  let lastTemplate: any;
  let lastData: any;
  let template = `{{#if (and (isDate ROW.A "DD/MM/YYYY") (isBlank ROW.G))}}
 {{date ROW.A "DD/MM/YYYY"}} {{ROW.C}}
    {{predictAccount prefix="Expenses"}}		{{ROW.F}} INR
    Assets:Checking
{{/if}}`;
  let preview = "";
  let output: string[] = [];
  let columnCount: number;
  let data: any[][] = [];
  let rows: Array<Record<string, any>> = [];

  let templateEditorDom: Element;

  let previewEditorDom: Element;
  let previewEditor: EditorView;

  onMount(async () => {
    accountTfIdf.set(await ajax("/api/account/tf_idf"));
    createTemplateEditor(template, templateEditorDom);
    previewEditor = createPreviewEditor(preview, previewEditorDom, { readonly: true });
  });

  const papaPromise = (file: File) =>
    new Promise((resolve, reject) => {
      Papa.parse(file, {
        skipEmptyLines: true,
        complete: function (results) {
          resolve(results);
        },
        error: function (error) {
          reject(error);
        },
        delimitersToGuess: [",", "\t", "|", ";", Papa.RECORD_SEP, Papa.UNIT_SEP, "^"]
      });
    });

  let input: any;

  $: if (!_.isEmpty(data) && $templateEditorState.template) {
    if (lastTemplate != $templateEditorState.template || lastData != data) {
      output = [];
      try {
        _.each(rows, (row) => {
          const rendered = _.trim($templateEditorState.template({ ROW: row }));
          if (!_.isEmpty(rendered)) {
            output.push(rendered);
          }
        });
        preview = output.join("\n\n");
        updatePreviewContent(previewEditor, preview);
        lastTemplate = $templateEditorState.template;
        lastData = data;
      } catch (e) {
        console.log(e);
      }
    }
  }

  async function handleFilesSelect(e: { detail: { acceptedFiles: File[] } }) {
    const { acceptedFiles } = e.detail;

    const results: any = await papaPromise(acceptedFiles[0]);
    data = results.data;

    rows = _.map(data, (row) => {
      return _.chain(row)
        .map((cell, j) => {
          return [String.fromCharCode(65 + j), cell];
        })
        .fromPairs()
        .value();
    });

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
        <div class="field">
          <div class="control">
            <div class="template-editor" bind:this={templateEditorDom} />
          </div>
        </div>
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
      <div class="column is-7">
        <Dropzone
          multiple={false}
          inputElement={input}
          accept=".csv,.txt"
          on:drop={handleFilesSelect}
        >
          Drag 'n' drop CSV, TXT file here or click to select
        </Dropzone>
        {#if !_.isEmpty(data)}
          <div class="table-wrapper">
            <table class="mt-2 table is-bordered is-size-7 is-narrow">
              <tr>
                <th />
                {#each _.range(0, columnCount) as ci}
                  <th>{String.fromCharCode(65 + ci)}</th>
                {/each}
              </tr>
              {#each data as row, ri}
                <tr>
                  <td><b>{ri + 1}</b></td>
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
  .clipboard {
    float: right;
    position: absolute;
    right: 0;
    margin-right: 20px;
    z-index: 10;
  }

  .table-wrapper {
    overflow-x: scroll;
  }
</style>
