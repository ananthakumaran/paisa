<!--
  ImportPreview is the M3-A preview UI.

  Inputs:
    - state      : PreviewState (writable). Drives every rendered row.
    - accounts   : known accounts from /api/config; used to populate the
                   counterpart / source-account autocomplete.

  Outputs: the user can pick an importer, edit any cell, choose a source
  account + ledger file, and hit "Save N transactions". The actual HTTP
  calls live in $lib/importers/preview.ts — this component is purely
  presentational + event-wiring.

  Why "ledger file" is a text input rather than a dropdown: the importer
  flow lets users append to either an existing file (Assets/Bank/main.ledger)
  or a new one ("imports/2024-01-alipay.ledger"). Keep it free-form for
  power users; component callers can pass a sensible default.
-->
<script lang="ts">
  import type { Writable } from "svelte/store";
  import type { PreviewState } from "$lib/importers/preview";
  import type { ParsedTxn } from "$lib/importers";

  export let state: Writable<PreviewState> & {
    uploadFile: (file: File) => Promise<void>;
    pickImporter: (code: string) => Promise<void>;
    updateTxn: (index: number, patch: Partial<ParsedTxn>) => void;
    doCommit: (sourceAccount: string, ledgerFile: string) => Promise<void>;
    reset: () => void;
  };
  export let accounts: string[] = [];
  export let defaultLedgerFile: string = "main.ledger";

  let sourceAccount = "";
  let ledgerFile = defaultLedgerFile;

  // Convert a RFC3339-ish backend date to the value HTML5 <input type="date">
  // expects (yyyy-mm-dd). The backend round-trips either; we display the
  // short form because the UI lets the user pick a calendar date.
  function dateForInput(s: string): string {
    if (!s) return "";
    const idx = s.indexOf("T");
    return idx >= 0 ? s.slice(0, idx) : s;
  }
</script>

<div class="importer-preview">
  {#if $state.phase === "idle"}
    <p class="has-text-grey">
      Drop a statement file in the dropzone to start. Each registered importer will try to recognise
      the file; if none matches you can still pick one manually.
    </p>
  {/if}

  {#if $state.phase === "detecting" || $state.phase === "parsing"}
    <progress class="progress is-small is-link" max="100">…</progress>
  {/if}

  {#if $state.error}
    <div class="notification is-danger is-light">
      {$state.error}
    </div>
  {/if}

  {#if ($state.phase === "ready" || $state.phase === "committing") && $state.filename}
    <div class="field is-grouped is-align-items-center mb-3">
      <div class="control">
        <span class="tag is-info is-light">{$state.filename}</span>
      </div>
      <div class="control">
        <div class="select is-small">
          <select
            value={$state.selectedCode || ""}
            on:change={(e) => state.pickImporter(e.currentTarget.value)}
            disabled={$state.detected.length === 0}
          >
            {#if $state.detected.length === 0}
              <option value="">No importer detected</option>
            {/if}
            {#each $state.detected as imp}
              <option value={imp.code}>{imp.name}</option>
            {/each}
          </select>
        </div>
      </div>
      <div class="control">
        <button class="button is-small" on:click={() => state.reset()} type="button">Clear</button>
      </div>
    </div>

    <div class="columns is-vcentered mb-2">
      <div class="column is-6">
        <div class="field">
          <label class="label is-small" for="src-account">Source account</label>
          <div class="control">
            <input
              id="src-account"
              class="input is-small"
              list="accounts-list"
              bind:value={sourceAccount}
              placeholder="Assets:Bank:…"
            />
            <datalist id="accounts-list">
              {#each accounts as a}
                <option value={a}></option>
              {/each}
            </datalist>
          </div>
        </div>
      </div>
      <div class="column is-6">
        <div class="field">
          <label class="label is-small" for="ledger-file">Ledger file</label>
          <div class="control">
            <input
              id="ledger-file"
              class="input is-small"
              bind:value={ledgerFile}
              placeholder="main.ledger"
            />
          </div>
        </div>
      </div>
    </div>

    {#if $state.txns.length > 0}
      <div class="table-wrapper">
        <table class="table is-bordered is-narrow is-size-7 is-fullwidth">
          <thead>
            <tr>
              <th>Date</th>
              <th>Payee</th>
              <th>Amount</th>
              <th>Currency</th>
              <th>Counterpart</th>
              <th>Note</th>
            </tr>
          </thead>
          <tbody>
            {#each $state.txns as txn, i}
              <tr>
                <td>
                  <input
                    type="date"
                    class="input is-small"
                    value={dateForInput(txn.date)}
                    on:change={(e) => state.updateTxn(i, { date: e.currentTarget.value })}
                  />
                </td>
                <td>
                  <input
                    class="input is-small"
                    value={txn.payee}
                    on:input={(e) => state.updateTxn(i, { payee: e.currentTarget.value })}
                  />
                </td>
                <td>
                  <input
                    class="input is-small"
                    value={txn.amount}
                    on:input={(e) => state.updateTxn(i, { amount: e.currentTarget.value })}
                  />
                </td>
                <td>
                  <input
                    class="input is-small"
                    value={txn.currency}
                    on:input={(e) => state.updateTxn(i, { currency: e.currentTarget.value })}
                  />
                </td>
                <td>
                  <input
                    class="input is-small"
                    list="accounts-list"
                    value={txn.suggested_account}
                    on:input={(e) =>
                      state.updateTxn(i, { suggested_account: e.currentTarget.value })}
                  />
                </td>
                <td>
                  <input
                    class="input is-small"
                    value={txn.note}
                    on:input={(e) => state.updateTxn(i, { note: e.currentTarget.value })}
                  />
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      <div class="is-flex is-justify-content-flex-end mt-3">
        <button
          class="button is-link is-small"
          disabled={$state.phase === "committing" || !sourceAccount.trim() || !ledgerFile.trim()}
          on:click={() => state.doCommit(sourceAccount, ledgerFile)}
        >
          {$state.phase === "committing" ? "Saving…" : `Save ${$state.txns.length} transactions`}
        </button>
      </div>
    {/if}
  {/if}
</div>

<style lang="scss">
  .table-wrapper {
    overflow-x: auto;
    max-height: calc(100vh - 360px);
    overflow-y: auto;
  }

  .importer-preview .input.is-small {
    width: 100%;
  }
</style>
