import { ajax } from "$lib/utils";
import { ledger } from "$lib/parser";
import { StreamLanguage } from "@codemirror/language";
import { keymap, type KeyBinding } from "@codemirror/view";
import { EditorState as State } from "@codemirror/state";
import { EditorView } from "codemirror";
import { basicSetup } from "./editor/base";
import { history, undoDepth, redoDepth } from "@codemirror/commands";
import { linter, lintGutter, type Diagnostic } from "@codemirror/lint";
import _ from "lodash";
import { editorState, initialEditorState } from "../store";
import {
  CompletionContext,
  autocompletion,
  completeFromList,
  ifIn
} from "@codemirror/autocomplete";
import { MergeView } from "@codemirror/merge";
import { schedulePlugin } from "./transaction_tag";
import dayjs from "dayjs";

export { editorState } from "../store";

async function lint(editor: EditorView): Promise<Diagnostic[]> {
  const doc = editor.state.doc;
  const response = await ajax("/api/editor/validate", {
    method: "POST",
    body: JSON.stringify({ name: "", content: editor.state.doc.toString() }),
    background: true
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

export function createDiffEditor(oldContent: string, newContent: string, dom: Element) {
  const extensions = [
    basicSetup,
    State.readOnly.of(true),
    StreamLanguage.define(ledger),
    EditorView.contentAttributes.of({ "data-enable-grammarly": "false" }),
    lintGutter(),
    linter(lint)
  ];
  return new MergeView({
    a: { extensions: extensions, doc: oldContent },
    b: { extensions: extensions, doc: newContent },
    parent: dom,
    collapseUnchanged: {}
  });
}

export function createEditor(
  content: string,
  dom: Element,
  opts: {
    autocompletions?: Record<string, string[]>;
    readonly?: boolean;
    keybindings?: readonly KeyBinding[];
  }
) {
  editorState.set(initialEditorState);

  return new EditorView({
    extensions: [
      keymap.of(opts.keybindings || []),
      basicSetup,
      State.readOnly.of(!!opts.readonly),
      EditorView.contentAttributes.of({ "data-enable-grammarly": "false" }),
      StreamLanguage.define(ledger),
      lintGutter(),
      linter(lint),
      history(),
      autocompletion({
        override: [
          (context: CompletionContext) => {
            if (context.matchBefore(/^20$/)) {
              return completeFromList([dayjs().format("YYYY/MM/DD") + " "])(context);
            }
            return null;
          },
          ..._.map(opts.autocompletions || [], (options: string[], node) =>
            ifIn([node], completeFromList(options))
          )
        ]
      }),
      EditorView.updateListener.of((viewUpdate) => {
        editorState.update((current) =>
          _.assign({}, current, {
            hasUnsavedChanges: current.hasUnsavedChanges || viewUpdate.docChanged,
            undoDepth: undoDepth(viewUpdate.state),
            redoDepth: redoDepth(viewUpdate.state)
          })
        );
      }),
      schedulePlugin
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

export function moveToLine(editor: EditorView, lineNumber: number, cursor = false) {
  try {
    const line = editor.state.doc.line(lineNumber);
    editor.dispatch(
      editor.state.update({
        effects: EditorView.scrollIntoView(line.from, { y: "center" })
      })
    );

    if (cursor) {
      editor.dispatch({ selection: { anchor: line.from, head: line.from } });
    }
  } catch (_e) {
    // ignore invalid line number
  }
}

export function updateContent(editor: EditorView, content: string) {
  const head = editor.state.selection.main.head;
  const line = editor.state.doc.lineAt(head);
  const lineNumber = line.number;
  const column = head - line.from;
  editor.dispatch(
    editor.state.update({ changes: { from: 0, to: editor.state.doc.length, insert: content } })
  );

  const newLine = editor.state.doc.line(lineNumber);
  const newColumn = Math.min(newLine.from + column, newLine.to);
  editor.dispatch({ selection: { anchor: newColumn, head: newColumn } });
}
