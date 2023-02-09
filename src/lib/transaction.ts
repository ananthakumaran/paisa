import _ from "lodash";
import type { Transaction } from "./utils";

export function filterTransactions(transactions: Transaction[], filter: string | null) {
  let filterRegex = new RegExp(".*", "i");
  if (filter) {
    filterRegex = new RegExp(filter, "i");
  }

  return _.filter(transactions, (t) => {
    return (
      filterRegex.test(t.payee) ||
      filterRegex.test(t.date.format("DD MMM YYYY")) ||
      _.some(t.postings, (p) => filterRegex.test(p.account))
    );
  });
}
