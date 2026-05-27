// Helpers for the kind-grouped Allocation tree shape produced by the
// `/api/allocation` endpoint after M1-E. Kept in its own module so it can
// be unit-tested under bun without pulling in the SvelteKit-only modules
// transitively imported by `./utils`.
import { getKindLabel, isValidAccountKind, type AccountKindValue } from "./account_kind";

// ALLOCATION_ROOT mirrors `allocationRootKey` in
// internal/server/allocation.go. The backend wraps every Aggregate in a
// synthetic colon path beginning with this prefix so the d3 hierarchy is
// rooted on a stable, kind-grouped tree.
export const ALLOCATION_ROOT = "Allocation";

// Minimal Aggregate shape this module reads from. We don't re-import the
// canonical `Aggregate` interface to avoid pulling in `./utils` here.
export interface AllocationNode {
  account?: string;
  original_account?: string;
  kind?: string;
}

// allocationNodeLabel returns the user-visible label for one node in the
// kind-grouped allocation tree. The shape of each node id is:
//
//   "Allocation"                              → root (rarely rendered)
//   "Allocation:<kind>"                       → kind bucket → localized label
//   "Allocation:<kind>:<sanitized-account>"   → leaf → real ledger account
//
// `aggregate` (when available) carries `original_account` / `kind` straight
// from the backend; we prefer those over re-parsing the synthetic id.
export function allocationNodeLabel(id: string, aggregate?: AllocationNode): string {
  if (id === ALLOCATION_ROOT) return ALLOCATION_ROOT;

  if (aggregate?.original_account) {
    return aggregate.original_account;
  }

  const parts = id.split(":");
  // Bucket node: Allocation:<kind>
  if (parts.length === 2 && parts[0] === ALLOCATION_ROOT) {
    const kind = parts[1];
    if (isValidAccountKind(kind)) {
      return getKindLabel(kind as AccountKindValue);
    }
    return kind;
  }
  // Leaf without a populated original_account — fall back to the last
  // segment of the synthetic id (the sanitizer replaces ':' with '/' so
  // this is the full original account with '/' separators).
  return parts[parts.length - 1] ?? id;
}

// timelineGroupLabel converts the kind code that
// `renderAllocationTimeline` uses as a series key into a localized legend
// label. Unknown codes pass through unchanged so the legend never blanks
// out on a forward-compatible payload.
export function timelineGroupLabel(kindCode: string): string {
  return isValidAccountKind(kindCode) ? getKindLabel(kindCode as AccountKindValue) : kindCode;
}
