import * as d3 from "d3";
import _ from "lodash";
import { formatCurrency, formatFixedWidthFloat, secondName, type Posting } from "./utils";

export function formatBreakdownLabel(amount: number, total: number) {
  return `${formatCurrency(amount)} ${formatFixedWidthFloat((amount / total) * 100, 6)}%`;
}

export function breakdownLabelWidth(labels: string[], minWidth = 135) {
  // Monospace labels need about 8px per character, plus two characters of padding.
  const length = d3.max(labels, (label) => label.length) || 0;
  return Math.max(minWidth, (length + 2) * 8);
}

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
