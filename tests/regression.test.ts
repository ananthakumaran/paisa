import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { once } from "node:events";
import path from "path";
import { describe, test } from "node:test";
import waitPort from "wait-port";
import fs from "fs";
import axios from "axios";
import { diffString } from "json-diff";
import type { LedgerFile } from "../src/lib/utils";

const fixture = "tests/fixture";
const taxFixture = "tests/tax";
const port = 5700;
axios.defaults.baseURL = `http://localhost:${port}`;

function updateConfig(dir: string, from: string, to: string) {
  const filename = path.join(dir, "paisa.yaml");
  let config = fs.readFileSync(filename).toString();
  config = config.replace(from, to);
  fs.writeFileSync(filename, config);
}

async function recordAndVerify(dir: string, route: string, name: string) {
  const { data: data } = await axios.get(route);

  const filename = path.join(dir, name + ".json");
  if (fs.existsSync(filename) && process.env["REGENERATE"] !== "true") {
    const current = JSON.parse(fs.readFileSync(filename).toString());
    const diff = diffString(data, current, {
      excludeKeys: ["id", "transaction_id", "endLine", "transaction_end_line"]
    });

    if (diff != "") {
      assert.fail(diff);
    }
    return;
  }
  fs.writeFileSync(filename, JSON.stringify(data, null, 2));
}

async function syncJournal() {
  const {
    data: { success }
  } = await axios.post("/api/sync", { journal: true });
  assert.strictEqual(success, true);
}

async function verifyApi(dir: string) {
  await syncJournal();

  const {
    data: { config }
  } = await axios.get("/api/config");
  const extension = path.extname(config.journal_path);
  for (const name of ["expenses", "expenses.txt"]) {
    const { data } = await axios.post("/api/editor/save", {
      name,
      content: "",
      operation: "create"
    });
    assert.strictEqual(data.saved, false);
    assert.strictEqual(data.message, `File name must end with ${extension}`);
  }

  await recordAndVerify(dir, "/api/dashboard", "dashboard");
  await recordAndVerify(dir, "/api/cash_flow", "cash_flow");
  await recordAndVerify(dir, "/api/income_statement", "income_statement");
  await recordAndVerify(dir, "/api/expense", "expense");
  await recordAndVerify(dir, "/api/recurring", "recurring");
  await recordAndVerify(dir, "/api/budget", "budget");
  await recordAndVerify(dir, "/api/assets/balance", "assets_balance");
  await recordAndVerify(dir, "/api/networth", "networth");
  await recordAndVerify(dir, "/api/investment", "investment");
  await recordAndVerify(dir, "/api/gain", "gain");
  await recordAndVerify(dir, "/api/allocation", "allocation");
  await recordAndVerify(dir, "/api/liabilities/balance", "liabilities_balance");
  await recordAndVerify(dir, "/api/liabilities/repayment", "liabilities_repayment");
  await recordAndVerify(dir, "/api/liabilities/interest", "liabilities_interest");
  await recordAndVerify(dir, "/api/credit_cards", "credit_cards");
  await recordAndVerify(dir, "/api/income", "income");
  await recordAndVerify(dir, "/api/transaction", "transaction");
  await recordAndVerify(dir, "/api/editor/files", "files");
  await recordAndVerify(dir, "/api/ledger", "ledger");
  await recordAndVerify(dir, "/api/price", "price");
  await recordAndVerify(dir, "/api/diagnosis", "diagnosis");
  await recordAndVerify(dir, "/api/config", "config");

  const {
    data: { files }
  } = await axios.get<{ files: LedgerFile[] }>("/api/editor/files");
  for (const file of files) {
    const { data } = await axios.post("/api/editor/validate", file);
    assert.deepStrictEqual(data.errors, [], `Validation failed for ${file.name}`);
  }

  const mainFile = files.find((file) => file.name === path.basename(config.journal_path));
  const { data: unnamed } = await axios.post("/api/editor/validate", { ...mainFile, name: "" });
  assert.deepStrictEqual(
    unnamed.errors,
    [],
    "Unnamed previews must use the main journal directory"
  );

  for (const name of [`../outside${extension}`, `nested/../../outside${extension}`, "."]) {
    const { data } = await axios.post("/api/editor/validate", { name, content: "" });
    assert.strictEqual(data.errors[0].message, "Invalid file name");
  }
}

async function verifyTaxApi(dir: string) {
  await syncJournal();

  await recordAndVerify(dir, "/api/capital_gains", "capital_gains");
  await recordAndVerify(dir, "/api/harvest", "harvest");
}

function readTaxSummary(dir: string) {
  const read = (name: string) =>
    JSON.parse(fs.readFileSync(path.join(taxFixture, dir, name + ".json")).toString());

  const { capital_gains } = read("capital_gains");
  for (const capitalGain of Object.values<any>(capital_gains)) {
    // the number of postings per sale could vary
    Object.values<any>(capitalGain.fy).forEach((fy) => delete fy.posting_pairs);
  }
  return { capital_gains, harvest: read("harvest") };
}

async function wait() {
  try {
    await waitPort({ port: port, output: "silent" });
  } catch (e) {
    // ignore
  }
}

async function check(directory: string, verify: (directory: string) => Promise<void>) {
  const child = spawn("./paisa", [
    "--config",
    path.join(directory, "paisa.yaml"),
    "--port",
    port.toString(),
    "--now",
    "2022-02-07",
    "serve"
  ]);
  try {
    await wait();
    await verify(directory);
  } finally {
    if (child.exitCode === null) {
      child.kill();
      await once(child, "exit");
    }
  }
}

describe("regression", () => {
  fs.readdirSync(fixture).forEach((dir) => {
    test(dir, async () => {
      const directory = path.join(fixture, dir);
      await check(directory, verifyApi);
    });
  });
});

describe("tax regression", () => {
  const dirs = fs.readdirSync(taxFixture);

  dirs.forEach((dir) => {
    test(dir, async () => {
      await check(path.join(taxFixture, dir), verifyTaxApi);
    });
  });

  test("same result irrespective of the ledger cli", () => {
    const [expected, ...rest] = dirs.map(readTaxSummary);
    rest.forEach((actual) => assert.equal(diffString(expected, actual, { precision: 6 }), ""));
  });
});
