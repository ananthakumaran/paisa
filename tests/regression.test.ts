import { spawn } from "bun";
import path from "path";
import { describe, expect, test } from "bun:test";
import waitPort from "wait-port";
import fs from "fs";
import axios from "axios";
import { diffString } from "json-diff";

const fixture = "tests/fixture";
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
      expect().fail(diff);
    }
  }
  fs.writeFileSync(filename, JSON.stringify(data, null, 2));
}

async function verifyApi(dir: string) {
  const {
    data: { success }
  } = await axios.post("/api/sync", { journal: true });
  expect(success).toBe(true);

  await recordAndVerify(dir, "/api/dashboard", "dashboard");
  await recordAndVerify(dir, "/api/cash_flow", "cash_flow");
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
  await recordAndVerify(dir, "/api/income", "income");
  await recordAndVerify(dir, "/api/transaction", "transaction");
  await recordAndVerify(dir, "/api/editor/files", "files");
  await recordAndVerify(dir, "/api/ledger", "ledger");
  await recordAndVerify(dir, "/api/price", "price");
  await recordAndVerify(dir, "/api/diagnosis", "diagnosis");
}

async function wait() {
  try {
    await waitPort({ port: port, output: "silent" });
  } catch (e) {
    // ignore
  }
}

async function check(directory: string) {
  const process = spawn([
    "./paisa",
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
    await verifyApi(directory);
  } finally {
    process.kill();
    await process.exited;
  }
}

describe("regression", () => {
  fs.readdirSync(fixture).forEach((dir) => {
    test(dir, async () => {
      const directory = path.join(fixture, dir);
      await check(directory);
    });
  });
});
