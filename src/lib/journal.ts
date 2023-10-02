import _ from "lodash";

interface State {
  inTransaction: boolean;
  lines: string[];
}

export function format(text: string) {
  const state: State = { inTransaction: false, lines: [] };
  return text
    .split("\n")
    .reduce((state: State, line: string) => {
      state.lines.push(formatLine(line, state));
      return state;
    }, state)
    .lines.join("\n");
}

function space(length: number) {
  return " ".repeat(length);
}

const DATE = /^\d{4}[/-]\d{2}[/-]\d{2}/;

// https://ledger-cli.org/doc/ledger3.html#Journal-Format
function formatLine(line: string, state: State) {
  if (line.match(DATE) || line.match(/^[~=]/)) {
    state.inTransaction = true;
    return line;
  }

  if (_.isEmpty(_.trim(line)) || line.match(/^[^ \t]/)) {
    state.inTransaction = false;
  }

  if (!state.inTransaction) {
    return line;
  }

  const fullMatch = line.match(
    /^[ \t]+(?<account>(?:[*!]\s+)?[^; \t\n](?:(?!\s{2})[^;\t\n])+)[ \t]+(?<prefix>[^;]*?)(?<amount>[+-]?[.,0-9]+)(?<suffix>.*)$/
  );
  if (fullMatch) {
    const { account, prefix, amount, suffix } = fullMatch.groups;
    if (account.length + prefix.length + amount.length <= 46) {
      return (
        space(4) +
        account +
        space(48 - account.length - prefix.length - amount.length) +
        prefix +
        amount +
        suffix
      );
    }
  }

  const partialMatch = line.match(
    /^[ \t]+(?<account>(?:[*!]\s+)?[^; \t\n](?:(?!\s{2})[^;\t\n])+)$/
  );
  if (partialMatch) {
    const { account } = partialMatch.groups;
    return space(4) + account;
  }

  return line;
}
