import { type CellComponent } from "tabulator-tables";
import { formatCurrency, formatFloat, formatPercentage, isZero, lastName } from "./utils";
import { iconText } from "./icon";

export function indendedAssetAccountName(cell: CellComponent) {
  const account = cell.getValue();
  let children = "";
  const data = cell.getData();
  if ((data._children?.length || 0) > 0) {
    children = `(${data._children?.length})`;
  }
  return `
<span class="whitespace-nowrap" style="max-width: max(15rem, 33.33vw); overflow: hidden;">
  <span class="has-text-grey custom-icon">${iconText(account)}</span>
  <a href="/assets/gain/${account}">${lastName(account)}</a>
  <span class="has-text-grey-light is-size-7">${children}</span>
</span>
`;
}

export function indendedLiabilityAccountName(cell: CellComponent) {
  const account = cell.getValue();
  let children = "";
  const data = cell.getData();
  if ((data._children?.length || 0) > 0) {
    children = `(${data._children?.length})`;
  }
  return `
<span class="whitespace-nowrap" style="max-width: max(15rem, 33.33vw); overflow: hidden;">
  <span class="has-text-grey custom-icon">${iconText(account)}</span>
  <span>${lastName(account)}</span>
  <span class="has-text-grey-light is-size-7">${children}</span>
</span>
`;
}

export function accountName(cell: CellComponent) {
  const account = cell.getValue();
  return `
<span class="whitespace-nowrap" style="max-width: max(15rem, 33.33vw); overflow: hidden;">
  <span class="has-text-grey custom-icon">${iconText(account)}</span>
  <a href="/assets/gain/${account}">${account}</a>
</span>
`;
}

function calculateChangeClass(gain: number) {
  let changeClass = "";
  if (gain > 0) {
    changeClass = "has-text-success";
  } else if (gain < 0) {
    changeClass = "has-text-danger";
  }
  return changeClass;
}

export function nonZeroCurrency(cell: CellComponent) {
  const value = cell.getValue();
  return isZero(value) ? "" : formatCurrency(value);
}

export function nonZeroFloatChange(cell: CellComponent) {
  const value = cell.getValue();
  return isZero(value)
    ? ""
    : `<span class="${calculateChangeClass(value)}">${formatFloat(value)}</span>`;
}

export function nonZeroPercentageChange(cell: CellComponent) {
  const value = cell.getValue();
  return isZero(value)
    ? ""
    : `<span class="${calculateChangeClass(value)}">${formatPercentage(value, 2)}</span>`;
}

export function formatCurrencyChange(cell: CellComponent) {
  const value = cell.getValue();
  return isZero(value)
    ? ""
    : `<span class="${calculateChangeClass(value)}">${formatCurrency(value)}</span>`;
}
