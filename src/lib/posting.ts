import _ from "lodash";
import { now, type Posting } from "./utils";

export interface Change {
  class: string;
  value: number;
  percentage: number;
  days: number;
}

export function change(p: Posting): Change {
  let changePercentage = 0,
    days = 0,
    changeAmount = 0,
    changeClass = "";
  if (p.commodity !== USER_CONFIG.default_currency) {
    days = now().diff(p.date, "days");
    if (p.quantity > 0 && days > 0) {
      changeAmount = p.market_amount - p.amount;
      if (changeAmount > 0) {
        changeClass = "has-text-success";
      } else if (changeAmount < 0) {
        changeClass = "has-text-danger";
      }
      const perYear = 365 / days;
      changePercentage = (changeAmount / p.amount) * 100 * perYear;
    }
  }

  return {
    class: changeClass,
    value: changeAmount,
    percentage: changePercentage,
    days
  };
}

export function filterPostings(rows: { date: string; posting: Posting }[], filter: string) {
  let filterRegex = new RegExp(".*", "i");
  if (filter) {
    filterRegex = new RegExp(filter, "i");
  }

  return _.filter(
    rows,
    (r) =>
      filterRegex.test(r.posting.account) ||
      filterRegex.test(r.posting.payee) ||
      filterRegex.test(r.date)
  );
}
