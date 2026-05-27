// Preview state machine for the M3-A importer pipeline.
//
// Phases:
//   idle      -> nothing uploaded yet
//   detecting -> /api/import/detect is in flight
//   parsing   -> user picked an importer, /api/import/parse in flight
//   ready     -> rows in `txns` are editable; "Commit" is enabled
//   committing-> /api/import/commit in flight
//   error     -> any of the above failed; `error` field is set
//
// Splitting the state into a tiny FSM (rather than free-floating booleans
// like `isLoading`, `isDetecting`, …) prevents impossible combinations
// (`isDetecting && isParsing == true`) from being rendered. The Svelte UI
// switches on `state.phase` to pick what to render.

import { writable, type Writable } from "svelte/store";
import { commit, detect, parse, type CommitTxn, type DetectedImporter, type ParsedTxn } from ".";

export type Phase = "idle" | "detecting" | "parsing" | "ready" | "committing" | "error";

export interface PreviewState {
  phase: Phase;
  filename: string;
  /** Raw file bytes; kept so the user can pick a different importer
   *  without re-uploading. */
  bytes: Uint8Array | null;
  detected: DetectedImporter[];
  /** Currently selected importer code. May differ from detected[0] if the
   *  user overrode the suggestion. */
  selectedCode: string | null;
  txns: ParsedTxn[];
  error: string | null;
  /** Number of transactions written in the most recent successful commit.
   *  Drives the "Saved N transactions" toast. */
  lastCommitCount: number;
}

const initial: PreviewState = {
  phase: "idle",
  filename: "",
  bytes: null,
  detected: [],
  selectedCode: null,
  txns: [],
  error: null,
  lastCommitCount: 0
};

export function createPreviewStore(): Writable<PreviewState> & {
  uploadFile: (file: File) => Promise<void>;
  pickImporter: (code: string) => Promise<void>;
  updateTxn: (index: number, patch: Partial<ParsedTxn>) => void;
  doCommit: (sourceAccount: string, ledgerFile: string) => Promise<void>;
  reset: () => void;
} {
  const store = writable<PreviewState>(initial);

  // Helper to apply a mutation atomically; avoids subscribing/unsubscribing
  // each time we want to read-then-write.
  function patch(p: Partial<PreviewState>) {
    store.update((s) => ({ ...s, ...p }));
  }

  async function uploadFile(file: File) {
    const buf = new Uint8Array(await file.arrayBuffer());
    patch({
      phase: "detecting",
      filename: file.name,
      bytes: buf,
      detected: [],
      selectedCode: null,
      txns: [],
      error: null
    });
    try {
      const detected = await detect(file.name, buf);
      // Whether or not anything matched, transition to "ready": the preview
      // UI shows a "No importer detected" disabled option for the empty
      // case so the user can still see their file is uploaded. Future work
      // (issue M3-B+) may add a manual-override list of all registered
      // importers; for now the surface stays minimal.
      patch({ phase: "ready", detected });
      // Auto-select the first detected importer; user can switch in the UI.
      if (detected.length > 0) {
        await pickImporter(detected[0].code);
      }
    } catch (e) {
      patch({ phase: "error", error: errorMessage(e) });
    }
  }

  async function pickImporter(code: string) {
    let bytes: Uint8Array | null = null;
    store.update((s) => {
      bytes = s.bytes;
      return { ...s, phase: "parsing", selectedCode: code, error: null };
    });
    if (!bytes) {
      patch({ phase: "error", error: "no file uploaded" });
      return;
    }
    try {
      const txns = await parse(code, bytes);
      patch({ phase: "ready", txns });
    } catch (e) {
      patch({ phase: "error", error: errorMessage(e) });
    }
  }

  function updateTxn(index: number, p: Partial<ParsedTxn>) {
    store.update((s) => {
      const next = s.txns.slice();
      if (index < 0 || index >= next.length) return s;
      next[index] = { ...next[index], ...p };
      return { ...s, txns: next };
    });
  }

  async function doCommit(sourceAccount: string, ledgerFile: string) {
    let txns: ParsedTxn[] = [];
    store.update((s) => {
      txns = s.txns;
      return { ...s, phase: "committing", error: null };
    });
    const commitBody: CommitTxn[] = txns.map((t) => ({
      date: t.date,
      payee: t.payee,
      note: t.note,
      amount: t.amount,
      currency: t.currency,
      suggested_account: t.suggested_account
    }));
    try {
      const res = await commit(sourceAccount, ledgerFile, commitBody);
      if (!res.saved) {
        patch({
          phase: "error",
          error: (res.errors || []).join("; ") || "commit failed"
        });
        return;
      }
      // Successful commit returns to idle so the user can upload another
      // file. The last count survives in `lastCommitCount` for the toast.
      store.set({ ...initial, lastCommitCount: res.count });
    } catch (e) {
      patch({ phase: "error", error: errorMessage(e) });
    }
  }

  function reset() {
    store.set(initial);
  }

  return Object.assign(store, { uploadFile, pickImporter, updateTxn, doCommit, reset });
}

function errorMessage(e: unknown): string {
  if (e instanceof Error) return e.message;
  return String(e);
}
