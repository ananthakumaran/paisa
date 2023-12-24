import Papa from "papaparse";
import * as XLSX from "xlsx";
import _ from "lodash";
import { format } from "./journal";
import { pdf2array } from "./pdf";

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
  options: { reverse?: boolean } = {}
) {
  const output: string[] = [];
  _.each(rows, (row) => {
    const rendered = _.trim(template(_.assign({ ROW: row, SHEET: rows }, COLUMN_REFS)));
    if (!_.isEmpty(rendered)) {
      output.push(rendered);
    }
  });
  if (options.reverse) {
    output.reverse();
  }
  return format(output.join("\n\n"));
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
  const sheet = XLSX.read(buffer, { type: "binary" });
  const json = XLSX.utils.sheet_to_json<string[]>(sheet.Sheets[sheet.SheetNames[0]], {
    header: 1,
    blankrows: false,
    rawNumbers: false
  });
  return { data: json };
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
