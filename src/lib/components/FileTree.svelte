<script lang="ts">
  import type { LedgerFile } from "$lib/utils";
  import _ from "lodash";
  import { createEventDispatcher } from "svelte";

  export let ledgerFiles: LedgerFile[] = [];
  export let selectedFileName: string;
  export let hasUnsavedChanges: boolean;

  const dispatch = createEventDispatcher();
</script>

<ul class="menu-list is-size-6">
  {#each _.sortBy(ledgerFiles, (l) => l.name) as file}
    <li>
      <a on:click={() => dispatch("select", file)} class:is-active={file.name == selectedFileName}
        >{file.name}
        {#if file.name == selectedFileName && hasUnsavedChanges}
          <span class="ml-1 tag is-danger">unsaved</span>
        {/if}
      </a>
    </li>
  {/each}
</ul>
