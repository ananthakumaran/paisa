import type { StreamParser, StringStream } from "@codemirror/language";
import helpers from "$lib/template_helpers";
import _ from "lodash";

const KEYWORDS = /^(?:[#/](?:if|with|true|false))/;
const HELPERS = new RegExp(`^(?:(?:${_.keys(helpers).join("|")}))`); //;
const NAMES = /^(?:ROW)/;
const COLUMN = /^(?:\.[A-Z])/;

export const handlebars: StreamParser<object> = {
  name: "handlebars",
  startState: function () {
    return {};
  },
  token: function (stream: StringStream) {
    if (stream.eatSpace()) return null;

    if (stream.match(KEYWORDS)) return "keyword";
    if (stream.match(HELPERS)) return "keyword";
    if (stream.match(NAMES)) return "tagName";
    if (stream.match(COLUMN)) return "strong";
    if (stream.match(/^"[^"]*"/)) return "string";

    const ch = stream.next();
    if (!ch) {
      return;
    }

    if (ch == ";") {
      stream.skipToEnd();
      return "comment";
    }

    if (_.includes(["(", ")", "{", "}"], ch)) {
      return "bracket";
    }
  }
};
