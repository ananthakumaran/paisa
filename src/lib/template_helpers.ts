import dayjs from "dayjs";
import _ from "lodash";
import { get } from "svelte/store";
import { accountTfIdf, accountRules } from "../store";
import similarity from "compute-cosine-similarity";

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

/**
 * Specifically finds matches between transactions based on description similarity
 * @param queryTransaction The transaction description to find matches for
 * @param taggedTransactions Array of previously tagged transaction descriptions
 * @param threshold Minimum similarity score to consider a match (default: 0.3)
 * @returns Sorted array of matching transactions with similarity scores
 */
function findTransactionMatches(
  queryTransaction: string, 
  taggedTransactions: string[], 
  threshold = 0.3
) {
  if (!queryTransaction || !taggedTransactions?.length) {
    return [];
  }

  // Process the query transaction text
  const queryTerms = new Set(
    queryTransaction.toLowerCase().split(/\s+/).filter(term => term.length > 2)
  );

  // Compare with each tagged transaction
  return taggedTransactions
    .map(transaction => {
      const transactionTerms = new Set(
        transaction.toLowerCase().split(/\s+/).filter(term => term.length > 2)
      );

      // Calculate Jaccard similarity
      const intersection = new Set(
        [...queryTerms].filter(term => transactionTerms.has(term))
      );
      const union = new Set([...queryTerms, ...transactionTerms]);
      
      const similarity = union.size > 0 ? intersection.size / union.size : 0;
      
      return {
        transaction,
        score: similarity,
        matchingTerms: [...intersection]
      };
    })
    .filter(match => match.score >= threshold)
    .sort((a, b) => b.score - a.score);
}

function scrubAmount(str: string) {
  const amount = _.trim(str)
    .replace(/\((.+)\)/, "-$1")
    .replace(/[^0-9.-]/g, "");

  if (!isNaN(amount as any) && !isNaN(parseFloat(amount))) {
    return amount;
  }
}

function parseAmount(str: string | number) {
  if (_.isNumber(str)) {
    return str;
  }

  const amount = scrubAmount(str);
  if (amount) {
    return parseFloat(amount);
  }
}

export default {
  eq: (a: any, b: any) => a === b,
  ne: (a: any, b: any) => a !== b,
  not: (value: any) => !value,
  gte: (a: string | number, b: string | number) => parseAmount(a) >= parseAmount(b),
  gt: (a: string | number, b: string | number) => parseAmount(a) > parseAmount(b),
  lte: (a: string | number, b: string | number) => parseAmount(a) <= parseAmount(b),
  lt: (a: string | number, b: string | number) => parseAmount(a) < parseAmount(b),
  negate: (value: string) => parseAmount(value) * -1,
  round(str: string, options: any) {
    return _.round(parseAmount(str), options.hash.precision || 0);
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
  predictAccountWithRules(...args: any) {
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

    // First, try to match against regex rules from accountRules store
    const rules = get(accountRules);
    if (rules && rules.length > 0) {
      for (const rule of rules) {
        if (rule.enabled) {
          try {
            const regex = new RegExp(rule.pattern, "i");
            if (regex.test(query)) {
              // If prefix is provided, make sure the account matches
              if (!prefix || rule.account.startsWith(prefix)) {
                return rule.account;
              }
            }
          } catch (e) {
            // Invalid regex, skip this rule
            console.warn(`Invalid regex pattern in rule "${rule.name}": ${rule.pattern}`);
          }
        }
      }
    }

    // Instead of using findMatch, use findTransactionMatches
    if (accountTfIdf === null || get(accountTfIdf) == null) {
      // Return a default account if no TF-IDF data is available
      return prefix.endsWith(":") ? prefix + "Unknown" : prefix + ":Unknown";
    }

    // Extract accounts from the tf_idf data
    const { tf_idf } = get(accountTfIdf);
    const taggedAccounts = Object.keys(tf_idf);
    
    // Use findTransactionMatches to match the query against account names
    const matches = findTransactionMatches(query, taggedAccounts);
    
    // Find the first match that starts with the prefix
    const match = matches.find(item => item.transaction.startsWith(prefix));
    
    if (match) {
      return match.transaction;
    }
    
    // Return default if no match found
    return prefix.endsWith(":") ? prefix + "Unknown" : prefix + ":Unknown";
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
  amount(str: string, options: any) {
    const amount = scrubAmount(str);
    return amount || options.hash.default || "";
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
