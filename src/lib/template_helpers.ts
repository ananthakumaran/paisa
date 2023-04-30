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
  eq: (v1: any, v2: any) => v1 === v2,
  ne: (v1: any, v2: any) => v1 !== v2,
  lt: (v1: any, v2: any) => v1 < v2,
  gt: (v1: any, v2: any) => v1 > v2,
  lte: (v1: any, v2: any) => v1 <= v2,
  gte: (v1: any, v2: any) => v1 >= v2,
  not: (pred: any) => !pred,
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
    return dayjs(str, format).isValid();
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

    const prefix = options.hash.prefix || "";
    const matches = findMatch(query);
    const match = _.find(matches, ([account]) => account.toString().startsWith(prefix));
    if (match) {
      return match[0];
    }
    return prefix + ":Unknown";
  },
  isBlank(str: string) {
    return _.isEmpty(str) || _.trim(str) === "";
  },
  date(str: string, format: string) {
    return dayjs(str, format).format("YYYY/MM/DD");
  },
  acronym(str: string) {
    return _.chain(str.split(" "))
      .filter((s) => !_.includes(STOP_WORDS, s.toLowerCase()))
      .map((s) => {
        if (s.match(/^[0-9]+$/)) {
          return s.toUpperCase();
        }
        return s[0];
      })
      .value()
      .join("");
  }
};
