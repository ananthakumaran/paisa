import { styleTags, tags as t } from "@lezer/highlight";

export const queryHighlighting = styleTags({
  Quoted: t.string,
  UnQuoted: t.string,
  Number: t.number,
  DateValue: t.tagName,
  RegExp: t.regexp,
  "AND OR NOT": t.keyword,
  Account: t.special(t.variableName),
  Commodity: t.special(t.variableName),
  Filename: t.special(t.variableName),
  Amount: t.special(t.variableName),
  Total: t.special(t.variableName),
  Date: t.special(t.variableName),
  Payee: t.special(t.variableName),
  "( )": t.paren,
  "[ ]": t.squareBracket,
  "= =~ < > <= >=": t.operator
});
