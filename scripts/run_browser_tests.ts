async function waitForPort(port: number, timeout = 120_000) {
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

const server = new Deno.Command(Deno.execPath(), {
  args: [
    "run",
    "--allow-env",
    "--allow-net",
    "--allow-read",
    "--allow-run",
    "--allow-write",
    "scripts/test_server.ts",
  ],
  stdin: "null",
  stdout: Deno.env.get("PAISA_TEST_SERVER_LOG") ? "inherit" : "null",
  stderr: Deno.env.get("PAISA_TEST_SERVER_LOG") ? "inherit" : "null",
}).spawn();

try {
  await waitForPort(5173);
  const status = await new Deno.Command(Deno.execPath(), {
    args: ["run", "-A", "npm:playwright@1.61.1", "test", ...Deno.args],
    stdin: "inherit",
    stdout: "inherit",
    stderr: "inherit",
  }).spawn().status;
  Deno.exitCode = status.code;
} finally {
  try {
    server.kill("SIGTERM");
  } catch (error) {
    if (!(error instanceof Deno.errors.NotFound)) {
      console.error("Failed to stop browser test server", error);
    }
  }
  await server.status;
}
