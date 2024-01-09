<script lang="ts">
  import { createEditor, sheetEditorState } from "$lib/sheet";
  import { moveToLine, updateContent } from "$lib/editor";
  import {
    ajax,
    buildDirectoryTree,
    formatFloatUptoPrecision,
    type Posting,
    type SheetFile
  } from "$lib/utils";
  import { redo, undo } from "@codemirror/commands";
  import type { KeyBinding } from "@codemirror/view";
  import * as toast from "bulma-toast";
  import type { EditorView } from "codemirror";
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
  let filesMap: Record<string, SheetFile> = {};
  let postings: Posting[] = [];
  let selectedFile: SheetFile = null;
  let selectedVersion: string = null;
  let lineNumber = 0;

  function command(fn: Function) {
    return () => {
      fn();
      return true;
    };
  }

  const keybindings: readonly KeyBinding[] = [
    {
      key: "Ctrl-s",
      run: command(save),
      preventDefault: true
    }
  ];

  let cancelled = false;
  beforeNavigate(async ({ cancel }) => {
    if ($sheetEditorState.hasUnsavedChanges) {
      const confirmed = confirm("You have unsaved changes. Are you sure you want to leave?");
      if (!confirmed) {
        cancel();
        cancelled = true;
      } else {
        $sheetEditorState = _.assign({}, $sheetEditorState, { hasUnsavedChanges: false });
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
    ({ files, postings } = await ajax("/api/sheets/files"));
    filesMap = _.fromPairs(_.map(files, (f) => [f.name, f]));
    if (!_.isEmpty(files)) {
      selectedFile = _.find(files, (f) => f.name == selectedFileName) || files[0];
    }
  }

  async function selectFile(file: SheetFile) {
    const success = await navigate(`/more/sheets/${encodeURIComponent(file.name)}`);
    if (success) {
      selectedFile = file;
    }
  }

  async function revert(version: string) {
    const { file } = await ajax("/api/sheets/file", {
      method: "POST",
      body: JSON.stringify({ name: version })
    });

    updateContent(editor, file.content);
  }

  async function deleteBackups() {
    const { file } = await ajax("/api/sheets/file/delete_backups", {
      method: "POST",
      body: JSON.stringify({ name: selectedFile.name })
    });

    selectedFile.versions = file.versions;
  }

  async function save() {
    const doc = editor.state.doc;
    const { saved, file, message } = await ajax("/api/sheets/save", {
      method: "POST",
      body: JSON.stringify({ name: selectedFile.name, content: doc.toString() })
    });

    if (!saved) {
      toast.toast({
        message: `Failed to save ${selectedFile.name}. reason: ${message}`,
        type: "is-danger",
        duration: 10000
      });
    } else {
      toast.toast({
        message: `Saved ${selectedFile.name}`,
        type: "is-success"
      });
      filesMap[file.name] = file;
      selectedFile = file;
      selectedVersion = null;
      $sheetEditorState = _.assign({}, $sheetEditorState, { hasUnsavedChanges: false });
    }
  }

  $: if (selectedFile) {
    if (!editor || editor.state.doc.toString() != selectedFile.content) {
      if (editor) {
        editor.destroy();
      }

      editor = createEditor(selectedFile.content, editorDom, postings, { keybindings });
      if (lineNumber > 0) {
        if (!editor.hasFocus) {
          editor.focus();
        }
        moveToLine(editor, lineNumber, true);
        lineNumber = 0;
      }
    }
  }

  let modalOpen = false;
  function openCreateModal() {
    modalOpen = true;
  }

  async function createFile(destinationFile: string) {
    destinationFile = destinationFile.trim() + ".paisa";
    const { saved, message } = await ajax("/api/sheets/save", {
      method: "POST",
      body: JSON.stringify({ name: destinationFile, content: "", operation: "create" })
    });

    if (saved) {
      toast.toast({
        message: `Created <b><a href="/more/sheets/${encodeURIComponent(
          destinationFile
        )}">${destinationFile}</a></b>`,
        type: "is-success",
        duration: 5000
      });

      const success = await navigate(`/more/sheets/${encodeURIComponent(destinationFile)}`);
      if (success) {
        await loadFiles(destinationFile);
      }
    } else {
      toast.toast({
        message: `Failed to create ${destinationFile}. reason: ${message}`,
        type: "is-danger",
        duration: 10000
      });
    }
  }
</script>

<FileModal
  bind:open={modalOpen}
  on:save={(e) => createFile(e.detail)}
  label="Create"
  placeholder="scratch"
  help="Filename without any extension"
/>

<section class="section tab-editor max-h-screen" style="padding-bottom: 0 !important">
  <div class="container is-fluid">
    <div class="columuns">
      <div class="column is-12 px-0 pt-0 mb-2">
        <div class="box p-3 is-flex is-align-items-center overflow-x-auto" style="width: 100%">
          <div class="field has-addons mb-0">
            <p class="control">
              <button
                class="button is-small is-link invertable is-light"
                disabled={$sheetEditorState.hasUnsavedChanges}
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
                disabled={$sheetEditorState.hasUnsavedChanges == false}
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
                disabled={$sheetEditorState.undoDepth == 0}
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
                disabled={$sheetEditorState.redoDepth == 0}
                on:click={(_e) => redo(editor)}
              >
                <span>Redo</span>
                <span class="icon is-small">
                  <i class="fas fa-arrow-right" />
                </span>
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

          {#if $sheetEditorState.errors.length > 0}
            <div class="control ml-5">
              <a on:click={(_e) => moveToLine(editor, $sheetEditorState.errors[0].line_from)}
                ><span class="ml-1 tag invertable is-danger is-light"
                  >{$sheetEditorState.errors.length} error(s) found</span
                ></a
              >
            </div>
          {/if}

          <div class="control ml-5">
            <button
              class:is-loading={$sheetEditorState.pendingEval}
              class="is-loading button is-small pointer-events-none px-0"
              style="border: none"
            >
              {formatFloatUptoPrecision($sheetEditorState.evalDuration, 2)}ms
            </button>
          </div>
        </div>
      </div>
    </div>
    <div class="columns">
      <div class="column is-3-widescreen is-2-fullhd is-4">
        <div class="box px-2 full-height overflow-y-auto">
          <aside class="menu">
            <FileTree
              path=""
              on:select={(e) => selectFile(e.detail)}
              files={buildDirectoryTree(_.values(filesMap))}
              selectedFileName={selectedFile?.name}
              hasUnsavedChanges={$sheetEditorState.hasUnsavedChanges}
            />
          </aside>
        </div>
      </div>
      <div class="column is-9-widescreen is-10-fullhd is-8">
        <div class="flex">
          <div class="box box-r-none py-0 pr-1 mb-0 basis-[36rem] max-w-[48rem]">
            <div class="sheet-editor" bind:this={editorDom} />
          </div>
          <div
            class="box box-l-none has-text-right sheet-result"
            style="padding: 4px 0; width: 200px;"
          >
            {#each $sheetEditorState.results as result, i}
              <div
                class={i + 1 === $sheetEditorState.currentLine
                  ? "has-background-grey-lightest has-text-grey-dark has-text-weight-bold"
                  : ""}
                style="padding: 0 0.5rem"
              >
                <div
                  title={result.result}
                  class:underline={result.underline}
                  class:font-bold={result.bold}
                  class:text-left={result.align === "left"}
                  class="m-0 p-0 truncate {result.error ? 'has-text-danger' : ''}"
                  style="font-size: 0.9285714285714286rem; line-height: 1.4"
                >
                  &nbsp;{result.result}
                </div>
              </div>
            {/each}
          </div>
        </div>
      </div>
    </div>
  </div>
</section>
