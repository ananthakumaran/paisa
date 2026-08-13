import { join } from "@std/path";

const root = new URL("../", import.meta.url).pathname;
const fixture = await Deno.makeTempDir({ prefix: "paisa-browser-" });
const binary = join(
  fixture,
  Deno.build.os === "windows" ? "paisa.exe" : "paisa",
);
const children: Deno.ChildProcess[] = [];
let stopping = false;

async function run(command: string, args: string[], cwd = root) {
  const status = await new Deno.Command(command, {
    args,
    cwd,
    stdin: "null",
    stdout: "inherit",
    stderr: "inherit",
  }).spawn().status;
  if (!status.success) {
    throw new Error(`${command} failed with exit code ${status.code}`);
  }
}

async function waitForPort(port: number, timeout = 30_000) {
  const deadline = Date.now() + timeout;
  while (Date.now() < deadline) {
    try {
      const connection = await Deno.connect({ hostname: "127.0.0.1", port });
      connection.close();
      return;
    } catch (error) {
      if (!(error instanceof Deno.errors.ConnectionRefused)) throw error;
      await new Promise((resolve) => setTimeout(resolve, 100));
    }
  }
  throw new Error(`Timed out waiting for port ${port}`);
}

async function stop(code: number) {
  if (stopping) return;
  stopping = true;
  for (const child of children) {
    try {
      child.kill("SIGTERM");
    } catch (error) {
      if (!(error instanceof Deno.errors.NotFound)) throw error;
    }
  }
  await Promise.allSettled(children.map((child) => child.status));
  await Deno.remove(fixture, { recursive: true });
  Deno.exit(code);
}

for (
  const signal of Deno.build.os === "windows"
    ? ["SIGINT"] as const
    : ["SIGINT", "SIGTERM"] as const
) {
  Deno.addSignalListener(signal, () => void stop(0));
}

try {
  await Deno.copyFile(
    join(root, "tests/fixture/inr/main.ledger"),
    join(fixture, "main.ledger"),
  );
  await Deno.copyFile(
    join(root, "tests/fixture/inr/paisa.yaml"),
    join(fixture, "paisa.yaml"),
  );
  await run("go", ["build", "-o", binary, "."]);

  const backend = new Deno.Command(binary, {
    args: [
      "--config",
      join(fixture, "paisa.yaml"),
      "serve",
      "--port",
      "7500",
      "--now",
      "2022-02-07",
    ],
    cwd: fixture,
    env: { TZ: "UTC" },
    stdout: "inherit",
    stderr: "inherit",
  }).spawn();
  children.push(backend);
  await waitForPort(7500);

  const frontend = new Deno.Command(Deno.execPath(), {
    args: ["task", "dev"],
    cwd: root,
    stdout: "inherit",
    stderr: "inherit",
  }).spawn();
  children.push(frontend);
  await waitForPort(5173);

  const failed = await Promise.race(children.map((child) => child.status));
  await stop(failed.code || 1);
} catch (error) {
  console.error(error);
  await stop(1);
}
