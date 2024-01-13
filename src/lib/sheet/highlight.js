import { styleTags, tags as t } from "@lezer/highlight";

export const sheetHighlighting = styleTags({
  String: t.string,
  Number: t.number,
  Percent: t.number,
  Header: t.heading,
  Comment: t.comment,
  "FunctionDefinition/Identifier": t.function(t.variableName),
  "FunctionCall/Identifier": t.function(t.variableName),
  "{ }": t.operator,
  UnaryOperator: t.operator,
  BinaryOperator: t.operator,
  "BinaryOperator/AND": t.keyword,
  "BinaryOperator/OR": t.keyword,
  AssignmentOperator: t.operator
});
