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

abstract class AST {
  readonly id: number;
  constructor(readonly node: SyntaxNode) {
    this.id = node.type.id;
  }

  abstract validate(): Diagnostic[];

  abstract evaluate(): TransactionPredicate;

  get type(): string {
    return this.node.type.name;
  }
}

class StringAST extends AST {
  readonly quoted: boolean;
  readonly value: string;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.quoted = node.firstChild.type.is(Terms.Quoted);
    if (node.firstChild.type.is(Terms.Quoted)) {
      this.value = state.sliceDoc(node.from + 1, node.to - 1);
    } else {
      this.value = state.sliceDoc(node.from, node.to);
    }
  }

  validate(): Diagnostic[] {
    return [];
  }

  evaluate(): TransactionPredicate {
    return conditionFilter(Terms.Account, "=", this.value);
  }
}

class NumberAST extends AST {
  readonly value: number;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.value = Number(state.sliceDoc(node.from, node.to));
  }

  validate(): Diagnostic[] {
    return [];
  }

  evaluate(): TransactionPredicate {
    return conditionFilter(Terms.Amount, "=", this.value);
  }
}

class RegExpAST extends AST {
  readonly value: RegExp;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    const all = state.sliceDoc(node.from, node.to);
    const result = all.match(/\/(.*)\/(.*)/);
    this.value = new RegExp(result[1], result[2]);
  }

  validate(): Diagnostic[] {
    return [];
  }

  evaluate(): TransactionPredicate {
    return conditionFilter(Terms.Account, "=~", this.value);
  }
}

type DateRange = {
  start: dayjs.Dayjs;
  end: dayjs.Dayjs;
};

class DateValueAST extends AST {
  readonly value: DateRange;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    const value = state.sliceDoc(node.from + 1, node.to - 1);
    this.value = parseDate(value);
  }

  validate(): Diagnostic[] {
    if (!this.value) {
      return [
        {
          from: this.node.from,
          to: this.node.to,
          severity: "error",
          message: `Invalid date`
        }
      ];
    }
    return [];
  }

  evaluate(): TransactionPredicate {
    return conditionFilter(Terms.Date, "=", this.value);
  }
}

class PropertyAST extends AST {
  readonly value: string;
  readonly childId: number;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    const child = node.firstChild;
    this.childId = child.type.id;
    this.value = state.sliceDoc(node.from, node.to);
  }

  validate(): Diagnostic[] {
    return [];
  }

  evaluate(): TransactionPredicate {
    throw new Error("PropertyAST.evaluate() should never be called");
  }
}

class OperatorAST extends AST {
  readonly value: string;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.value = state.sliceDoc(node.from, node.to);
  }

  validate(): Diagnostic[] {
    return [];
  }

  evaluate(): TransactionPredicate {
    throw new Error("OperatorAST.evaluate() should never be called");
  }
}

class ConditionAST extends AST {
  readonly property: PropertyAST;
  readonly operator: OperatorAST;
  readonly value: ValueAST;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    const [property, operator, value] = childrens(node);
    this.property = new PropertyAST(property, state);
    this.operator = new OperatorAST(operator, state);
    this.value = new ValueAST(value, state);
  }

  validate(): Diagnostic[] {
    const allowed: number[] =
      allowedCombinations[this.property.childId.toString()][this.operator.value] || [];
    if (!allowed.includes(this.value.value.id)) {
      return [
        {
          from: this.node.from,
          to: this.node.to,
          severity: "error",
          message: `${this.property.value} cannot be used with ${this.operator.value} and ${this.value.value.type}`
        }
      ];
    }

    return [];
  }

  evaluate(): TransactionPredicate {
    return conditionFilter(this.property.childId, this.operator.value, this.value.value.value);
  }
}

class ValueAST extends AST {
  readonly value: StringAST | NumberAST | RegExpAST | DateValueAST;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);

    const child = node.firstChild;
    switch (child.type.id) {
      case Terms.String:
        this.value = new StringAST(child, state);
        break;
      case Terms.Number:
        this.value = new NumberAST(child, state);
        break;
      case Terms.RegExp:
        this.value = new RegExpAST(child, state);
        break;
      case Terms.DateValue:
        this.value = new DateValueAST(child, state);
        break;
    }
  }

  validate(): Diagnostic[] {
    return this.value.validate();
  }

  evaluate(): TransactionPredicate {
    return this.value.evaluate();
  }
}

class BooleanBinaryAST extends AST {
  readonly left: AST;
  readonly operator: "AND" | "OR";
  readonly right: AST;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    const [left, operator, right] = childrens(node);
    this.left = new ClauseAST(left, state);
    this.operator = state.sliceDoc(operator.from, operator.to) as "AND" | "OR";
    this.right = new ClauseAST(right, state);
  }

  validate(): Diagnostic[] {
    return this.left.validate().concat(this.right.validate());
  }

  evaluate(): TransactionPredicate {
    switch (this.operator) {
      case "AND":
        return andFilter(this.left.evaluate(), this.right.evaluate());
      case "OR":
        return orFilter(this.left.evaluate(), this.right.evaluate());
    }
  }
}

class BooleanUnaryAST extends AST {
  readonly operator: "NOT";
  readonly right: AST;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    const [operator, right] = childrens(node);
    this.operator = state.sliceDoc(operator.from, operator.to) as "NOT";
    this.right = new ClauseAST(right, state);
  }

  validate(): Diagnostic[] {
    return this.right.validate();
  }

  evaluate(): TransactionPredicate {
    switch (this.operator) {
      case "NOT":
        return notFilter(this.right.evaluate());
    }
  }
}

class BooleanConditionAST extends AST {
  readonly value: BooleanBinaryAST | BooleanUnaryAST;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);

    const cs = childrens(node);
    if (cs.length === 3) {
      this.value = new BooleanBinaryAST(node, state);
    }

    if (cs.length === 2) {
      this.value = new BooleanUnaryAST(node, state);
    }
  }

  validate(): Diagnostic[] {
    return this.value.validate();
  }

  evaluate(): TransactionPredicate {
    return this.value.evaluate();
  }
}

class ClauseAST extends AST {
  readonly value: ValueAST | ConditionAST | BooleanConditionAST | ExpressionAST;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);

    const child = node.firstChild;
    switch (child.type.id) {
      case Terms.Expression:
        this.value = new ExpressionAST(child, state);
        break;
      case Terms.Condition:
        this.value = new ConditionAST(child, state);
        break;
      case Terms.BooleanCondition:
        this.value = new BooleanConditionAST(child, state);
        break;
      case Terms.Value:
        this.value = new ValueAST(child, state);
        break;
    }
  }

  validate(): Diagnostic[] {
    return this.value.validate();
  }

  evaluate(): TransactionPredicate {
    return this.value.evaluate();
  }
}

class ExpressionAST extends AST {
  readonly clauses: AST[];
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.clauses = childrens(node).map((child) => new ClauseAST(child, state));
  }

  validate(): Diagnostic[] {
    return this.clauses.flatMap((clause) => clause.validate());
  }

  evaluate(): TransactionPredicate {
    return andFilter(...this.clauses.map((clause) => clause.evaluate()));
  }
}

class QueryAST extends AST {
  readonly clauses: AST[];
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.clauses = childrens(node).map((child) => new ClauseAST(child, state));
  }

  validate(): Diagnostic[] {
    return this.clauses.flatMap((clause) => clause.validate());
  }

  evaluate(): TransactionPredicate {
    return andFilter(...this.clauses.map((clause) => clause.evaluate()));
  }
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

  const tree = syntaxTree(editor.state);

  tree.cursor().iterate((node) => {
    if (node.type.isError) {
      diagnostics.push({
        from: node.from,
        to: node.to,
        severity: "error",
        message: "Invalid syntax"
      });
    }
  });

  if (diagnostics.length === 0) {
    const ast = buildAST(editor.state, tree.topNode);
    diagnostics.push(...ast.validate());

    if (diagnostics.length === 0) {
      editorState.update((current) => _.assign({}, current, { predicate: ast.evaluate() }));
    }
  }

  return diagnostics;
}

export function buildAST(state: EditorState, node: SyntaxNode): QueryAST {
  return new QueryAST(node, state);
}

export type TransactionPredicate = (transaction: Transaction) => boolean;

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
