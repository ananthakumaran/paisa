import Papa from "papaparse";
import * as XLSX from "xlsx";
import _ from "lodash";
import { format } from "./journal";
import { pdf2array } from "./pdf";
import * as XlsxPopulate from "xlsx-populate";

interface Result {
  data: string[][];
  error?: string;
}

export function parse(file: File): Promise<Result> {
  let extension = file.name.split(".").pop();
  extension = extension?.toLowerCase();
  if (extension === "csv" || extension === "txt") {
    return parseCSV(file);
  } else if (extension === "xlsx" || extension === "xls") {
    return parseXLSX(file);
  } else if (extension === "pdf") {
    return parsePDF(file);
  }
  throw new Error(`Unsupported file type ${extension}`);
}

export function asRows(result: Result): Array<Record<string, any>> {
  return _.map(result.data, (row, i) => {
    return _.chain(row)
      .map((cell, j) => {
        return [String.fromCharCode(65 + j), cell];
      })
      .concat([["index", i as any]])
      .fromPairs()
      .value();
  });
}

const COLUMN_REFS = _.chain(_.range(65, 90))
  .map((i) => String.fromCharCode(i))
  .map((a) => [a, a])
  .fromPairs()
  .value();

export function render(
  rows: Array<Record<string, any>>,
  template: Handlebars.TemplateDelegate,
  options: { reverse?: boolean; trim?: boolean } = {}
) {
  const output: string[] = [];
  _.each(rows, (row) => {
    let rendered = template(_.assign({ ROW: row, SHEET: rows }, COLUMN_REFS));
    if (options.trim) {
      rendered = _.trim(rendered);
    }
    if (!_.isEmpty(rendered)) {
      output.push(rendered);
    }
  });
  if (options.reverse) {
    output.reverse();
  }

  if (options.trim) {
    return format(output.join("\n\n"));
  } else {
    return format(output.join(""));
  }
}

function parseCSV(file: File): Promise<Result> {
  return new Promise((resolve, reject) => {
    Papa.parse<string[]>(file, {
      skipEmptyLines: true,
      complete: function (results) {
        resolve(results);
      },
      error: function (error) {
        reject(error);
      },
      delimitersToGuess: [",", "\t", "|", ";", Papa.RECORD_SEP, Papa.UNIT_SEP, "^"]
    });
  });
}

async function parseXLSX(file: File): Promise<Result> {
  const buffer = await readFile(file);
  try {
    const sheet = XLSX.read(buffer, { type: "binary" });
    const json = XLSX.utils.sheet_to_json<string[]>(sheet.Sheets[sheet.SheetNames[0]], {
      header: 1,
      blankrows: false,
      rawNumbers: false
    });
    return { data: json };
  } catch (e) {
    if (/password-protected/.test(e.message)) {
      const password = prompt(
        "Please enter the password to open this XLSX file. Press cancel to exit."
      );
      if (password === null) {
        return { data: [], error: "Password required." };
      }

      try {
        const workbook = await XlsxPopulate.fromDataAsync(buffer, { password });
        const sheet = workbook.sheet(0);
        if (sheet) {
          let json = sheet.usedRange().value();
          json = _.map(json, (row) => {
            return _.map(row, (cell) => {
              if (cell) {
                return cell.toString();
              }
              return "";
            });
          });

          return { data: json };
        }
      } catch (e) {
        // follow through to the error below
      }

      return { data: [], error: "Unable to parse Password protected XLSX" };
    }
    throw e;
  }
}

async function parsePDF(file: File): Promise<Result> {
  const buffer = await readFile(file);
  const array = await pdf2array(buffer);
  return { data: array };
}

function readFile(file: File): Promise<ArrayBuffer> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = (event) => {
      resolve(event.target.result as ArrayBuffer);
    };
    reader.onerror = (event) => {
      reject(event);
    };
    reader.readAsArrayBuffer(file);
  });
}
