import dayjs from "dayjs";
import _ from "lodash";
import { get } from "svelte/store";
import { accountTfIdf } from "../store";
import similarity from "compute-cosine-similarity";

const STOP_WORDS = ["fof", "growth", "direct", "plan", "the"];

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

function tfidf(query: string) {
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
    .filter(([, score]) => score > 0)
    .reverse()
    .value();
}

export default {
  eq: (a: any, b: any) => a === b,
  ne: (a: any, b: any) => a !== b,
  not: (value: any) => !value,
  negate: (value: string) => parseFloat(value) * -1,
  and(...args: any[]) {
    return Array.prototype.every.call(Array.prototype.slice.call(args, 0, -1), Boolean);
  },
  or(...args: any[]) {
    return Array.prototype.slice.call(args, 0, -1).some(Boolean);
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
  date(str: string, format: string) {
    return dayjs(_.trim(str), format).format("YYYY/MM/DD");
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
  findAbove(column: string, options: any) {
    const regexp = new RegExp(options.hash.regexp || ".+");
    let i: number = options.data.root.ROW.index;
    while (i >= 0) {
      const row = options.data.root.SHEET[i];
      const cell = row[column] || "";
      const match = cell.match(regexp);
      if (match) {
        return cell;
      }
      i--;
    }
    return null;
  },
  acronym(str: string) {
    return _.chain(str.split(" "))
      .filter((s) => !_.includes(STOP_WORDS, s.toLowerCase()))
      .map((s) => {
        if (s.match(/^[0-9]+$/)) {
          return "";
        }
        return s[0];
      })
      .value()
      .join("");
  }
};
