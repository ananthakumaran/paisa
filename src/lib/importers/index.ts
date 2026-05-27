// Typed wrappers for the M3-A importer-framework HTTP endpoints.
//
// The framework lets us add bank/payment-platform importers without
// touching the UI: each new importer registers itself in the Go backend
// (`internal/importer/<bank>/<bank>.go`) and shows up automatically here
// through /api/import/detect. This file owns the wire format — keep it in
// lockstep with internal/server/import.go.

// We do NOT funnel through src/lib/utils.ts::ajax because (a) ajax has a
// hand-written overload table for every route and adding three more
// overloads there would balloon a 900-line file, and (b) the importer
// endpoints are entirely independent of the rest of the app. A small,
// dedicated wrapper keeps responsibilities clean.

export interface ParsedTxn {
  date: string; // RFC3339 from Go (json.Marshal of time.Time)
  payee: string;
  note: string;
  amount: string; // shopspring/decimal serialises as a JSON number; we
  // accept either string or number on the wire and normalise to string
  // at the boundary so the editable preview always works with a textual
  // representation.
  currency: string;
  suggested_account: string;
  raw_text: string;
}

export interface DetectedImporter {
  code: string;
  name: string;
}

// CommitTxn matches the wire format the backend expects. Date and amount
// are strings because the user may edit them in the preview before commit.
export interface CommitTxn {
  date: string;
  payee: string;
  note: string;
  amount: string;
  currency: string;
  suggested_account: string;
}

export interface CommitResponse {
  saved: boolean;
  count: number;
  errors: string[];
}

interface ApiError {
  error?: string;
}

async function postJSON<T>(route: string, body: unknown): Promise<T> {
  const res = await fetch(route, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body)
  });
  const text = await res.text();
  let parsed: T & ApiError;
  try {
    parsed = JSON.parse(text);
  } catch {
    throw new Error(`Non-JSON response (${res.status}): ${text.slice(0, 200)}`);
  }
  if (!res.ok) {
    throw new Error(parsed.error || `Request failed with ${res.status}`);
  }
  return parsed;
}

// Browser-safe base64 encoder for ArrayBuffer / Uint8Array. We avoid the
// `Buffer` shim because the SvelteKit app is shipped as static assets and
// never runs through a bundler that polyfills Node globals.
export function toBase64(bytes: Uint8Array): string {
  let binary = "";
  const chunk = 0x8000;
  for (let i = 0; i < bytes.length; i += chunk) {
    binary += String.fromCharCode.apply(null, bytes.subarray(i, i + chunk) as unknown as number[]);
  }
  return btoa(binary);
}

export async function detect(filename: string, bytes: Uint8Array): Promise<DetectedImporter[]> {
  const { importers } = await postJSON<{ importers: DetectedImporter[] }>("/api/import/detect", {
    filename,
    content_base64: toBase64(bytes)
  });
  return importers || [];
}

export async function parse(importerCode: string, bytes: Uint8Array): Promise<ParsedTxn[]> {
  const { transactions } = await postJSON<{ transactions: ParsedTxn[] }>("/api/import/parse", {
    importer_code: importerCode,
    content_base64: toBase64(bytes)
  });
  // shopspring/decimal occasionally emits unquoted JSON numbers (because
  // MarshalJSONWithoutQuotes = true). Normalise to string at the boundary so
  // the rest of the UI doesn't have to care.
  return (transactions || []).map((t) => ({
    ...t,
    amount: String(t.amount)
  }));
}

export async function commit(
  sourceAccount: string,
  ledgerFile: string,
  txns: CommitTxn[]
): Promise<CommitResponse> {
  return postJSON<CommitResponse>("/api/import/commit", {
    source_account: sourceAccount,
    ledger_file: ledgerFile,
    txns
  });
}
