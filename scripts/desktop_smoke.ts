import { basename, dirname, join } from "@std/path";

const configured = Deno.env.get("PAISA_DESKTOP_BINARY");
if (!configured) {
  throw new Error(
    "PAISA_DESKTOP_BINARY must point to the packaged Wails executable",
  );
}
const executable = Deno.realPathSync(configured);
const root = new URL("../", import.meta.url).pathname;
const fixture = await Deno.makeTempDir({ prefix: "paisa-desktop-" });
await Deno.copyFile(
  join(root, "tests/fixture/inr/main.ledger"),
  join(fixture, "main.ledger"),
);
await Deno.copyFile(
  join(root, "tests/fixture/inr/paisa.yaml"),
  join(fixture, "paisa.yaml"),
);

const child = new Deno.Command(executable, {
  cwd: fixture,
  env: {
    PAISA_CONFIG: join(fixture, "paisa.yaml"),
    TZ: "UTC",
    PAISA_GPU_POLICY: "never",
  },
  stdout: "piped",
  stderr: "piped",
}).spawn();

try {
  const result = await Promise.race([
    child.status.then((status) => ({ exited: true, status })),
    new Promise<{ exited: false }>((resolve) =>
      setTimeout(() => resolve({ exited: false }), 10_000)
    ),
  ]);
  if (result.exited) {
    throw new Error(
      `${basename(executable)} exited early with ${result.status.code}`,
    );
  }
  const database = join(fixture, "paisa.db");
  try {
    await Deno.stat(database);
  } catch (error) {
    if (error instanceof Deno.errors.NotFound) {
      throw new Error(
        `Desktop app stayed alive but did not create ${database}`,
      );
    }
    throw error;
  }
  console.log(
    `Desktop smoke passed for ${basename(dirname(executable))}/${
      basename(executable)
    }`,
  );
} finally {
  try {
    child.kill("SIGTERM");
  } catch (error) {
    if (!(error instanceof Deno.errors.NotFound)) {
      console.error("Failed to stop desktop app", error);
    }
  }
  await child.status;
  await Deno.remove(fixture, { recursive: true });
}
