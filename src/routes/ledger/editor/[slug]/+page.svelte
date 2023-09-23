<script lang="ts">
  import { createEditor, editorState, moveToEnd, moveToLine, updateContent } from "$lib/editor";
  import { ajax, buildLedgerTree, type LedgerFile } from "$lib/utils";
  import { redo, undo } from "@codemirror/commands";
  import * as toast from "bulma-toast";
  import type { EditorView } from "codemirror";
  import { format } from "$lib/journal";
  import _ from "lodash";
  import { onMount } from "svelte";
  import { beforeNavigate, goto } from "$app/navigation";
  import type { PageData } from "./$types";
  import FileTree from "$lib/components/FileTree.svelte";
  import FileModal from "$lib/components/FileModal.svelte";
  import { page } from "$app/stores";

  export let data: PageData;
  let editorDom: Element;
  let editor: EditorView;
  let filesMap: Record<string, LedgerFile> = {};
  let selectedFile: LedgerFile = null;
  let accounts: string[] = [];
  let commodities: string[] = [];
  let payees: string[] = [];
  let selectedVersion: string = null;
  let lineNumber = 0;

  let cancelled = false;
  beforeNavigate(async ({ cancel }) => {
    if ($editorState.hasUnsavedChanges) {
      const confirmed = confirm("You have unsaved changes. Are you sure you want to leave?");
      if (!confirmed) {
        cancel();
        cancelled = true;
      } else {
        $editorState = _.assign({}, $editorState, { hasUnsavedChanges: false });
      }
    }
  });

  async function navigate(url: string) {
    await goto(url, { noScroll: true });
    if (cancelled) {
      cancelled = false;
      return false;
    }
    return true;
  }

  onMount(async () => {
    loadFiles(data.name);
    const line = _.toNumber($page.url.hash.substring(1));
    if (_.isNumber(line)) {
      lineNumber = line;
    }
  });

  async function loadFiles(selectedFileName: string) {
    let files;
    ({ files, accounts, commodities, payees } = await ajax("/api/editor/files"));
    filesMap = _.fromPairs(_.map(files, (f) => [f.name, f]));
    if (!_.isEmpty(files)) {
      selectedFile = _.find(files, (f) => f.name == selectedFileName) || files[0];
    }
  }

  async function selectFile(file: LedgerFile) {
    const success = await navigate(`/ledger/editor/${encodeURIComponent(file.name)}`);
    if (success) {
      selectedFile = file;
    }
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
    const { errors, saved, file, message } = await ajax("/api/editor/save", {
      method: "POST",
      body: JSON.stringify({ name: selectedFile.name, content: doc.toString() })
    });

    if (!saved) {
      toast.toast({
        message: `Failed to save ${selectedFile.name}. reason: ${message}`,
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
      if (lineNumber > 0) {
        if (!editor.hasFocus) {
          editor.focus();
        }
        moveToLine(editor, lineNumber, true);
        lineNumber = 0;
      } else {
        moveToEnd(editor);
      }
    }
  }

  let modalOpen = false;
  function openCreateModal() {
    modalOpen = true;
  }

  async function createFile(destinationFile: string) {
    const { saved, message } = await ajax("/api/editor/save", {
      method: "POST",
      body: JSON.stringify({ name: destinationFile, content: "", operation: "create" })
    });

    if (saved) {
      toast.toast({
        message: `Created <b><a href="/ledger/editor/${encodeURIComponent(
          destinationFile
        )}">${destinationFile}</a></b>`,
        type: "is-success",
        duration: 5000
      });

      const success = await navigate(`/ledger/editor/${encodeURIComponent(destinationFile)}`);
      if (success) {
        await loadFiles(destinationFile);
      }
    } else {
      toast.toast({
        message: `Failed to create ${destinationFile}. reason: ${message}`,
        type: "is-danger",
        duration: 5000
      });
    }
  }
</script>

<FileModal bind:open={modalOpen} on:save={(e) => createFile(e.detail)} label="Create" help="" />

<section class="section tab-editor" style="padding-bottom: 0 !important">
  <div class="container is-fluid">
    <div class="columuns">
      <div class="column is-12 px-0 pt-0 mb-2">
        <div class="box p-3 is-flex is-align-items-center" style="width: 100%">
          <div class="field has-addons mb-0">
            <p class="control">
              <button
                class="button is-small is-link invertable is-light"
                disabled={$editorState.hasUnsavedChanges}
                on:click={(_e) => openCreateModal()}
              >
                <span class="icon is-small">
                  <i class="fas fa-file-circle-plus" />
                </span>
                <span>Create</span>
              </button>
            </p>
          </div>

          <div class="field has-addons ml-5 mb-0">
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
            <div class="field has-addons ml-5 mb-0">
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
            <div class="control ml-5">
              <a on:click={(_e) => moveToLine(editor, $editorState.errors[0].line_from)}
                ><span class="ml-1 tag invertable is-danger is-light"
                  >{$editorState.errors.length} error(s) found</span
                ></a
              >
            </div>
          {/if}
        </div>
      </div>
    </div>
    <div class="columns">
      <div class="column is-2">
        <div class="box px-2">
          <aside class="menu" style="max-height: calc(100vh - 185px)">
            <FileTree
              on:select={(e) => selectFile(e.detail)}
              files={buildLedgerTree(_.values(filesMap))}
              selectedFileName={selectedFile?.name}
              hasUnsavedChanges={$editorState.hasUnsavedChanges}
            />
          </aside>
        </div>
      </div>
      <div class="column is-6">
        <div class="box py-0">
          <div class="editor" bind:this={editorDom} />
        </div>
      </div>
      <div class="column is-4">
        {#if !_.isEmpty($editorState.output)}
          <pre style="max-height: calc(100vh - 185px)">{$editorState.output}</pre>
        {/if}
      </div>
    </div>
  </div>
</section>
