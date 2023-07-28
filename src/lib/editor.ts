import { ajax, type LedgerFileError } from "$lib/utils";
import { ledger } from "$lib/parser";
import { StreamLanguage } from "@codemirror/language";
import { keymap } from "@codemirror/view";
import { EditorState as State } from "@codemirror/state";
import { basicSetup, EditorView } from "codemirror";
import { insertTab, history, undoDepth, redoDepth } from "@codemirror/commands";
import { linter, lintGutter, type Diagnostic } from "@codemirror/lint";
import _ from "lodash";
import { writable } from "svelte/store";
import { autocompletion, completeFromList, ifIn } from "@codemirror/autocomplete";

interface EditorState {
  hasUnsavedChanges: boolean;
  undoDepth: number;
  redoDepth: number;
  errors: LedgerFileError[];
  output: string;
}

const initialEditorState: EditorState = {
  hasUnsavedChanges: false,
  undoDepth: 0,
  redoDepth: 0,
  errors: [],
  output: ""
};

export const editorState = writable(initialEditorState);

async function lint(editor: EditorView): Promise<Diagnostic[]> {
  const doc = editor.state.doc;
  const response = await ajax("/api/editor/validate", {
    method: "POST",
    body: JSON.stringify({ name: "", content: editor.state.doc.toString() })
  });

  editorState.update((current) =>
    _.assign({}, current, { errors: response.errors, output: response.output })
  );

  return _.map(response.errors, (error) => {
    const lineFrom = doc.line(error.line_from);
    const lineTo = doc.line(error.line_to);
    return {
      message: error.message,
      severity: "error",
      from: lineFrom.from,
      to: lineTo.to
    };
  });
}

export function createEditor(
  content: string,
  dom: Element,
  opts: {
    autocompletions?: Record<string, string[]>;
    readonly?: boolean;
  }
) {
  editorState.set(initialEditorState);

  return new EditorView({
    extensions: [
      keymap.of([{ key: "Tab", run: insertTab }]),
      basicSetup,
      State.readOnly.of(!!opts.readonly),
      EditorView.theme({
        "&": {
          fontSize: "12px"
        }
      }),
      EditorView.contentAttributes.of({ "data-enable-grammarly": "false" }),
      StreamLanguage.define(ledger),
      lintGutter(),
      linter(lint),
      history(),
      autocompletion({
        override: _.map(opts.autocompletions || [], (options: string[], node) =>
          ifIn([node], completeFromList(options))
        )
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

export function moveToEnd(editor: EditorView) {
  editor.dispatch(
    editor.state.update({
      effects: EditorView.scrollIntoView(editor.state.doc.length, { y: "end" })
    })
  );
}

export function moveToLine(editor: EditorView, lineNumber: number) {
  const line = editor.state.doc.line(lineNumber);
  editor.dispatch(
    editor.state.update({
      effects: EditorView.scrollIntoView(line.from, { y: "center" })
    })
  );
}

export function updateContent(editor: EditorView, content: string) {
  editor.dispatch(
    editor.state.update({ changes: { from: 0, to: editor.state.doc.length, insert: content } })
  );
}
