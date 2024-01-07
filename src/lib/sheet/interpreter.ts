import type { SyntaxNode } from "@lezer/common";
import * as Terms from "./parser.terms";
import type { EditorState } from "@codemirror/state";
import { BigNumber } from "bignumber.js";
import { asTransaction, type Posting, type SheetLineResult } from "$lib/utils";
import {
  buildFilter,
  buildAST as buildSearchAST,
  type TransactionPredicate
} from "$lib/search_query_editor";

const STACK_LIMIT = 1000;

export class Environment {
  scope: Record<string, any>;
  depth: number;
  postings: Posting[];

  constructor() {
    this.scope = {};
    this.depth = 0;
  }

  extend(scope: Record<string, any>): Environment {
    const env = new Environment();
    env.postings = this.postings;
    env.depth = this.depth + 1;
    if (this.depth > STACK_LIMIT) {
      throw new Error("Call stack overflow");
    }
    env.scope = { ...this.scope, ...scope };
    return env;
  }
}

abstract class AST {
  readonly id: number;
  constructor(readonly node: SyntaxNode) {
    this.id = node.type.id;
  }

  abstract evaluate(env: Environment): any;
}

class NumberAST extends AST {
  readonly value: BigNumber;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.value = new BigNumber(state.sliceDoc(node.from, node.to));
  }

  evaluate(): any {
    return this.value;
  }
}

class IdentifierAST extends AST {
  readonly name: string;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.name = state.sliceDoc(node.from, node.to);
  }

  evaluate(env: Environment): any {
    if (env.scope[this.name] === undefined) {
      throw new Error(`Undefined variable ${this.name}`);
    }
    return env.scope[this.name];
  }
}

class UnaryExpressionAST extends AST {
  readonly operator: string;
  readonly value: ExpressionAST;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.operator = state.sliceDoc(node.firstChild.from, node.firstChild.to);
    this.value = new ExpressionAST(node.lastChild, state);
  }

  evaluate(env: Environment): any {
    switch (this.operator) {
      case "-":
        return (this.value.evaluate(env) as BigNumber).negated();
      default:
        throw new Error("Unexpected operator");
    }
  }
}

class BinaryExpressionAST extends AST {
  readonly operator: string;
  readonly left: ExpressionAST;
  readonly right: ExpressionAST;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.left = new ExpressionAST(node.firstChild, state);
    this.operator = state.sliceDoc(
      node.firstChild.nextSibling.from,
      node.firstChild.nextSibling.to
    );
    this.right = new ExpressionAST(node.lastChild, state);
  }

  evaluate(env: Environment): any {
    switch (this.operator) {
      case "+":
        return (this.left.evaluate(env) as BigNumber).plus(this.right.evaluate(env));
      case "-":
        return (this.left.evaluate(env) as BigNumber).minus(this.right.evaluate(env));
      case "*":
        return (this.left.evaluate(env) as BigNumber).times(this.right.evaluate(env));
      case "/":
        return (this.left.evaluate(env) as BigNumber).div(this.right.evaluate(env));
      case "^":
        return (this.left.evaluate(env) as BigNumber).exponentiatedBy(this.right.evaluate(env));
      default:
        throw new Error("Unexpected operator");
    }
  }
}

class FunctionCallAST extends AST {
  readonly identifier: string;
  readonly arguments: ExpressionAST[];
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.identifier = state.sliceDoc(node.firstChild.from, node.firstChild.to);
    this.arguments = childrens(node.firstChild.nextSibling).map(
      (node) => new ExpressionAST(node, state)
    );
  }

  evaluate(env: Environment): any {
    const fun = env.scope[this.identifier];
    if (typeof fun !== "function") {
      throw new Error(`Undefined function ${this.identifier}`);
    }
    return fun(env, ...this.arguments.map((arg) => arg.evaluate(env)));
  }
}

class PostingsAST extends AST {
  readonly predicate: TransactionPredicate;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.predicate = buildFilter(buildSearchAST(state, node.lastChild.firstChild.nextSibling));
  }

  evaluate(env: Environment): any {
    return env.postings
      .map(asTransaction)
      .filter(this.predicate)
      .map((t) => t.postings[0]);
  }
}

class ExpressionAST extends AST {
  readonly value:
    | NumberAST
    | IdentifierAST
    | UnaryExpressionAST
    | BinaryExpressionAST
    | ExpressionAST
    | FunctionCallAST
    | PostingsAST;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    switch (node.firstChild.type.id) {
      case Terms.Literal:
        switch (node.firstChild.firstChild.type.id) {
          case Terms.Number:
            this.value = new NumberAST(node.firstChild, state);
            break;
          default:
            throw new Error("Unexpected node type");
        }
        break;
      case Terms.UnaryExpression:
        this.value = new UnaryExpressionAST(node.firstChild, state);
        break;
      case Terms.BinaryExpression:
        this.value = new BinaryExpressionAST(node.firstChild, state);
        break;

      case Terms.Grouping:
        this.value = new ExpressionAST(node.firstChild.firstChild, state);
        break;

      case Terms.Identifier:
        this.value = new IdentifierAST(node.firstChild, state);
        break;

      case Terms.FunctionCall:
        this.value = new FunctionCallAST(node.firstChild, state);
        break;

      case Terms.Postings:
        this.value = new PostingsAST(node.firstChild, state);
        break;

      default:
        throw new Error("Unexpected node type");
    }
  }

  evaluate(env: Environment): any {
    return this.value.evaluate(env);
  }
}

class AssignmentAST extends AST {
  readonly identifier: string;
  readonly value: ExpressionAST;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.identifier = state.sliceDoc(node.firstChild.from, node.firstChild.to);
    this.value = new ExpressionAST(node.lastChild, state);
  }

  evaluate(env: Environment): any {
    env.scope[this.identifier] = this.value.evaluate(env);
    return env.scope[this.identifier];
  }
}

class HeaderAST extends AST {
  readonly text: string;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.text = state.sliceDoc(node.from, node.to);
  }

  evaluate(): any {
    return this.text;
  }
}

class FunctionDefinitionAST extends AST {
  readonly identifier: string;
  readonly parameters: string[];
  readonly body: ExpressionAST;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.identifier = state.sliceDoc(node.firstChild.from, node.firstChild.to);
    this.parameters = childrens(node.firstChild.nextSibling).map((node) =>
      state.sliceDoc(node.from, node.to)
    );
    this.body = new ExpressionAST(node.lastChild, state);
  }

  evaluate(env: Environment): any {
    env.scope[this.identifier] = (env: Environment, ...args: any[]) => {
      const newEnv = env.extend({});
      for (let i = 0; i < args.length; i++) {
        newEnv.scope[this.parameters[i]] = args[i];
      }
      return this.body.evaluate(newEnv);
    };
    return null;
  }
}

class LineAST extends AST {
  readonly lineNumber: number;
  readonly valueId: number;
  readonly value: ExpressionAST | AssignmentAST | FunctionDefinitionAST | HeaderAST;
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    this.lineNumber = state.doc.lineAt(node.from).number;
    const child = node.firstChild;
    this.valueId = child.type.id;
    switch (child.type.id) {
      case Terms.Expression:
        this.value = new ExpressionAST(child, state);
        break;
      case Terms.Assignment:
        this.value = new AssignmentAST(child, state);
        break;
      case Terms.FunctionDefinition:
        this.value = new FunctionDefinitionAST(child, state);
        break;
      case Terms.Header:
        this.value = new HeaderAST(child, state);
        break;
      default:
        throw new Error("Unexpected node type");
    }
  }

  evaluate(env: Environment): Record<string, any> {
    let value = this.value.evaluate(env);
    if (value instanceof BigNumber) {
      value = value.toFixed(2);
    }
    switch (this.valueId) {
      case Terms.Assignment:
      case Terms.Expression:
        return { result: value?.toString() || "" };
      case Terms.FunctionDefinition:
        return { result: "" };
      case Terms.Header:
        return { result: value?.toString() || "", align: "left", bold: true, underline: true };
      default:
        throw new Error("Unexpected node type");
    }
  }
}

class SheetAST extends AST {
  readonly lines: LineAST[];
  constructor(node: SyntaxNode, state: EditorState) {
    super(node);
    const nodes = childrens(node);
    this.lines = [];
    for (const node of nodes) {
      try {
        this.lines.push(new LineAST(node, state));
      } catch (e) {
        console.log(e);
        break;
      }
    }
  }

  evaluate(env: Environment): SheetLineResult[] {
    const results: SheetLineResult[] = [];
    let lastLineNumber = 0;
    for (const line of this.lines) {
      while (line.lineNumber > lastLineNumber + 1) {
        results.push({ line: lastLineNumber + 1, error: false, result: "" });
        lastLineNumber++;
      }
      try {
        const resultObject = line.evaluate(env);
        results.push({ line: line.lineNumber, error: false, ...resultObject } as SheetLineResult);
        lastLineNumber++;
      } catch (e) {
        console.log(e);
        results.push({ line: line.lineNumber, error: true, result: e.message });
        break;
      }
    }
    return results;
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

export function buildAST(node: SyntaxNode, state: EditorState): SheetAST {
  return new SheetAST(node, state);
}
