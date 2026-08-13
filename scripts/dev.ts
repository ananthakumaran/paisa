const includeFrontend = Deno.args.includes("--frontend");
const fixedDate = Deno.args.includes("--now");
const ignoredDirectories = new Set([
  ".git",
  ".svelte-kit",
  "graphify-out",
  "node_modules",
  "web/static",
]);

let backend: Deno.ChildProcess | undefined;
let frontend: Deno.ChildProcess | undefined;
const backendBinary = await Deno.makeTempFile({
  prefix: "paisa-dev-",
  suffix: Deno.build.os === "windows" ? ".exe" : undefined,
});
let restartTimer: ReturnType<typeof setTimeout> | undefined;
let restarting = false;
let shuttingDown = false;

async function startBackend() {
  const buildStatus = await new Deno.Command("go", {
    args: ["build", "-o", backendBinary, "."],
    stdin: "inherit",
    stdout: "inherit",
    stderr: "inherit",
  }).spawn().status;
  if (!buildStatus.success) {
    await shutdown(buildStatus.code || 1);
    return;
  }

  const args = ["serve"];
  if (fixedDate) args.push("--now", "2022-02-07");

  backend = new Deno.Command(backendBinary, {
    args,
    env: fixedDate ? { TZ: "UTC" } : undefined,
    stdin: "inherit",
    stdout: "inherit",
    stderr: "inherit",
  }).spawn();
  const current = backend;
  current.status.then((status) => {
    if (!restarting && !shuttingDown && backend === current) {
      void shutdown(status.code || 1);
    }
  });
}

function startFrontend() {
  frontend = new Deno.Command(Deno.execPath(), {
    args: ["task", "dev"],
    stdin: "inherit",
    stdout: "inherit",
    stderr: "inherit",
  }).spawn();
  frontend.status.then((status) => {
    if (!shuttingDown) void shutdown(status.code || 1);
  });
}

async function stop(child: Deno.ChildProcess | undefined) {
  if (!child) return;
  try {
    child.kill("SIGTERM");
  } catch (error) {
    if (!(error instanceof Deno.errors.NotFound)) throw error;
  }
  await child.status;
}

async function restartBackend() {
  restarting = true;
  await stop(backend);
  await startBackend();
  restarting = false;
}

async function shutdown(code: number) {
  if (shuttingDown) return;
  shuttingDown = true;
  if (restartTimer !== undefined) clearTimeout(restartTimer);
  await Promise.allSettled([stop(backend), stop(frontend)]);
  await Deno.remove(backendBinary).catch((error) => {
    if (!(error instanceof Deno.errors.NotFound)) throw error;
  });
  Deno.exit(code);
}

function shouldRestart(paths: string[]) {
  return paths.some((path) => {
    const normalized = path.replaceAll("\\", "/");
    if (
      [...ignoredDirectories].some((dir) =>
        normalized.includes(`/${dir}/`) || normalized.endsWith(`/${dir}`)
      )
    ) return false;
    return normalized.endsWith(".go") || normalized.endsWith(".json");
  });
}

const terminationSignals = Deno.build.os === "windows"
  ? (["SIGINT"] as const)
  : (["SIGINT", "SIGTERM"] as const);
for (const signal of terminationSignals) {
  Deno.addSignalListener(signal, () => void shutdown(0));
}

if (includeFrontend) startFrontend();
await startBackend();

for await (const event of Deno.watchFs(".")) {
  if (shuttingDown || !shouldRestart(event.paths)) continue;
  if (restartTimer !== undefined) clearTimeout(restartTimer);
  restartTimer = setTimeout(() => {
    restartTimer = undefined;
    void restartBackend();
  }, 2_000);
}
