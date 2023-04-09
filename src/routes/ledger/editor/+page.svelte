<script lang="ts">
  import { createEditor, editorState, moveToEnd, moveToLine } from "$lib/editor";
  import { ajax, type LedgerFile } from "$lib/utils";
  import { redo, undo } from "@codemirror/commands";
  import * as toast from "bulma-toast";
  import type { EditorView } from "codemirror";
  import _ from "lodash";
  import { onMount } from "svelte";

  let editorDom: Element;
  let editor: EditorView;
  let files: LedgerFile[] = [];
  let selectedFile: string = null;

  onMount(async () => {
    ({ files } = await ajax("/api/editor/files"));
    if (!_.isEmpty(files)) {
      selectedFile = files[0].name;
    }
  });

  async function save() {
    const doc = editor.state.doc;
    const { errors, saved } = await ajax("/api/editor/save", {
      method: "POST",
      body: JSON.stringify({ name: selectedFile, content: doc.toString() })
    });

    if (!saved) {
      toast.toast({
        message: `Failed to save ${selectedFile}`,
        type: "is-danger",
        dismissible: true,
        duration: 5000
      });
    } else {
      toast.toast({
        message: `Saved ${selectedFile}`,
        type: "is-success",
        dismissible: true
      });

      $editorState = _.assign({}, $editorState, { hasUnsavedChanges: false });
    }

    if (!_.isEmpty(errors)) {
      moveToLine(editor, errors[0].line_from);
    }
  }

  $: if (selectedFile) {
    if (editor) {
      editor.destroy();
    }
    const file = _.find(files, (f) => f.name == selectedFile);

    editor = createEditor(file, editorDom);
    moveToEnd(editor);
  }
</script>

<section class="section tab-editor">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 pb-0">
        <div class="tabs is-boxed">
          <ul>
            {#each files as file}
              <li class:is-active={file.name == selectedFile}>
                <a on:click={(_e) => (selectedFile = file.name)}>
                  <span
                    >{file.name}
                    {#if file.name == selectedFile && $editorState.hasUnsavedChanges}
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
      <div class="column is-12 px-0 pt-0 is-flex is-align-items-center">
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
        </div>
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
    <div class="columns">
      <div class="column is-12">
        <div class="editor" bind:this={editorDom} />
      </div>
    </div>
  </div>
</section>
