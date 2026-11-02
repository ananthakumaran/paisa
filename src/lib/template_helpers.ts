import dayjs from "dayjs";
import _ from "lodash";
import { get } from "svelte/store";
import { accountTfIdf } from "../store";
import similarity from "compute-cosine-similarity";
import { BigNumber } from "bignumber.js";
import { Amount, parseAmount as parseLocalizedAmount } from "./amount";

const STOP_WORDS = ["", "fof", "growth", "direct", "plan", "the"];

function tokenize(s: string) {
  return _.mapValues(
    _.groupBy(
      s
        .split(/[ .()/:]+/)
        .map((s) => s.toLowerCase())
        .filter((s) => s.trim() !== ""),
      _.identity
    ),
    (v) => v.length
  );
}

function nextChar(key: string): string {
  if (key === "Z") {
    return "AA";
  } else {
    const last = key.slice(-1);
    const butlast = key.slice(0, -1);
    if (last === "Z") {
      return nextChar(butlast) + "A";
    } else {
      return butlast + String.fromCharCode(last.charCodeAt(0) + 1);
    }
  }
}

function tfidf(query: string) {
  if (accountTfIdf === null || get(accountTfIdf) == null) {
    return {};
  }

  const { index } = get(accountTfIdf);
  const tokens = tokenize(query);
  return _.chain(tokens)
    .map((freq, token) => {
      const tf = freq / Object.keys(tokens).length;
      const idf =
        Math.log(
          Object.keys(index.docs).length / (1 + Object.keys(index.tokens[token] || []).length)
        ) + 1;
      return [token, tf * idf];
    })
    .fromPairs()
    .value();
}

function findMatch(query: string) {
  if (accountTfIdf === null || get(accountTfIdf) == null) {
    return [];
  }

  const queryVector = tfidf(query);
  const { tf_idf, index } = get(accountTfIdf);
  const accounts = Object.keys(index.docs);
  return _.chain(accounts)
    .map((account) => {
      const tokens = _.uniq(_.concat(Object.keys(queryVector), Object.keys(tf_idf[account])));
      const q = tokens.map((token) => queryVector[token] || 0);
      const a = tokens.map((token) => tf_idf[account][token] || 0);
      return [account, similarity(q, a)];
    })
    .sortBy(([, score]) => score)
    .filter(([, score]: [string, number]) => score > 0)
    .reverse()
    .value();
}

function parseAmount(value: unknown) {
  return parseLocalizedAmount(value, globalThis.USER_CONFIG?.locale || "en-IN");
}

function compareAmounts(a: unknown, b: unknown, test: (comparison: number) => boolean) {
  const left = parseAmount(a);
  const right = parseAmount(b);
  return !!left && !!right && test(left.value.comparedTo(right.value));
}

export default {
  eq: (a: any, b: any) => a === b,
  ne: (a: any, b: any) => a !== b,
  not: (value: any) => !value,
  gte: (a: unknown, b: unknown) => compareAmounts(a, b, (comparison) => comparison >= 0),
  gt: (a: unknown, b: unknown) => compareAmounts(a, b, (comparison) => comparison > 0),
  lte: (a: unknown, b: unknown) => compareAmounts(a, b, (comparison) => comparison <= 0),
  lt: (a: unknown, b: unknown) => compareAmounts(a, b, (comparison) => comparison < 0),
  negate(value: unknown) {
    const amount = parseAmount(value);
    return amount ? new Amount(amount.value.negated()) : "";
  },
  round(value: unknown, options: any) {
    const amount = parseAmount(value);
    if (!amount) {
      return "";
    }
    const precision = _.clamp(_.toInteger(options.hash.precision), 0, 292);
    return new Amount(amount.value.decimalPlaces(precision, BigNumber.ROUND_HALF_CEIL));
  },
  and(...args: any[]) {
    return Array.prototype.every.call(Array.prototype.slice.call(args, 0, -1), Boolean);
  },
  or(...args: any[]) {
    for (const arg of Array.prototype.slice.call(args, 0, -1)) {
      if (arg) {
        return arg;
      }
    }
  },
  isDate(str: string, format: string) {
    if (!_.isString(str)) {
      return false;
    }
    return dayjs(_.trim(str), format, true).isValid();
  },
  predictAccount(...args: any) {
    const options = args.pop();

    let query: string;
    if (args.length === 0) {
      query = Object.values(options.data.root.ROW).join(" ");
    } else {
      query = _.chain(args)
        .map((a) => {
          if (_.isObject(a)) {
            return Object.values(a);
          }
          return a;
        })
        .flattenDeep()
        .value()
        .join(" ");
    }

    const prefix: string = options.hash.prefix || "";
    const matches = findMatch(query);
    const match = _.find(matches, ([account]) => account.toString().startsWith(prefix));
    if (match) {
      return match[0];
    }
    if (prefix.endsWith(":")) {
      return prefix + "Unknown";
    } else {
      return prefix + ":Unknown";
    }
  },
  isBlank(str: string) {
    return _.isEmpty(str) || _.trim(str) === "";
  },
  amount(value: unknown, options: any) {
    return parseAmount(value) ?? options.hash.default ?? "";
  },
  date(str: string, format: string) {
    return dayjs(_.trim(str), format, true).format("YYYY/MM/DD");
  },
  trim(str: string) {
    return _.trim(str);
  },
  replace(str: string, search: string, replace: string) {
    if (!_.isString(str)) {
      return;
    }
    return str.replaceAll(search, replace);
  },
  textRange(fromColumn: string, toColumn: string, options: any) {
    const row: Record<string, string> = options.data.root.ROW;
    const cells = [];
    let i = 0;
    let current = fromColumn;
    while (i < 1000) {
      cells.push(row[current]);
      if (current === toColumn) {
        break;
      }
      current = nextChar(current);
      i++;
    }
    return cells.join(options.hash.separator || " ");
  },
  regexpTest(str: string, regexp: string) {
    if (!_.isString(str)) {
      return;
    }

    return new RegExp(regexp).test(str);
  },
  regexpMatch(str: string, regexp: string, options: any) {
    if (!_.isString(str)) {
      return;
    }

    const group = options.hash.group || 0;

    const match = new RegExp(regexp).exec(str);
    if (match) {
      return match[group];
    }
  },
  match(str: string, options: any) {
    for (const [value, regexp] of Object.entries(options.hash as Record<string, string>)) {
      if (new RegExp(regexp).test(str)) {
        return value;
      }
    }
    return null;
  },
  findAbove(column: string, options: any) {
    const regexp = new RegExp(options.hash.regexp || ".+");
    let i: number = options.data.root.ROW.index - 1;
    while (i >= 0) {
      const row = options.data.root.SHEET[i];
      const cell = row[column] || "";
      const match = cell.match(regexp);
      if (match) {
        if (options.hash.group) {
          return match[options.hash.group];
        }
        return cell;
      }
      i--;
    }
    return null;
  },
  findBelow(column: string, options: any) {
    const regexp = new RegExp(options.hash.regexp || ".+");
    let i: number = options.data.root.ROW.index + 1;
    while (i < options.data.root.SHEET.length) {
      const row = options.data.root.SHEET[i];
      const cell = row[column] || "";
      const match = cell.match(regexp);
      if (match) {
        if (options.hash.group) {
          return match[options.hash.group];
        }
        return cell;
      }
      i++;
    }
    return null;
  },
  acronym(str: string) {
    return _.chain(str.replaceAll(/[^a-zA-Z ]/g, "").split(" "))
      .filter((s) => !_.includes(STOP_WORDS, s.toLowerCase()))
      .map((s) => {
        return s[0].toUpperCase();
      })
      .value()
      .join("");
  },
  toLowerCase(str: string) {
    return str.toLowerCase();
  },
  toUpperCase(str: string) {
    return str.toUpperCase();
  },
  capitalize(str: string) {
    return _.capitalize(str);
  }
};
