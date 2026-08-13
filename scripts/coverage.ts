import { emptyDir } from "@std/fs";

await emptyDir("coverage/deno");
const unit = await new Deno.Command(Deno.execPath(), {
  args: [
    "test",
    "--no-check",
    "--allow-read",
    "--allow-env",
    "--ignore=src/lib/components",
    "--coverage=coverage/deno",
    "src/",
  ],
  stdin: "inherit",
  stdout: "inherit",
  stderr: "inherit",
}).spawn().status;
if (!unit.success) Deno.exit(unit.code);

const denoReport = await new Deno.Command(Deno.execPath(), {
  args: ["coverage", "coverage/deno", "--lcov", "--output=coverage/deno.lcov"],
  stdout: "inherit",
  stderr: "inherit",
}).spawn().status;
if (!denoReport.success) Deno.exit(denoReport.code);

const component = await new Deno.Command(Deno.execPath(), {
  args: [
    "run",
    "-A",
    "npm:vitest@2.1.9",
    "run",
    "--config",
    "vitest.config.ts",
    "--coverage",
    "--pool=threads",
    "--maxWorkers=1",
    "--minWorkers=1",
  ],
  stdin: "inherit",
  stdout: "inherit",
  stderr: "inherit",
}).spawn().status;
if (!component.success) Deno.exit(component.code);

const parts = ["coverage/deno.lcov", "coverage/component/lcov.info"];
const merged = (await Promise.all(parts.map((path) => Deno.readTextFile(path))))
  .join("\n");
await Deno.writeTextFile("coverage/lcov.info", merged);
console.log("Combined LCOV report: coverage/lcov.info");

const summary = JSON.parse(
  await Deno.readTextFile("coverage/component/coverage-summary.json"),
).total as Record<string, { pct: number }>;
const threshold = 60;
const failed = ["lines", "statements", "functions", "branches"].filter(
  (metric) => summary[metric].pct < threshold,
);
if (failed.length) {
  const message = `Frontend coverage is below ${threshold}%: ${
    failed.map((metric) => `${metric}=${summary[metric].pct}%`).join(", ")
  }`;
  if (Deno.args.includes("--report-only")) {
    console.warn(`Coverage target not enforced yet: ${message}`);
  } else {
    throw new Error(message);
  }
}
