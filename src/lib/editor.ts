import { ajax, type LedgerFile, type LedgerFileError } from "$lib/utils";
import { ledger } from "$lib/parser";
import { StreamLanguage } from "@codemirror/language";
import { placeholder, keymap } from "@codemirror/view";
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
}

const initialEditorState: EditorState = {
  hasUnsavedChanges: false,
  undoDepth: 0,
  redoDepth: 0,
  errors: []
};

export const editorState = writable(initialEditorState);

async function lint(editor: EditorView): Promise<Diagnostic[]> {
  const doc = editor.state.doc;
  const response = await ajax("/api/editor/validate", {
    method: "POST",
    body: JSON.stringify({ name: "", content: editor.state.doc.toString() })
  });

  editorState.update((current) => _.assign({}, current, { errors: response.errors }));

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
  autocompletions: Record<string, string[]>
) {
  editorState.set(initialEditorState);

  return new EditorView({
    extensions: [
      keymap.of([{ key: "Tab", run: insertTab }]),
      basicSetup,
      placeholder(
        "2023/01/27 Description\n\t Income:Salary:Globex   100,000 INR\n\t Assets:Checking"
      ),
      EditorView.contentAttributes.of({ "data-enable-grammarly": "false" }),
      StreamLanguage.define(ledger),
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
