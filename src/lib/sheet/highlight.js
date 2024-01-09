import { styleTags, tags as t } from "@lezer/highlight";

export const sheetHighlighting = styleTags({
  String: t.string,
  Number: t.number,
  Header: t.heading,
  "FunctionDefinition/Identifier": t.function(t.variableName),
  "FunctionCall/Identifier": t.function(t.variableName),
  Postings: t.special(t.variableName),
  "{ }": t.operator,
  UnaryOperator: t.operator,
  BinaryOperator: t.operator,
  AssignmentOperator: t.operator
});
