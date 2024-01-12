import type { LedgerFileError } from "$lib/utils";
import { handlebars } from "$lib/handlebars_parser";
import { StreamLanguage, syntaxHighlighting } from "@codemirror/language";
import { keymap } from "@codemirror/view";
import { basicSetup, EditorView } from "codemirror";
import { insertTab, history, undoDepth, redoDepth } from "@codemirror/commands";
import { linter, lintGutter, type Diagnostic } from "@codemirror/lint";
import _ from "lodash";
import { writable } from "svelte/store";
import { autocompletion, completeFromList, ifIn } from "@codemirror/autocomplete";
import Handlebars from "handlebars";
import { classHighlighter } from "@lezer/highlight";

interface EditorState {
  hasUnsavedChanges: boolean;
  undoDepth: number;
  redoDepth: number;
  errors: LedgerFileError[];
  template: HandlebarsTemplateDelegate;
}

const initialEditorState: EditorState = {
  hasUnsavedChanges: false,
  undoDepth: 0,
  redoDepth: 0,
  errors: [],
  template: null
};

export const editorState = writable(initialEditorState);

function lint(editor: EditorView): Diagnostic[] {
  const doc = editor.state.doc;
  try {
    Handlebars.parse(doc.toString());
    const compiled = Handlebars.compile(doc.toString(), { noEscape: true });
    editorState.update((current) => _.assign({}, current, { template: compiled }));
  } catch (e) {
    const lines = e.message.split("\n");
    const match = lines[0].match(/Parse error on line (\d+):/);
    if (match != null) {
      const line = doc.line(parseInt(match[1], 10));
      return [
        {
          message: lines[3],
          severity: "error",
          from: line.from,
          to: line.to
        }
      ];
    }
  }
  return [];
}

export function createEditor(content: string, dom: Element) {
  const autocompletions: Record<string, string[]> = {};

  editorState.set(initialEditorState);

  return new EditorView({
    extensions: [
      keymap.of([{ key: "Tab", run: insertTab }]),
      basicSetup,
      syntaxHighlighting(classHighlighter),
      EditorView.contentAttributes.of({ "data-enable-grammarly": "false" }),
      StreamLanguage.define(handlebars),
      lintGutter(),
      linter(lint),
      history(),
      autocompletion({
        override: _.map(autocompletions, (options, node) => ifIn([node], completeFromList(options)))
      }),
      EditorView.updateListener.of((viewUpdate) => {
        editorState.update((current) =>
          _.assign({}, current, {
            hasUnsavedChanges: current.hasUnsavedChanges || viewUpdate.docChanged,
            undoDepth: undoDepth(viewUpdate.state),
            redoDepth: redoDepth(viewUpdate.state)
          })
        );
      })
    ],
    doc: content,
    parent: dom
  });
}

export function updateContent(editor: EditorView, content: string) {
  editor.dispatch(
    editor.state.update({ changes: { from: 0, to: editor.state.doc.length, insert: content } })
  );
}
