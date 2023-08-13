import { getColorPreference } from "./utils";

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
  assets: "#76b7b2",
  expenses: "#e15759",
  income: "#59a14f",
  liabilities: "#e78ac3",
  equity: "#edc949"
};
export default COLORS;

export function genericBarColor() {
  return getColorPreference() == "dark" ? "hsl(215, 18%, 15%)" : "hsl(0, 0%, 91%)";
}
