import { autocompletion, completionKeymap } from "@codemirror/autocomplete";
import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
import { syntaxHighlighting } from "@codemirror/language";
import { classHighlighter } from "@lezer/highlight";
import { lintKeymap } from "@codemirror/lint";
import { search, searchKeymap } from "@codemirror/search";
import type { Extension } from "@codemirror/state";
import {
  drawSelection,
  dropCursor,
  highlightActiveLine,
  highlightActiveLineGutter,
  highlightSpecialChars,
  keymap,
  lineNumbers
} from "@codemirror/view";

export const basicSetup: Extension = [
  lineNumbers(),
  highlightActiveLineGutter(),
  highlightSpecialChars(),
  history(),
  drawSelection(),
  dropCursor(),
  syntaxHighlighting(classHighlighter),
  autocompletion(),
  highlightActiveLine(),
  search({ top: true }),
  keymap.of([
    ...defaultKeymap,
    ...searchKeymap,
    ...historyKeymap,
    ...completionKeymap,
    ...lintKeymap
  ])
];
