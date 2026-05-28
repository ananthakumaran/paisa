// Pure helpers for sizing the /assets/gain SVG bar chart's left margin so that
// long account-name Y-axis labels (notably CJK account paths like
// "Bridge:Brokerage:DanJuan" or "Loans Receivable:LiQuanRong") are not clipped.
//
// Kept in a stand-alone module so it can be unit-tested without pulling in
// gain.ts's SvelteKit / d3 / DOM dependencies.

// Heuristic per-character widths at 12px sans-serif. d3.axisLeft uses 10px by
// default but in this codebase the global svg text style is 12px (see
// src/lib/utils.ts measurements + the screenshot in issue #64). These are
// upper-bound estimates so we err on the side of more left padding rather than
// clipping.
const ASCII_CHAR_PX = 7;
const CJK_CHAR_PX = 14;

/**
 * Returns an approximate rendered pixel width for `text` at 12px sans-serif.
 * CJK / fullwidth characters count as ~2x the width of ASCII letters.
 *
 * This is intentionally a pure heuristic (no canvas / no DOM), so it works in
 * bun:test under happy-dom and gives a stable answer regardless of the host's
 * available fonts.
 */
export function estimateLabelWidth(text: string): number {
  if (!text) return 0;
  let width = 0;
  for (const ch of text) {
    // CJK Unified Ideographs (一-鿿), Hiragana/Katakana, fullwidth
    // forms, Hangul, CJK symbols & punctuation — all roughly 2x ASCII width.
    if (/[　-鿿가-힯＀-￯]/.test(ch)) {
      width += CJK_CHAR_PX;
    } else {
      width += ASCII_CHAR_PX;
    }
  }
  return width;
}

/**
 * Compute the left margin (in CSS px, pre-`rem()` scaling) needed so that the
 * longest label in `labels` fits in the SVG gutter without being clipped.
 *
 * `minMargin` is the previous hard-coded margin; we never go below it to keep
 * the layout consistent for short Indian-style account paths. `padding` is the
 * gap between the longest label and the chart body.
 */
export function computeLeftMargin(labels: string[], minMargin: number, padding = 20): number {
  if (!labels || labels.length === 0) return minMargin;
  let widest = 0;
  for (const label of labels) {
    const w = estimateLabelWidth(label);
    if (w > widest) widest = w;
  }
  return Math.max(minMargin, widest + padding);
}
