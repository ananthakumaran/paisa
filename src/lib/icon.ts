import _ from "lodash";
import { stemmer } from "stemmer";

const ICONS: Record<string, [string, string]> = {
  rent: ["\ue065", "house-user"],
  check: ["\uf4d3", "piggy-bank"],
  bank: ["\uf4d3", "piggy-bank"],
  asset: ["\uf81d", "sack-dollar"],
  equiti: ["\uf201", "chart-line"],
  debt: ["\uf53d", "money-check-dollar"],
  gold: ["\uf70b", "ring"],
  realest: ["\uf015", "house"],
  hous: ["\uf015", "house"],
  liabil: ["\uf09d", "credit-card"],
  expens: ["\uf555", "wallet"],
  incom: ["\uf1ad", "building"],
  gift: ["\uf06b", "gift"],
  cloth: ["\uf553", "tshirt"],
  educ: ["\uf19d", "graduation-cap"],
  food: ["\uf805", "hamburger"],
  entertain: ["\uf008", "film"],
  insur: ["\uf5e1", "car-crash"],
  interest: ["\uf4c0", "hand-holding-dollar"],
  shop: ["\uf07a", "shopping-car"],
  restaur: ["\uf2e7", "utensils"],
  misc: ["\uf5fd", "layer-group"],
  util: ["\ue55b", "plug-circle-bolt"]
};

export function iconText(account: string): string {
  if (account == "") {
    return "";
  }
  const parts = account.split(":");
  const part = stemmer(_.last(parts).toLowerCase());
  return ICONS[part]?.[0] || iconText(_.dropRight(parts, 1).join(":"));
}

export function iconify(account: string, options?: { group?: string; suffix?: boolean }) {
  const icon = options?.group ? iconText(options.group + ":" + account) : iconText(account);
  if (icon == "") {
    return account;
  } else {
    if (options?.suffix) {
      return account + " " + icon;
    } else {
      return icon + " " + account;
    }
  }
}
