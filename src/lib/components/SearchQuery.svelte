<script lang="ts">
  import { createEditor } from "$lib/search_query_editor";
  import type { EditorView } from "codemirror";

  let editorDom: HTMLElement;
  export let autocomplete: Record<string, string[]>;
  let editor: EditorView;

  $: if (autocomplete && editorDom) {
    if (editor) {
      editor.destroy();
    }

    editor = createEditor("", editorDom, autocomplete);
  }
</script>

<div class="search-query-editor" bind:this={editorDom} />
