import type { Posting } from "$lib/utils";
import type { Environment } from "./interpreter";
import { BigNumber } from "bignumber.js";

function cost(env: Environment, ps: Posting[]) {
  return ps.reduce((acc, p) => acc.plus(new BigNumber(p.amount)), new BigNumber(0));
}

export const functions = { cost };
