import type { Transaction } from "$lib/utils";
import { queryExtension } from "./search/parser/query";
import { EditorView, minimalSetup } from "codemirror";
import { linter, type Diagnostic } from "@codemirror/lint";
import { placeholder } from "@codemirror/view";
import _ from "lodash";
import { writable } from "svelte/store";
import {
  CompletionContext,
  autocompletion,
  closeBrackets,
  completeFromList,
  ifIn
} from "@codemirror/autocomplete";
import { syntaxTree, bracketMatching } from "@codemirror/language";
import type { SyntaxNode } from "@lezer/common";
import * as Terms from "./search/parser/parser.terms";
import type { EditorState } from "@codemirror/state";
import dayjs from "dayjs";
import * as chrono from "chrono-node";

type StringAST = {
  type: "string";
  node: SyntaxNode;
  quoted?: boolean;
  value: string;
  id: number;
};

type NumberAST = {
  type: "number";
  node: SyntaxNode;
  value: number;
  id: number;
};

type RegExpAST = {
  type: "regexp";
  node: SyntaxNode;
  value: RegExp;
  id: number;
};

type DateRange = {
  start: dayjs.Dayjs;
  end: dayjs.Dayjs;
};

type DateValueAST = {
  type: "date";
  node: SyntaxNode;
  value: DateRange;
  id: number;
};

type PropertyAST = {
  type: "property";
  node: SyntaxNode;
  value: string;
  id: number;
};

type OperatorAST = {
  type: "operator";
  node: SyntaxNode;
  value: string;
  id: number;
};

type ConditionAST = {
  type: "condition";
  node: SyntaxNode;
  property: PropertyAST;
  operator: OperatorAST;
  value: ValueAST;
  id: number;
};

type BooleanBinaryAST = {
  type: "booleanBinary";
  node: SyntaxNode;
  left: ClauseAST;
  operator: "AND" | "OR";
  right: ClauseAST;
  id: number;
};

type BooleanUnaryAST = {
  type: "booleanUnary";
  node: SyntaxNode;
  operator: "NOT";
  right: ClauseAST;
  id: number;
};

interface ExpressionAST {
  type: "expression";
  node: SyntaxNode;
  clauses: ClauseAST[];
  id: number;
}

type BooleanConditionAST = BooleanBinaryAST | BooleanUnaryAST;

type ValueAST = StringAST | NumberAST | RegExpAST | DateValueAST;

type ClauseAST = ValueAST | ConditionAST | BooleanConditionAST | ExpressionAST;

interface QueryAST {
  type: "query";
  node: SyntaxNode;
  clauses: ClauseAST[];
  id: number;
}

interface QueryEditorEditorState {
  predicate: TransactionPredicate;
}

const initialEditorState: QueryEditorEditorState = {
  predicate: () => true
};

export const editorState = writable(initialEditorState);

const allowedCombinations: Record<string, Record<string, [number]>> = {
  [Terms.Account]: {
    "=": [Terms.String],
    "=~": [Terms.RegExp]
  },
  [Terms.Commodity]: {
    "=": [Terms.String],
    "=~": [Terms.RegExp]
  },
  [Terms.Amount]: {
    "=": [Terms.Number],
    ">": [Terms.Number],
    "<": [Terms.Number],
    ">=": [Terms.Number],
    "<=": [Terms.Number]
  },
  [Terms.Total]: {
    "=": [Terms.Number],
    ">": [Terms.Number],
    "<": [Terms.Number],
    ">=": [Terms.Number],
    "<=": [Terms.Number]
  },
  [Terms.Date]: {
    "=": [Terms.DateValue],
    ">": [Terms.DateValue],
    "<": [Terms.DateValue],
    ">=": [Terms.DateValue],
    "<=": [Terms.DateValue]
  },
  [Terms.Payee]: {
    "=": [Terms.String],
    "=~": [Terms.RegExp]
  },
  [Terms.Filename]: {
    "=": [Terms.String],
    "=~": [Terms.RegExp]
  },
  [Terms.Note]: {
    "=": [Terms.String],
    "=~": [Terms.RegExp]
  }
};

function lint(editor: EditorView): Diagnostic[] {
  const diagnostics: Diagnostic[] = [];
  let hasErrors = false;

  syntaxTree(editor.state)
    .cursor()
    .iterate((node) => {
      if (node.type.isError) {
        hasErrors = true;
        diagnostics.push({
          from: node.from,
          to: node.to,
          severity: "error",
          message: "Invalid syntax"
        });
      }
    });

  if (!hasErrors) {
    const ast = buildAST(editor);

    const conditions = ast.clauses.flatMap(collectConditionASTs);
    for (const condition of conditions) {
      const allowed: number[] =
        allowedCombinations[condition.property.id.toString()][condition.operator.value] || [];
      if (!allowed.includes(condition.value.id)) {
        hasErrors = true;
        diagnostics.push({
          from: condition.node.from,
          to: condition.node.to,
          severity: "error",
          message: `${condition.property.value} cannot be used with ${condition.operator.value} and ${condition.value.type}`
        });
      }
    }

    const dateValues = ast.clauses.flatMap(collectDateValueASTs);
    for (const dateValue of dateValues) {
      if (!dateValue.value) {
        hasErrors = true;
        diagnostics.push({
          from: dateValue.node.from,
          to: dateValue.node.to,
          severity: "error",
          message: `Invalid date`
        });
      }
    }

    if (!hasErrors) {
      editorState.update((current) => _.assign({}, current, { predicate: buildFilter(ast) }));
    }
  }

  return diagnostics;
}

function collectConditionASTs(ast: ClauseAST): ConditionAST[] {
  switch (ast.type) {
    case "condition":
      return [ast];
    case "expression":
      return ast.clauses.flatMap(collectConditionASTs);
    case "booleanBinary":
      return [...collectConditionASTs(ast.left), ...collectConditionASTs(ast.right)];
    case "booleanUnary":
      return collectConditionASTs(ast.right);
  }
  return [];
}

function collectDateValueASTs(ast: ClauseAST): DateValueAST[] {
  switch (ast.type) {
    case "date":
      return [ast];
    case "expression":
      return ast.clauses.flatMap(collectDateValueASTs);
    case "condition":
      return collectDateValueASTs(ast.value);
    case "booleanBinary":
      return [...collectDateValueASTs(ast.left), ...collectDateValueASTs(ast.right)];
    case "booleanUnary":
      return collectDateValueASTs(ast.right);
  }
  return [];
}

function buildAST(editor: EditorView): QueryAST {
  return constructQueryAST(editor.state, syntaxTree(editor.state).topNode);
}

type TransactionPredicate = (transaction: Transaction) => boolean;

function buildFilter(ast: QueryAST): TransactionPredicate {
  return andFilter(...ast.clauses.map((clause) => buildFilterFromClauseAST(clause)));
}

function buildFilterFromConditionAST(ast: ConditionAST): TransactionPredicate {
  return conditionFilter(ast.property.id, ast.operator.value, ast.value.value);
}

function buildFilterFromBooleanConditionAST(ast: BooleanConditionAST): TransactionPredicate {
  switch (ast.type) {
    case "booleanBinary":
      return buildFilterFromBooleanBinaryAST(ast);
    case "booleanUnary":
      return buildFilterFromBooleanUnaryAST(ast);
  }

  return assertUnreachable(ast);
}

function buildFilterFromBooleanBinaryAST(ast: BooleanBinaryAST): TransactionPredicate {
  switch (ast.operator) {
    case "AND":
      return andFilter(buildFilterFromClauseAST(ast.left), buildFilterFromClauseAST(ast.right));
    case "OR":
      return orFilter(buildFilterFromClauseAST(ast.left), buildFilterFromClauseAST(ast.right));
  }

  return assertUnreachable(ast.operator);
}

function buildFilterFromBooleanUnaryAST(ast: BooleanUnaryAST): TransactionPredicate {
  switch (ast.operator) {
    case "NOT":
      return notFilter(buildFilterFromClauseAST(ast.right));
  }

  return assertUnreachable(ast.operator);
}

function buildFilterFromClauseAST(ast: ClauseAST): TransactionPredicate {
  switch (ast.type) {
    case "string":
    case "number":
    case "regexp":
    case "date":
      return buildFilterFromValueAST(ast);

    case "expression":
      return buildFilterFromExpressionAST(ast);
    case "condition":
      return buildFilterFromConditionAST(ast);
    case "booleanBinary":
    case "booleanUnary":
      return buildFilterFromBooleanConditionAST(ast);
  }

  return assertUnreachable(ast);
}

function buildFilterFromExpressionAST(ast: ExpressionAST): TransactionPredicate {
  return andFilter(...ast.clauses.map((clause) => buildFilterFromClauseAST(clause)));
}

function buildFilterFromValueAST(ast: ValueAST): TransactionPredicate {
  switch (ast.type) {
    case "string":
      return conditionFilter(Terms.Account, "=", ast.value);
    case "number":
      return conditionFilter(Terms.Amount, "=", ast.value);
    case "regexp":
      return conditionFilter(Terms.Account, "=~", ast.value);
    case "date":
      return conditionFilter(Terms.Date, "=", ast.value);
  }

  return assertUnreachable(ast);
}

function andFilter(...filters: TransactionPredicate[]): TransactionPredicate {
  return (transaction: Transaction) => {
    return _.every(filters, (filter) => filter(transaction));
  };
}

function orFilter(...filters: TransactionPredicate[]): TransactionPredicate {
  return (transaction: Transaction) => {
    return _.some(filters, (filter) => filter(transaction));
  };
}

function notFilter(...filters: TransactionPredicate[]): TransactionPredicate {
  return (transaction: Transaction) => {
    return !_.some(filters, (filter) => filter(transaction));
  };
}

function conditionFilter(
  property: number,
  operator: string,
  expected: string | number | RegExp | { start: dayjs.Dayjs; end: dayjs.Dayjs }
) {
  return (transaction: Transaction) => {
    const operatorFn = getOperator(operator);
    return _.some(getProperty(transaction, property), (actual) => operatorFn(actual, expected));
  };
}

function getOperator(operator: string): (actual: any, expected: any) => boolean {
  return (actual: any, expected: any) => {
    switch (operator) {
      case "=":
        if (typeof actual === "string") {
          return _.includes(actual.toLowerCase(), expected.toLowerCase());
        }

        if (typeof actual === "number") {
          return actual === expected;
        }

        if (typeof actual === "object") {
          const expectedDate = expected as { start: dayjs.Dayjs; end: dayjs.Dayjs };
          return (
            actual.isSameOrAfter(expectedDate.start) && actual.isSameOrBefore(expectedDate.end)
          );
        }

        return actual === expected;

      case "=~":
        return (expected as RegExp).test(actual);

      case ">":
        if (typeof actual === "object") {
          return actual.isAfter(expected.start);
        }

        return actual > expected;

      case "<":
        if (typeof actual === "object") {
          return actual.isBefore(expected.end);
        }

        return actual < expected;

      case ">=":
        if (typeof actual === "object") {
          return actual.isSameOrAfter(expected.start);
        }

        return actual >= expected;

      case "<=":
        if (typeof actual === "object") {
          return actual.isSameOrBefore(expected.end);
        }
        return actual <= expected;
    }
  };
}

function getProperty(
  transaction: Transaction,
  property: number
): Array<string | number | dayjs.Dayjs> {
  switch (property) {
    case Terms.Date:
      return [transaction.date];
    case Terms.Payee:
      return [transaction.payee];
    case Terms.Account:
      return transaction.postings.map((posting) => posting.account);
    case Terms.Commodity:
      return transaction.postings.map((posting) => posting.commodity);
    case Terms.Amount:
      return transaction.postings.map((posting) => posting.amount);
    case Terms.Total:
      return [
        _.sumBy(transaction.postings, (posting) => (posting.amount > 0 ? posting.amount : 0))
      ];
    case Terms.Filename:
      return [transaction.fileName];
    case Terms.Note:
      return [transaction.note].concat(transaction.postings.map((posting) => posting.note));
  }
}

function constructQueryAST(state: EditorState, node: SyntaxNode): QueryAST {
  return {
    type: "query",
    node,
    clauses: childrens(node).map((child) => constructClauseAST(state, child)),
    id: node.type.id
  };
}

function constructClauseAST(state: EditorState, node: SyntaxNode): ClauseAST {
  let cs: SyntaxNode[];
  switch (node.type.id) {
    case Terms.Value:
      return constructValueAST(state, node.firstChild);
    case Terms.Clause:
      return constructClauseAST(state, node.firstChild);

    case Terms.Expression:
      return {
        type: "expression",
        node,
        clauses: childrens(node).map((child) => constructClauseAST(state, child)),
        id: node.type.id
      };
    case Terms.Condition:
      return {
        type: "condition",
        node,
        property: constructPropertyAST(state, node.getChild(Terms.Property).firstChild),
        operator: constructOperatorAST(state, node.getChild(Terms.Operator).firstChild),
        value: constructValueAST(state, node.getChild(Terms.Value).firstChild),
        id: node.type.id
      };

    case Terms.BooleanCondition:
      cs = childrens(node);
      if (cs.length === 3) {
        return {
          type: "booleanBinary",
          node,
          left: constructClauseAST(state, cs[0]),
          operator: state.sliceDoc(cs[1].from, cs[1].to) as "AND" | "OR",
          right: constructClauseAST(state, cs[2]),
          id: node.type.id
        };
      }

      if (cs.length === 2) {
        return {
          type: "booleanUnary",
          node,
          operator: state.sliceDoc(cs[0].from, cs[0].to) as "NOT",
          right: constructClauseAST(state, cs[1]),
          id: node.type.id
        };
      }
  }
}

function constructPropertyAST(state: EditorState, node: SyntaxNode): PropertyAST {
  return {
    type: "property",
    node,
    value: state.sliceDoc(node.from, node.to),
    id: node.type.id
  };
}

function constructOperatorAST(state: EditorState, node: SyntaxNode): OperatorAST {
  return {
    type: "operator",
    node,
    value: state.sliceDoc(node.from, node.to),
    id: node.type.id
  };
}

function constructValueAST(state: EditorState, node: SyntaxNode): ValueAST {
  if (node.type.is(Terms.String)) {
    const quoted = node.firstChild.type.is(Terms.Quoted);
    let value: string;
    if (node.firstChild.type.is(Terms.Quoted)) {
      value = state.sliceDoc(node.from + 1, node.to - 1);
    } else {
      value = state.sliceDoc(node.from, node.to);
    }

    return {
      type: "string",
      node,
      quoted,
      value,
      id: node.type.id
    };
  }

  if (node.type.is(Terms.Number)) {
    return {
      type: "number",
      node,
      value: Number(state.sliceDoc(node.from, node.to)),
      id: node.type.id
    };
  }

  if (node.type.is(Terms.RegExp)) {
    const all = state.sliceDoc(node.from, node.to);
    const result = all.match(/\/(.*)\/(.*)/);
    return {
      type: "regexp",
      node,
      value: new RegExp(result[1], result[2]),
      id: node.type.id
    };
  }

  if (node.type.is(Terms.DateValue)) {
    const value = state.sliceDoc(node.from + 1, node.to - 1);
    return {
      type: "date",
      node,
      value: parseDate(value),
      id: node.type.id
    };
  }
}

export function parseDate(value: string, reference = new Date()): DateRange {
  if (/^[0-9]{4}$/.test(value)) {
    const date = dayjs(value, "YYYY");
    return {
      start: date.startOf("year"),
      end: date.endOf("year")
    };
  }

  if (/^[0-9]{4}[/-][0-9]{2}$/.test(value)) {
    const date = dayjs(value.replaceAll(/[/-]/g, " "), "YYYY MM");
    return {
      start: date.startOf("month"),
      end: date.endOf("month")
    };
  }

  const results = chrono.parse(value, reference);
  if (results.length === 0) {
    return null;
  }

  const result = results[0];
  const start = adjustInterval(value, result.start, "start");
  const end = adjustInterval(value, result.end || result.start, "end");

  return {
    start: start,
    end: end
  };
}

function adjustInterval(text: string, parsed: chrono.ParsedComponents, direction: "start" | "end") {
  let interval: dayjs.OpUnitType = "day";
  if (parsed.isCertain("day")) {
    interval = "day";
  } else if (/week/i.test(text) || parsed.isCertain("weekday")) {
    interval = "week";
  } else if (parsed.isCertain("month")) {
    interval = "month";
  } else if (parsed.isCertain("year")) {
    interval = "year";
  }

  const date = dayjs(parsed.date());

  switch (direction) {
    case "start":
      return date.startOf(interval);
    case "end":
      return date.endOf(interval);
  }
}

function childrens(node: SyntaxNode): SyntaxNode[] {
  if (!node) {
    return [];
  }

  const cur = node.cursor();
  const result: SyntaxNode[] = [];
  if (!cur.firstChild()) {
    return result;
  }

  do {
    result.push(cur.node);
  } while (cur.nextSibling());
  return result;
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
function assertUnreachable(_x: never): never {
  throw new Error("Didn't expect to get here");
}

export function createEditor(
  content: string,
  dom: Element,
  autocomplete: Record<string, string[]>
) {
  const autocompletions: Record<string, string[]> = {
    UnQuoted: [
      "account",
      "commodity",
      "amount",
      "date",
      "payee",
      "filename",
      "note",
      "total",
      "AND",
      "OR",
      "NOT"
    ]
  };

  const completions = _.chain(autocomplete)
    .mapValues((values) => completeFromList(values))
    .value();

  editorState.set(initialEditorState);

  return new EditorView({
    extensions: [
      minimalSetup,
      bracketMatching(),
      closeBrackets(),
      EditorView.theme({
        "&": {
          fontSize: "14px"
        }
      }),
      EditorView.contentAttributes.of({ "data-enable-grammarly": "false" }),
      queryExtension(),
      linter(lint),
      autocompletion({
        override: [
          (context: CompletionContext) => {
            for (const [key, completionSource] of Object.entries(completions)) {
              if (context.matchBefore(new RegExp(`${key}\\s*=[~]?\\s*[^ ]*$`))) {
                return completionSource(context);
              }
            }

            return null;
          },
          ..._.map(autocompletions, (options, node) => ifIn([node], completeFromList(options)))
        ]
      }),
      placeholder(
        "(account = Expenses:Utilities OR payee =~ /Pacific Gas & Electric/i) AND [2023-04]"
      )
    ],
    doc: content,
    parent: dom
  });
}
