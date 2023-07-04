<script lang="ts">
  import { createEditor, editorState, moveToEnd, moveToLine, updateContent } from "$lib/editor";
  import { ajax, type LedgerFile } from "$lib/utils";
  import { redo, undo } from "@codemirror/commands";
  import * as toast from "bulma-toast";
  import type { EditorView } from "codemirror";
  import { format } from "$lib/journal";
  import _ from "lodash";
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";

  export let data: { name: string };
  let editorDom: Element;
  let editor: EditorView;
  let filesMap: Record<string, LedgerFile> = {};
  let selectedFile: LedgerFile = null;
  let accounts: string[] = [];
  let commodities: string[] = [];
  let payees: string[] = [];
  let selectedVersion: string = null;

  onMount(async () => {
    let files;
    ({ files, accounts, commodities, payees } = await ajax("/api/editor/files"));
    filesMap = _.fromPairs(_.map(files, (f) => [f.name, f]));
    if (!_.isEmpty(files)) {
      selectedFile = _.find(files, (f) => f.name == data.name) || files[0];
    }
  });

  async function selectFile(file: LedgerFile) {
    selectedFile = file;
    await goto(`/ledger/editor/${file.name}`, { noScroll: true });
  }

  async function revert(version: string) {
    const { file } = await ajax("/api/editor/file", {
      method: "POST",
      body: JSON.stringify({ name: version })
    });

    updateContent(editor, file.content);
  }

  async function pretty() {
    const formatted = format(editor.state.doc.toString());
    if (formatted != editor.state.doc.toString()) {
      updateContent(editor, formatted);
    }
  }

  async function deleteBackups() {
    const { file } = await ajax("/api/editor/file/delete_backups", {
      method: "POST",
      body: JSON.stringify({ name: selectedFile.name })
    });

    selectedFile.versions = file.versions;
  }

  async function save() {
    const doc = editor.state.doc;
    const { errors, saved, file } = await ajax("/api/editor/save", {
      method: "POST",
      body: JSON.stringify({ name: selectedFile.name, content: doc.toString() })
    });

    if (!saved) {
      toast.toast({
        message: `Failed to save ${selectedFile.name}`,
        type: "is-danger",
        duration: 5000
      });
      if (!_.isEmpty(errors)) {
        moveToLine(editor, errors[0].line_from);
      }
    } else {
      toast.toast({
        message: `Saved ${selectedFile.name}`,
        type: "is-success"
      });
      filesMap[file.name] = file;
      selectedFile = file;
      selectedVersion = null;
      $editorState = _.assign({}, $editorState, { hasUnsavedChanges: false });
    }
  }

  $: if (selectedFile) {
    if (!editor || editor.state.doc.toString() != selectedFile.content) {
      if (editor) {
        editor.destroy();
      }

      editor = createEditor(selectedFile.content, editorDom, {
        autocompletions: {
          string: accounts,
          strong: payees,
          unit: commodities
        }
      });
      moveToEnd(editor);
    }
  }
</script>

<section class="section tab-editor" style="padding-bottom: 0 !important">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 py-0">
        <div class="tabs is-boxed">
          <ul>
            {#each _.values(filesMap) as file}
              <li class:is-active={file.name == selectedFile.name}>
                <a on:click={(_e) => selectFile(file)}>
                  <span
                    >{file.name}
                    {#if file.name == selectedFile.name && $editorState.hasUnsavedChanges}
                      <span class="ml-1 tag is-danger">unsaved</span>
                    {/if}
                  </span>
                </a>
              </li>
            {/each}
          </ul>
        </div>
      </div>
    </div>
    <div class="columuns">
      <div class="column is-12 px-0 pt-0">
        <div class="box p-3 is-flex is-align-items-center" style="width: 100%">
          <div class="field has-addons mb-0">
            <p class="control">
              <button
                class="button is-small"
                disabled={$editorState.hasUnsavedChanges == false}
                on:click={(_e) => save()}
              >
                <span class="icon is-small">
                  <i class="fas fa-floppy-disk" />
                </span>
                <span>Save</span>
              </button>
            </p>
            <p class="control">
              <button
                class="button is-small"
                disabled={$editorState.undoDepth == 0}
                on:click={(_e) => undo(editor)}
              >
                <span class="icon is-small">
                  <i class="fas fa-arrow-left" />
                </span>
                <span>Undo</span>
              </button>
            </p>
            <p class="control">
              <button
                class="button is-small"
                disabled={$editorState.redoDepth == 0}
                on:click={(_e) => redo(editor)}
              >
                <span>Redo</span>
                <span class="icon is-small">
                  <i class="fas fa-arrow-right" />
                </span>
              </button>
            </p>
            <p class="control">
              <button class="button is-small" on:click={(_e) => pretty()}>
                <span class="icon is-small">
                  <i class="fas fa-code" />
                </span>
                <span>Prettify</span>
              </button>
            </p>
          </div>

          {#if !_.isEmpty(selectedFile?.versions)}
            <div class="field has-addons ml-2 mb-0">
              <p class="control">
                <button
                  class="button is-small"
                  disabled={!selectedVersion}
                  on:click={(_e) => revert(selectedVersion)}
                >
                  <span class="icon is-small">
                    <i class="fas fa-clock-rotate-left" />
                  </span>
                  <span>Revert</span>
                </button>
              </p>

              <div class="control">
                <div class="select is-small">
                  <select bind:value={selectedVersion}>
                    {#each selectedFile.versions as version}
                      <option>{version}</option>
                    {/each}
                  </select>
                </div>
              </div>

              <p class="control">
                <button class="button is-small" on:click={(_e) => deleteBackups()}>
                  <span class="icon is-small">
                    <i class="fas fa-trash" />
                  </span>
                </button>
              </p>
            </div>
          {/if}

          {#if $editorState.errors.length > 0}
            <div class="control">
              <a on:click={(_e) => moveToLine(editor, $editorState.errors[0].line_from)}
                ><span class="ml-1 tag is-danger is-light"
                  >{$editorState.errors.length} error(s) found</span
                ></a
              >
            </div>
          {/if}
        </div>
      </div>
    </div>
    <div class="columns">
      <div class="column is-12">
        <div class="box py-0">
          <div class="editor" bind:this={editorDom} />
        </div>
      </div>
    </div>
  </div>
</section>
