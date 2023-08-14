import chroma from "chroma-js";
import _ from "lodash";
import { getColorPreference } from "./utils";
import * as d3 from "d3";

const COLORS = {
  gain: "#b2df8a",
  gainText: "#48c78e",
  loss: "#fb9a99",
  lossText: "#f14668",
  danger: "#cc0f35",
  success: "#257953",
  warn: "#f4a261",
  primary: "#1f77b4",
  secondary: "#17becf",
  tertiary: "#ff7f0e",
  diff: "#4a4a4a",
  assets: "#4e79a7",
  expenses: "#e15759",
  income: "#59a14f",
  liabilities: "#e78ac3",
  equity: "#edc949"
};
export default COLORS;

export function generateColorScheme(domain: string[]) {
  let colors: string[];
  const n = domain.length;

  if (_.every(domain, (d) => _.has(COLORS, d.toLowerCase()))) {
    colors = _.map(domain, (d) => (COLORS as Record<string, string>)[d.toLowerCase()]);
  } else {
    if (n <= 12) {
      colors = {
        1: ["#7570b3"],
        2: ["#7fc97f", "#fdc086"],
        3: ["#66c2a5", "#fc8d62", "#8da0cb"],
        4: ["#66c2a5", "#fc8d62", "#8da0cb", "#e78ac3"],
        5: ["#66c2a5", "#fc8d62", "#8da0cb", "#e78ac3", "#a6d854"],
        6: ["#4e79a7", "#f28e2c", "#e15759", "#76b7b2", "#59a14f", "#edc949"],
        7: ["#4e79a7", "#f28e2c", "#e15759", "#76b7b2", "#59a14f", "#edc949", "#af7aa1"],
        8: ["#4e79a7", "#f28e2c", "#e15759", "#76b7b2", "#59a14f", "#edc949", "#af7aa1", "#ff9da7"],
        9: [
          "#4e79a7",
          "#f28e2c",
          "#e15759",
          "#76b7b2",
          "#59a14f",
          "#edc949",
          "#af7aa1",
          "#ff9da7",
          "#9c755f"
        ],
        10: [
          "#4e79a7",
          "#f28e2c",
          "#e15759",
          "#76b7b2",
          "#59a14f",
          "#edc949",
          "#af7aa1",
          "#ff9da7",
          "#9c755f",
          "#bab0ab"
        ],
        11: [
          "#8dd3c7",
          "#ffed6f",
          "#bebada",
          "#fb8072",
          "#80b1d3",
          "#fdb462",
          "#b3de69",
          "#fccde5",
          "#d9d9d9",
          "#bc80bd",
          "#ccebc5"
        ],
        12: [
          "#8dd3c7",
          "#e5c494",
          "#bebada",
          "#fb8072",
          "#80b1d3",
          "#fdb462",
          "#b3de69",
          "#fccde5",
          "#d9d9d9",
          "#bc80bd",
          "#ccebc5",
          "#ffed6f"
        ]
      }[n];
    } else {
      const z = d3
        .scaleSequential()
        .domain([0, n - 1])
        .interpolator(d3.interpolateSinebow);
      colors = _.map(_.range(0, n), (n) => chroma(z(n)).desaturate(1.5).hex());
    }
  }

  return d3.scaleOrdinal<string>().domain(domain).range(colors);
}

export function genericBarColor() {
  return getColorPreference() == "dark" ? "hsl(215, 18%, 15%)" : "hsl(0, 0%, 91%)";
}

export function accountColor(account: string) {
  const normalized = account.toLowerCase();

  if (_.includes(["assets", "expenses", "income", "liabilities", "equity"], normalized)) {
    return (COLORS as Record<string, string>)[normalized];
  }

  return "hsl(0, 0%, 48%)";
}
