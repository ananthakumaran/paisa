import { styleTags, tags as t } from "@lezer/highlight";

export const jsonHighlighting = styleTags({
  Quoted: t.string,
  UnQuoted: t.string,
  Number: t.number,
  DateValue: t.tagName,
  RegExp: t.regexp,
  "AND OR NOT": t.keyword,
  Account: t.typeName,
  Commodity: t.typeName,
  Filename: t.typeName,
  Amount: t.typeName,
  Total: t.typeName,
  Date: t.typeName,
  Payee: t.typeName,
  "( )": t.paren,
  "[ ]": t.squareBracket,
  "= =~ < > <= >=": t.operator
});
