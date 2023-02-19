import type { StreamParser, StringStream } from "@codemirror/language";

const DATE = /^\d{4}[/-]\d{2}[/-]\d{2}/;
const AMOUNT = /^[+-]?(?:[0-9,])+(\.(?:[0-9,])+)?/;
const KEYWORDS = /^(?:commodity)/;
const UNIT = /^(?:INR)/;
const ACCOUNT = /^[A-Za-z]([A-Za-z0-9:])*/;

interface State {
  inTransaction: boolean;
  inPosting: boolean;
  inInclude: boolean;
}

export const ledger: StreamParser<State> = {
  name: "ledger",
  startState: function () {
    return {
      inTransaction: false,
      inPosting: false,
      inInclude: false
    };
  },
  token: function (stream: StringStream, state: State) {
    if (stream.eatSpace()) return null;

    if (stream.match("include")) {
      state.inInclude = true;
      return "keyword";
    }

    if (state.inInclude) {
      state.inInclude = false;
      stream.skipToEnd();
      return "link";
    }

    if (stream.match(KEYWORDS)) return "keyword";
    if (stream.match(UNIT)) return "unit";

    if (stream.match(DATE)) {
      state.inTransaction = true;
      state.inPosting = false;
      return "tagName";
    }

    if (stream.match(AMOUNT)) return "atom";

    const ch = stream.next();
    if (!ch) {
      return;
    }

    if (ch == ";") {
      stream.skipToEnd();
      return "comment";
    }

    if (state.inTransaction) {
      state.inTransaction = false;
      state.inPosting = true;
      stream.skipToEnd();
      return "strong";
    }

    if (state.inPosting && stream.match(ACCOUNT)) {
      return "string";
    }

    if (ch == "@") {
      return "operator";
    }
  }
};
