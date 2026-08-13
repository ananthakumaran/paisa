import { join } from "@std/path";
import { describe, it as test } from "@std/testing/bdd";
import { expect } from "@std/expect";
import axios from "axios";
import { diffString } from "json-diff";

const fixture = "tests/fixture";
const port = 5700;
axios.defaults.baseURL = `http://localhost:${port}`;

function updateConfig(dir: string, from: string, to: string) {
  const filename = join(dir, "paisa.yaml");
  let config = Deno.readTextFileSync(filename);
  config = config.replace(from, to);
  Deno.writeTextFileSync(filename, config);
}

async function recordAndVerify(dir: string, route: string, name: string) {
  const { data: data } = await axios.get(route);

  const filename = join(dir, name + ".json");
  let exists = true;
  try {
    Deno.statSync(filename);
  } catch (error) {
    if (error instanceof Deno.errors.NotFound) exists = false;
    else throw error;
  }
  if (exists && Deno.env.get("REGENERATE") !== "true") {
    const current = JSON.parse(Deno.readTextFileSync(filename));
    const diff = diffString(data, current, {
      excludeKeys: ["id", "transaction_id", "endLine", "transaction_end_line"],
    });

    if (diff != "") {
      expect(diff).toBe("");
    }
  }
  Deno.writeTextFileSync(filename, JSON.stringify(data, null, 2));
}

async function verifyApi(dir: string) {
  const {
    data: { success },
  } = await axios.post("/api/sync", { journal: true });
  expect(success).toBe(true);

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
  await recordAndVerify(
    dir,
    "/api/liabilities/repayment",
    "liabilities_repayment",
  );
  await recordAndVerify(
    dir,
    "/api/liabilities/interest",
    "liabilities_interest",
  );
  await recordAndVerify(dir, "/api/income", "income");
  await recordAndVerify(dir, "/api/transaction", "transaction");
  await recordAndVerify(dir, "/api/editor/files", "files");
  await recordAndVerify(dir, "/api/ledger", "ledger");
  await recordAndVerify(dir, "/api/price", "price");
  await recordAndVerify(dir, "/api/diagnosis", "diagnosis");
  await recordAndVerify(dir, "/api/config", "config");
}

async function waitForPort(timeout = 10_000) {
  const deadline = Date.now() + timeout;
  while (Date.now() < deadline) {
    try {
      const connection = await Deno.connect({
        hostname: "127.0.0.1",
        port,
      });
      connection.close();
      return;
    } catch (error) {
      if (!(error instanceof Deno.errors.ConnectionRefused)) throw error;
      await new Promise((resolve) => setTimeout(resolve, 100));
    }
  }
  throw new Error(`Timed out waiting for port ${port}`);
}

async function check(directory: string) {
  const command = new Deno.Command("./paisa", {
    args: [
      "--config",
      join(directory, "paisa.yaml"),
      "--port",
      port.toString(),
      "--now",
      "2022-02-07",
      "serve",
    ],
  });
  const child = command.spawn();
  try {
    await waitForPort();
    await verifyApi(directory);
  } finally {
    child.kill("SIGTERM");
    await child.status;
  }
}

describe("regression", () => {
  Array.from(Deno.readDirSync(fixture)).forEach(({ name: dir }) => {
    test(dir, async () => {
      const directory = join(fixture, dir);
      await check(directory);
    });
  });
});
