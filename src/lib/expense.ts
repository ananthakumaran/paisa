import * as d3 from "d3";
import _ from "lodash";
import { secondName, type Posting } from "./utils";

export function pieData(expenses: Posting[]) {
  return d3
    .pie<{ category: string; total: number }>()
    .value((g) => g.total)
    .sort((a, b) => a.category.localeCompare(b.category))(_.values(byExpenseGroup(expenses)));
}

export function byExpenseGroup(expenses: Posting[]) {
  return _.chain(expenses)
    .groupBy(expenseGroup)
    .mapValues((ps, category) => {
      return {
        category: category,
        postings: ps,
        total: _.sumBy(ps, (p) => p.amount)
      };
    })
    .value();
}

export function expenseGroup(posting: Posting) {
  return secondName(posting.account);
}
