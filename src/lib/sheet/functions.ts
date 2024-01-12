import type { Posting } from "$lib/utils";
import _ from "lodash";
import { assertType, type Environment, type Query } from "./interpreter";
import { BigNumber } from "bignumber.js";

type PostingsOrQuery = Posting[] | Query;

function cost(env: Environment, q: PostingsOrQuery): BigNumber {
  assertType("Postings", q);

  const ps = toPostings(env, q);
  return ps.reduce((acc, p) => acc.plus(new BigNumber(p.amount)), new BigNumber(0));
}

function negate(env: Environment, q: PostingsOrQuery): Posting[] {
  assertType("Postings", q);

  const ps = toPostings(env, q);
  return ps.map((p) => {
    p = { ...p };
    p.quantity = -p.quantity;
    p.amount = -p.amount;
    p.market_amount = -p.market_amount;
    return p;
  });
}

function fifo(env: Environment, q: PostingsOrQuery): Posting[] {
  assertType("Postings", q);

  const ps = toPostings(env, q);
  return _.chain(ps)
    .groupBy((p) => [p.account, p.commodity].join(":"))
    .map((ps) => {
      ps = _.sortBy(ps, (p) => p.date);
      const available: Posting[] = [];
      while (ps.length > 0) {
        const p = ps.shift();
        if (p.quantity >= 0) {
          available.push(p);
        } else {
          let quantity = -p.quantity;
          while (quantity > 0 && available.length > 0) {
            const a = available.shift();
            if (a.quantity > quantity) {
              const diff = a.quantity - quantity;
              const price = a.amount / a.quantity;
              const marketPrice = a.market_amount / a.quantity;
              available.unshift({
                ...a,
                quantity: diff,
                amount: diff * price,
                market_amount: diff * marketPrice
              });
              quantity = 0;
            } else {
              quantity -= a.quantity;
            }
          }
        }
      }
      return available;
    })
    .flatten()
    .value();
}

function toPostings(env: Environment, q: PostingsOrQuery) {
  if (Array.isArray(q)) {
    return q;
  }
  return q.resolve(env);
}

export const functions = { cost, fifo, negate };
