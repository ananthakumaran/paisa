<script lang="ts">
  import { ajax, type LedgerFile, type Transaction as T } from "$lib/utils";
  import _ from "lodash";
  import { onMount } from "svelte";
  import VirtualList from "svelte-tiny-virtual-list";
  import Transaction from "$lib/components/Transaction.svelte";
  import BulkEditForm from "$lib/components/BulkEditForm.svelte";
  import { slide } from "svelte/transition";
  import * as bulkEdit from "$lib/bulk_edit";
  import * as toast from "bulma-toast";
  import DiffViewModal from "$lib/components/DiffViewModal.svelte";
  import SearchQuery from "$lib/components/SearchQuery.svelte";
  import { editorState } from "$lib/search_query_editor";
  import { get } from "svelte/store";

  let buldEditOpen = false;
  let transactions: T[] = null;
  let filtered: T[] = [];
  let files: LedgerFile[] = [];
  let newFiles: LedgerFile[] = [];
  let updatedTransactionsCount = 0;
  let openPreviewModal = false;
  let accounts: string[] = [];
  let commodities: string[] = [];

  const debits = (t: T) => {
    return _.filter(t.postings, (p) => p.amount < 0);
  };

  const credits = (t: T) => {
    return _.filter(t.postings, (p) => p.amount >= 0);
  };

  function handleInputRaw(predicate: (t: T) => boolean) {
    filtered = _.filter(transactions, predicate);
  }

  const handleInput = _.debounce(handleInputRaw, 100);

  editorState.subscribe((state) => {
    handleInput(state.predicate);
  });

  const itemSize = (i: number) => {
    const t = filtered[i];
    return 8 + Math.max(credits(t).length, debits(t).length) * 22;
  };

  async function loadTransactions() {
    ({ files, accounts, commodities } = await ajax("/api/editor/files"));
    ({ transactions } = await ajax("/api/transaction"));
    handleInputRaw(get(editorState).predicate);

    newFiles = files;
  }

  function showPreview(detail: any) {
    ({ newFiles, updatedTransactionsCount } = bulkEdit.applyChanges(
      files,
      filtered,
      detail.operation,
      detail.args
    ));
    openPreviewModal = true;
  }

  async function saveAll(newFiles: LedgerFile[]) {
    for (const newFile of newFiles) {
      const { saved, message } = await ajax("/api/editor/save", {
        method: "POST",
        body: JSON.stringify({ name: newFile.name, content: newFile.content })
      });

      if (!saved) {
        toast.toast({
          message: `Failed to save ${newFile.name}. reason: ${message}`,
          type: "is-danger",
          duration: 5000
        });
      } else {
        toast.toast({
          message: `Saved ${newFile.name}`,
          type: "is-success"
        });
      }
    }
    await loadTransactions();
  }

  onMount(async () => {
    await loadTransactions();
  });
</script>

<DiffViewModal
  on:save={(e) => saveAll(e.detail)}
  bind:open={openPreviewModal}
  oldFiles={files}
  {newFiles}
  {updatedTransactionsCount}
/>

{#if transactions}
  <section class="section tab-journal">
    <div class="container is-fluid">
      <div class="columns">
        <div class="column is-12">
          <nav class="level">
            <div class="level-left">
              <div class="level-item">
                <div class="field">
                  <div class="control">
                    <SearchQuery
                      autocomplete={{
                        account: accounts,
                        commodity: commodities,
                        filename: files.map((f) => f.name)
                      }}
                    />
                  </div>
                </div>
              </div>
              <div class="level-item">
                <div class="field">
                  <div class="control">
                    <button
                      class="button is-link is-light invertable"
                      on:click={(_e) => (buldEditOpen = !buldEditOpen)}
                    >
                      <span>Bulk Edit</span>
                      <span class="icon is-small">
                        <i class="fas {buldEditOpen ? 'fa-angle-up' : 'fa-angle-down'}"></i>
                      </span>
                    </button>
                  </div>
                </div>
              </div>
            </div>
            <div class="level-right">
              <div class="level-item">
                <p class="is-6"><b>{filtered.length}</b> transaction(s)</p>
              </div>
            </div>
          </nav>
        </div>
      </div>

      {#if buldEditOpen}
        <div class="columns">
          <div class="column is-12" transition:slide>
            <BulkEditForm {accounts} on:preview={(e) => showPreview(e.detail)} />
          </div>
        </div>
      {/if}

      <div class="columns">
        <div class="column is-12">
          <div class="box">
            <VirtualList
              width="100%"
              height={window.innerHeight - 150}
              itemCount={filtered.length}
              {itemSize}
            >
              <div slot="item" let:index let:style {style}>
                {@const t = filtered[index]}
                <Transaction {t} />
              </div>
            </VirtualList>
          </div>
        </div>
      </div>
    </div>
  </section>
{/if}
