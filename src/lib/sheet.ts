import { closeBrackets } from "@codemirror/autocomplete";
import { keymap, type KeyBinding } from "@codemirror/view";
import { history, redoDepth, undoDepth } from "@codemirror/commands";
import {
  HighlightStyle,
  bracketMatching,
  syntaxHighlighting,
  defaultHighlightStyle,
  syntaxTree
} from "@codemirror/language";
import { lintGutter, linter, type Diagnostic } from "@codemirror/lint";
import { tags } from "@lezer/highlight";
import { EditorView } from "codemirror";
import _ from "lodash";
export { sheetEditorState } from "../store";
import { sheetEditorState } from "../store";
import { basicSetup } from "./editor/base";
import { sheetExtension } from "./sheet/language";
import { schedulePlugin } from "./transaction_tag";

import { buildAST } from "./sheet/interpreter";

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

  return diagnostics;
}

export function createEditor(
  content: string,
  dom: Element,
  opts: {
    keybindings?: readonly KeyBinding[];
  }
) {
  const highlightStyle = HighlightStyle.define(
    defaultHighlightStyle.specs.concat([
      { tag: tags.function(tags.variableName), color: "hsl(229, 53%, 53%)" },
      { tag: tags.number, color: "hsl(229, 53%, 53%)", fontWeight: "bold" }
    ])
  );

  return new EditorView({
    extensions: [
      keymap.of(opts.keybindings || []),
      basicSetup,
      syntaxHighlighting(highlightStyle),
      bracketMatching(),
      closeBrackets(),
      EditorView.contentAttributes.of({ "data-enable-grammarly": "false" }),
      sheetExtension(),
      linter(lint),
      lintGutter(),
      history(),
      EditorView.updateListener.of((viewUpdate) => {
        const doc = viewUpdate.state.doc.toString();
        const currentLine = viewUpdate.state.doc.lineAt(viewUpdate.state.selection.main.head);
        sheetEditorState.update((current) => {
          let results = current.results;
          if (current.doc !== doc) {
            const tree = syntaxTree(viewUpdate.state);
            try {
              const ast = buildAST(tree.topNode, viewUpdate.state);
              results = ast.evaluate();
            } catch (e) {
              // ignore
            }
          }

          return _.assign({}, current, {
            results: results,
            doc: doc,
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
