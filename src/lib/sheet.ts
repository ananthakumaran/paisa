import { closeBrackets } from "@codemirror/autocomplete";
import { history, redoDepth, undoDepth } from "@codemirror/commands";
import { bracketMatching, syntaxTree } from "@codemirror/language";
import { lintGutter, linter, type Diagnostic } from "@codemirror/lint";
import { keymap, type KeyBinding } from "@codemirror/view";
import { EditorView } from "codemirror";
import _ from "lodash";
import { sheetEditorState } from "../store";
import { basicSetup } from "./editor/base";
import { sheetExtension } from "./sheet/language";
import { schedulePlugin } from "./transaction_tag";
export { sheetEditorState } from "../store";
import { functions } from "./sheet/functions";

import { Environment, buildAST } from "./sheet/interpreter";
import type { Posting } from "./utils";

function lint(env: Environment) {
  return function (editor: EditorView): Diagnostic[] {
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

    sheetEditorState.update((current) => {
      if (!current.pendingEval) {
        return current;
      }

      const startTime = performance.now();
      let results = current.results;
      try {
        const ast = buildAST(tree.topNode, editor.state);
        results = ast.evaluate(env);
      } catch (e) {
        console.log(e);
        // ignore
      }
      const endTime = performance.now();

      return _.assign({}, current, {
        pendingEval: false,
        evalDuration: endTime - startTime,
        results
      });
    });

    return diagnostics;
  };
}

export function createEditor(
  content: string,
  dom: Element,
  postings: Posting[],
  opts: {
    keybindings?: readonly KeyBinding[];
  }
) {
  const env = new Environment();
  env.scope = functions;
  env.postings = postings;

  let firstLoad = true;

  return new EditorView({
    extensions: [
      keymap.of(opts.keybindings || []),
      basicSetup,
      bracketMatching(),
      closeBrackets(),
      EditorView.contentAttributes.of({ "data-enable-grammarly": "false" }),
      sheetExtension(),
      linter(lint(env), {
        delay: 300,
        needsRefresh: () => {
          if (firstLoad) {
            firstLoad = false;
            return true;
          }

          return false;
        }
      }),
      lintGutter(),
      history(),
      EditorView.updateListener.of((viewUpdate) => {
        const doc = viewUpdate.state.doc.toString();
        const currentLine = viewUpdate.state.doc.lineAt(viewUpdate.state.selection.main.head);
        sheetEditorState.update((current) => {
          let pendingEval = current.pendingEval;
          if (current.doc !== doc) {
            pendingEval = true;
          }

          return _.assign({}, current, {
            pendingEval,
            doc,
            currentLine: currentLine.number,
            hasUnsavedChanges: current.hasUnsavedChanges || viewUpdate.docChanged,
            undoDepth: undoDepth(viewUpdate.state),
            redoDepth: redoDepth(viewUpdate.state)
          });
        });
      }),
      schedulePlugin
    ],
    doc: content,
    parent: dom
  });
}
