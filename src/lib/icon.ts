import _ from "lodash";
import { stemmer } from "stemmer";

interface IconLookup {
  class: string;
  glyph: string;
  words: string[];
}

const ICON_LOOKUP: IconLookup[] = [
  { class: "house-user", glyph: "\ue065", words: ["rent"] },
  { class: "motorcycle", glyph: "\uf21c", words: ["bike", "motorcycl"] },
  { class: "car", glyph: "\uf1b9", words: ["car", "vehicl"] },
  { class: "piggy-bank", glyph: "\uf4d3", words: ["check", "bank"] },
  { class: "sack-dollar", glyph: "\uf81d", words: ["asset"] },
  { class: "chart-line", glyph: "\uf201", words: ["equiti"] },
  { class: "money-check-dollar", glyph: "\uf53d", words: ["debt"] },
  { class: "ring", glyph: "\uf70b", words: ["gold"] },
  { class: "house", glyph: "\uf015", words: ["realest", "hous", "homeloan"] },
  { class: "credit-card", glyph: "\uf09d", words: ["liabil"] },
  { class: "cc-visa", glyph: "\uf1f0", words: ["idfc"] },
  { class: "cc-amazon-pay", glyph: "\uf42d", words: ["icici"] },
  { class: "wallet", glyph: "\uf555", words: ["expens"] },
  { class: "building", glyph: "\uf1ad", words: ["incom"] },
  { class: "gift", glyph: "\uf06b", words: ["gift", "reward"] },
  { class: "tshirt", glyph: "\uf553", words: ["cloth"] },
  { class: "graduation-cap", glyph: "\uf19d", words: ["educ"] },
  { class: "hamburger", glyph: "\uf805", words: ["food"] },
  { class: "film", glyph: "\uf008", words: ["entertain"] },
  { class: "car-crash", glyph: "\uf5e1", words: ["insur"] },
  { class: "hand-holding-dollar", glyph: "\uf4c0", words: ["interest"] },
  { class: "shopping-car", glyph: "\uf07a", words: ["shop"] },
  { class: "utensils", glyph: "\uf2e7", words: ["restaur"] },
  { class: "layer-group", glyph: "\uf5fd", words: ["misc"] },
  { class: "plug-circle-bolt", glyph: "\ue55b", words: ["util"] }
];

const ICONS: Record<string, string> = _.chain(ICON_LOOKUP)
  .flatMap((icon) => _.map(icon.words, (word) => [word, icon.glyph]))
  .fromPairs()
  .value();

export function iconText(account: string): string {
  if (account == "") {
    return "";
  }
  const parts = account.split(":");
  const part = stemmer(_.last(parts).toLowerCase());
  return ICONS[part] || iconText(_.dropRight(parts, 1).join(":"));
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
