import _ from "lodash";

// M3-G (#25) refund-aware helpers. Kept free of `$lib/utils` so the
// unit tests under bun:test can import this file directly without
// pulling in `$app/navigation` (which only resolves through Vite).
//
// In paisa's convention an Expenses:* posting is positive for a real
// expense and negative when it records a refund / 红冲 against an
// earlier expense (style A in issue #25, sibling of style B
// `Income:Refund:<category>` which doesn't appear in Expenses at all).

export interface PostingLike {
  amount: number;
}

export function isRefundPosting(posting: PostingLike): boolean {
  return posting.amount < 0;
}

// filterRefunds returns the postings to render under the requested
// view mode. `showGross = true` drops refund postings so the chart
// shows the original (gross) outflow; `showGross = false` (the
// default) keeps them so the chart shows the user's actual net spend
// — that's the "扣除退款后净支出" view the user asked for in #25.
export function filterRefunds<P extends PostingLike>(postings: P[], showGross: boolean): P[] {
  if (!showGross) return postings;
  return _.filter(postings, (p) => !isRefundPosting(p));
}
