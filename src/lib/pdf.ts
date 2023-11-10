import * as pdfjs from "pdfjs-dist";
import type { TextItem } from "pdfjs-dist/types/src/display/api";

export type TextItemWithPosition = TextItem & {
  x: number;
  y: number;
};

export interface Row {
  page: number;
  rowNumber: number;
  y: number;
  xs: number[];
  items: TextItemWithPosition[];
}

/**
 * Transform an (x, y) coordinate by a pdf transformation matrix.
 * @param x
 * @param y
 * @param transform
 * @private
 */
function _transform(x: number, y: number, transform: number[]) {
  // [x', y', 1] = [x, y, 1] * [a, b, 0]
  //                           [c, d, 0]
  //                           [e, f, 1]
  const [a, b, c, d, e, f] = transform;
  const xt = x * a + y * c + e;
  const yt = x * b + y * d + f;
  return [xt, yt];
}

const WORD_SPACE_TOLERANCE = 0.5;

function makeRow(cells: TextItemWithPosition[]): string[] {
  const row = [];
  let lastCell: TextItemWithPosition;
  for (const cell of cells) {
    if (
      lastCell !== undefined &&
      Math.abs(cell.x - (lastCell.x + lastCell.width)) < WORD_SPACE_TOLERANCE
    ) {
      row[row.length - 1] += cell.str;
    } else {
      row.push(cell.str);
    }
    lastCell = cell;
  }
  return row;
}

/**
 * Loads a PDF file and returns text values arranged into a
 * 2d array.
 *
 * @param data
 */
export async function pdf2array(data: ArrayBuffer): Promise<string[][]> {
  const loader = await pdfjs.getDocument(data);
  loader.onPassword = (cb: any) => {
    const password = prompt("Please enter the password to open this PDF file.");
    cb(password);
  };
  const doc = await loader.promise;

  const rows: Row[] = [];
  let currentRow: Row = undefined;

  for (let i = 0; i < doc.numPages; ++i) {
    const page = await doc.getPage(i + 1);
    const text = await page.getTextContent();

    // Start a new row for this page
    if (currentRow !== undefined) {
      rows.push(currentRow);
      currentRow = undefined;
    }

    // Get the position of each item in page space first and remove any zero size items
    const items: TextItemWithPosition[] = (text.items as TextItem[])
      .filter((item) => item.width > 0 && item.height > 0)
      .map((item) => {
        const [left, top] = _transform(0, 0, item.transform);
        return {
          ...item,
          x: left,
          y: top
        };
      });

    // If there are no text items on this page then skip to the next page
    if (items.length === 0) {
      continue;
    }

    // Find the minimum height of any element. We will use this to determine the
    // tolerance for deciding if two items are on the same line or not.
    const minHeight = Math.max(Math.min(...items.map((item) => item.height)), 0.001);
    const yTolerance = minHeight / 2;

    // Sort the items by x and y positions
    items.sort((a, b) => {
      if (a.y >= b.y + yTolerance) return -1;
      if (a.y < b.y - yTolerance) return 1;
      if (a.x < b.x) return -1;
      if (a.x > b.x) return 1;
      return 0;
    });

    // Build a list of rows
    for (const item of items) {
      // Check if this item starts a new row
      if (
        currentRow === undefined ||
        item.y < currentRow.y - yTolerance ||
        item.y >= currentRow.y + yTolerance
      ) {
        // Add the current row to the list of rows
        if (currentRow !== undefined) {
          rows.push(currentRow);
          currentRow = undefined;
        }

        // And start a new row
        currentRow = {
          page: i,
          rowNumber: rows.length,
          y: item.y,
          xs: [item.x],
          items: [item]
        };
      } else {
        // Else add to the current row
        currentRow.xs.push(item.x);
        currentRow.items.push(item);
      }
    }
  }

  // Add the final row if there is one
  if (currentRow !== undefined) {
    rows.push(currentRow);
  }

  return rows.map((row) => makeRow(row.items));
}
