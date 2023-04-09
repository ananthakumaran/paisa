<script lang="ts">
  import { createEditor, moveToEnd, moveToLine } from "$lib/editor";
  import { ajax, type LedgerFile } from "$lib/utils";
  import type { EditorView } from "codemirror";
  import * as toast from "bulma-toast";
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
                  <span>{file.name}</span>
                </a>
              </li>
            {/each}
          </ul>
        </div>
      </div>
    </div>
    <div class="columuns">
      <div class="column is-12 px-0 pt-0">
        <div class="field has-addons">
          <p class="control">
            <button class="button is-small" on:click={(_e) => save()}>
              <span class="icon is-small">
                <i class="fas fa-floppy-disk" />
              </span>
              <span>Save</span>
            </button>
          </p>
        </div>
      </div>
    </div>
    <div class="columns">
      <div class="column is-12">
        <div class="editor" bind:this={editorDom} />
      </div>
    </div>
  </div>
</section>
