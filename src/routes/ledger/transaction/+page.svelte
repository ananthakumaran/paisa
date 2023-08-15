<script lang="ts">
  import { ajax, type LedgerFile, type Transaction as T } from "$lib/utils";
  import { filterTransactions } from "$lib/transaction";
  import _ from "lodash";
  import { onMount } from "svelte";
  import VirtualList from "svelte-tiny-virtual-list";
  import Transaction from "$lib/components/Transaction.svelte";
  import BulkEditForm from "$lib/components/BulkEditForm.svelte";
  import { slide } from "svelte/transition";
  import * as bulkEdit from "$lib/bulk_edit";
  import * as toast from "bulma-toast";
  import DiffViewModal from "$lib/components/DiffViewModal.svelte";

  let buldEditOpen = false;
  let transactions: T[] = [];
  let filtered: T[] = [];
  let files: LedgerFile[] = [];
  let newFiles: LedgerFile[] = [];
  let updatedTransactionsCount = 0;
  let openPreviewModal = false;
  let filter: string;
  let accounts: string[] = [];

  const debits = (t: T) => {
    return _.filter(t.postings, (p) => p.amount < 0);
  };

  const credits = (t: T) => {
    return _.filter(t.postings, (p) => p.amount >= 0);
  };

  const handleInput = _.debounce((event) => {
    filter = event.srcElement.value;
    filtered = filterTransactions(transactions, filter);
  }, 100);

  const itemSize = (i: number) => {
    const t = filtered[i];
    return 2 + Math.max(credits(t).length, debits(t).length) * 30;
  };

  async function loadTransactions() {
    ({ files, accounts } = await ajax("/api/editor/files"));
    ({ transactions } = await ajax("/api/transaction"));
    filtered = filterTransactions(transactions, filter);

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

<section class="section tab-journal">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <nav class="level">
          <div class="level-left">
            <div class="level-item">
              <div class="field">
                <p class="control">
                  <input
                    class="d3-transaction-filter input"
                    style="width: 500px"
                    type="text"
                    placeholder="filter by account or description or date"
                    on:input={handleInput}
                  />
                </p>
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
