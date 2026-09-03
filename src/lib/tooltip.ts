export type TooltipCell = string | string[];
export type TooltipRow = TooltipCell[];

const TOOLTIP_ROW_HEIGHT = 30;
const TOOLTIP_VIEWPORT_PADDING = 32;

export function fitTooltipRows(
  rows: TooltipRow[],
  options: {
    header?: string;
    total?: string;
  } = {},
  viewportHeight = typeof window === "undefined" ? Infinity : window.innerHeight
) {
  if (rows.length === 0 || !Number.isFinite(viewportHeight)) {
    return [...rows];
  }

  const fixedRows = Number(Boolean(options.header)) + Number(Boolean(options.total));
  const tableCapacity = Math.max(
    1,
    Math.floor((viewportHeight - TOOLTIP_VIEWPORT_PADDING) / TOOLTIP_ROW_HEIGHT)
  );

  if (rows.length + fixedRows <= tableCapacity) {
    return [...rows];
  }

  const visibleRowCount = Math.max(0, tableCapacity - fixedRows - 1);
  const omittedRowCount = rows.length - visibleRowCount;
  const columnCount = rows[0].length;
  const omittedRow: TooltipRow = [
    [`… ${omittedRowCount} rows omitted`, "has-text-grey has-text-centered", columnCount.toString()]
  ];

  const visibleRows = visibleRowCount === 0 ? [] : rows.slice(-visibleRowCount);
  return [omittedRow, ...visibleRows];
}
