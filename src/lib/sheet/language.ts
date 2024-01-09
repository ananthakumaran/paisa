import { parser } from "./parser";
import { parser as searchQueryParser } from "../search/parser/parser";
import { LRLanguage, LanguageSupport } from "@codemirror/language";
import { parseMixed } from "@lezer/common";

export const sheetLanguage = LRLanguage.define({
  name: "sheet",
  parser: parser.configure({
    wrap: parseMixed((node) => {
      if (node.name == "SearchQuery") {
        return { parser: searchQueryParser };
      }
      return null;
    })
  }),
  languageData: {
    closeBrackets: { brackets: ["[", "(", "/", '"', "`", "{"] }
  }
});

export function sheetExtension() {
  return new LanguageSupport(sheetLanguage);
}
