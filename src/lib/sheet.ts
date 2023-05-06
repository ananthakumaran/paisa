import Papa from "papaparse";
import * as XLSX from "xlsx";

interface Result {
  data: string[][];
}

export function parse(file: File): Promise<Result> {
  const extension = file.name.split(".").pop();
  if (extension === "csv" || extension === "txt") {
    return parseCSV(file);
  } else if (extension === "xlsx" || extension === "xls") {
    return parseXLSX(file);
  }
  throw new Error("Unsupported file type");
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
