import type { StreamParser, StringStream } from "@codemirror/language";
import helpers from "$lib/template_helpers";
import _ from "lodash";

const KEYWORDS = /^(?:[#/](?:if|with|true|false|unless|each))/;
const HELPERS = new RegExp(
  `^(?:(?:${_.keys(helpers)
    .sort((a, b) => b.length - a.length)
    .join("|")}))`
); //;
const NAMES = /^(?:ROW|SHEET)/;
const COLUMN = /^(?:\.[A-Z]|[A-Z] |\.[0-9]+)/;

interface HandlebarsState {
  inInterpolation: boolean;
}

export const handlebars: StreamParser<HandlebarsState> = {
  name: "handlebars",
  startState: function () {
    return { inInterpolation: false };
  },
  token: function (stream: StringStream, state: HandlebarsState) {
    if (stream.eatSpace()) return null;

    if (stream.match(/^{{/)) {
      state.inInterpolation = true;
      return "bracket";
    }

    if (stream.match(/^}}/)) {
      state.inInterpolation = false;
      return "bracket";
    }

    if (state.inInterpolation) {
      if (stream.match(KEYWORDS)) return "keyword";
      if (stream.match(HELPERS)) return "def";
      if (stream.match(NAMES)) return "tagName";
      if (stream.match(COLUMN)) return "strong";
      if (stream.match(/^"[^"]*"/)) return "string";
    }

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
