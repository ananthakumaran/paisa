import _ from "lodash";
import arcticons from "../../fonts/arcticons-info.json";
import fa6brands from "../../fonts/fa6-brands-info.json";
import fa6regular from "../../fonts/fa6-regular-info.json";
import fa6solid from "../../fonts/fa6-solid-info.json";
import fluentemoji from "../../fonts/fluent-emoji-high-contrast-info.json";
import mdi from "../../fonts/mdi-info.json";
import { stemmer } from "stemmer";

const icons = {
  arcticons: arcticons["codepoints"],
  "fa6-brands": fa6brands["codepoints"],
  "fa6-regular": fa6regular["codepoints"],
  "fa6-solid": fa6solid["codepoints"],
  "fluent-emoji-high-contrast": fluentemoji["codepoints"],
  mdi: mdi["codepoints"]
};

export function iconGlyph(symbol: string): string {
  if (!symbol) {
    return String.fromCodePoint(65533);
  }
  const [font, name] = symbol.split(":");
  const code = (icons as Record<string, Record<string, number>>)[font]?.[name] || 65533;
  return String.fromCodePoint(code);
}

export const iconsList = _.flatMap(icons, (glyph, font) => {
  return _.map(glyph, (code, name) => `${font}:${name}`);
});

interface IconLookup {
  symbol: string;
  words: string[];
}

const ICON_LOOKUP: IconLookup[] = [
  { symbol: "fa6-solid:house-user", words: ["rent"] },
  { symbol: "fa6-solid:motorcycle", words: ["bike", "motorcycl"] },
  { symbol: "fa6-solid:car", words: ["car", "vehicl"] },
  { symbol: "fa6-solid:piggy-bank", words: ["check", "bank"] },
  { symbol: "fa6-solid:sack-dollar", words: ["asset"] },
  { symbol: "fa6-solid:chart-line", words: ["equiti"] },
  { symbol: "fa6-solid:money-check-dollar", words: ["debt"] },
  { symbol: "fa6-solid:ring", words: ["gold"] },
  { symbol: "fa6-solid:house", words: ["realest", "hous", "homeloan"] },
  { symbol: "fa6-solid:credit-card", words: ["liabil"] },
  { symbol: "fa6-solid:wallet", words: ["expens"] },
  { symbol: "fa6-solid:building", words: ["incom"] },
  { symbol: "fa6-solid:gift", words: ["gift", "reward"] },
  { symbol: "fa6-solid:shirt", words: ["cloth"] },
  { symbol: "fa6-solid:graduation-cap", words: ["educ"] },
  { symbol: "fa6-solid:burger", words: ["food"] },
  { symbol: "fa6-solid:film", words: ["entertain"] },
  { symbol: "fa6-solid:car-crash", words: ["insur"] },
  { symbol: "fa6-solid:hand-holding-dollar", words: ["interest"] },
  { symbol: "fa6-solid:cart-shopping", words: ["shop"] },
  { symbol: "fa6-solid:utensils", words: ["restaur"] },
  { symbol: "fa6-solid:layer-group", words: ["misc"] },
  { symbol: "fa6-solid:plug-circle-bolt", words: ["util"] },
  { symbol: "fa6-solid:carrot", words: ["veget"] },
  { symbol: "fa6-solid:taxi", words: ["transport"] },
  { symbol: "fa6-solid:money-bill", words: ["cash"] }
];

const ICONS: Record<string, string> = _.chain(ICON_LOOKUP)
  .flatMap((icon) => _.map(icon.words, (word) => [word, iconGlyph(icon.symbol)]))
  .fromPairs()
  .value();

export function iconText(account: string): string {
  if (account == "") {
    return "";
  }

  const accountConfig = (USER_CONFIG.accounts || []).find((a) => a.name == account);
  if (!_.isEmpty(accountConfig?.icon)) {
    return iconGlyph(accountConfig.icon);
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
