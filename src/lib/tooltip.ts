const TOOLTIP_ROW_HEIGHT = 24;
const TOOLTIP_VERTICAL_SPACING = 32;

export function tooltip(
  rows: Array<Array<string | string[]>>,
  options: {
    header?: string;
    total?: string;
  } = {}
) {
  const total = options.total;
  const header = options.header;
  const hasTotal = total !== undefined;
  const hasHeader = Boolean(header);
  const columnCount = Math.max(2, rows[0]?.length ?? 2);
  const maxRows =
    typeof window === "undefined"
      ? Infinity
      : Math.max(
          2,
          Math.floor((window.innerHeight / 2 - TOOLTIP_VERTICAL_SPACING) / TOOLTIP_ROW_HEIGHT)
        );
  const willTruncate = hasTotal && rows.length + (hasHeader ? 1 : 0) + 1 > maxRows;
  const showHeader = hasHeader && (!willTruncate || maxRows >= 3);
  let tooltipRows = rows;

  if (willTruncate) {
    // Tippy can flip above or below the cursor, so fit within the half of the
    // viewport available on the chosen side. The row estimate matches the
    // narrow table's line height and cell padding.
    const detailsToKeep = Math.max(0, maxRows - (showHeader ? 1 : 0) - 2);
    const omittedCount = rows.length - detailsToKeep;
    const omittedRow: Array<Array<string | string[]>> = [
      [
        [
          `${omittedCount} earlier rows omitted`,
          "has-text-grey has-text-centered",
          columnCount.toString()
        ]
      ]
    ];
    tooltipRows = [...omittedRow, ...(detailsToKeep > 0 ? rows.slice(-detailsToKeep) : [])];
  }

  const tableRows = [...tooltipRows];

  if (total !== undefined) {
    const totalRow: Array<string | string[]> = [
      ["Total", "has-text-weight-bold"],
      [total, "has-text-weight-bold has-text-right"]
    ];

    for (let i = 2; i < columnCount; i++) {
      totalRow.unshift("");
    }

    tableRows.push(totalRow);
  }

  if (showHeader && header !== undefined && tableRows.length > 0) {
    const headerRow: Array<string | string[]> = [
      [header, "has-text-weight-bold has-text-centered", columnCount.toString()]
    ];
    tableRows.unshift(headerRow);
  }

  const trs = tableRows
    .map((r) => {
      const cells = r
        .map((c) => {
          if (typeof c == "string") {
            return `<td>${c}</td>`;
          } else {
            if (c.length == 3) {
              return `<td class='${c[1]}' colspan='${c[2]}'>${c[0]}</td>`;
            }
            return `<td class='${c[1]}'>${c[0]}</td>`;
          }
        })
        .join("\n");

      return `<tr>${cells}</tr>`;
    })
    .join("\n");
  return `<table class='table is-narrow is-size-7 popup-table'><tbody>${trs}</tbody></table>`;
}
