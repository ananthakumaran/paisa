import { ajax, type LedgerFile } from "$lib/utils";
import { ledger } from "$lib/parser";
import { StreamLanguage } from "@codemirror/language";
import { placeholder, keymap } from "@codemirror/view";
import { basicSetup, EditorView } from "codemirror";
import { indentWithTab } from "@codemirror/commands";
import { linter, lintGutter, type Diagnostic } from "@codemirror/lint";
import _ from "lodash";

async function lint(editor: EditorView): Promise<Diagnostic[]> {
  const doc = editor.state.doc;
  const response = await ajax("/api/editor/validate", {
    method: "POST",
    body: JSON.stringify({ name: "", content: editor.state.doc.toString() })
  });
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

export function createEditor(file: LedgerFile, dom: Element) {
  return new EditorView({
    extensions: [
      basicSetup,
      keymap.of([indentWithTab]),
      placeholder(
        "2023/01/27 Description\n\t Income:Salary:Globex   100,000 INR\n\t Assets:Checking"
      ),
      EditorView.contentAttributes.of({ "data-enable-grammarly": "false" }),
      StreamLanguage.define(ledger),
      lintGutter(),
      linter(lint)
    ],
    doc: file.content,
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
