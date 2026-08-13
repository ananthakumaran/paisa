import { expandGlob } from "@std/fs";

async function run(args: string[]): Promise<void> {
  const command = new Deno.Command(Deno.execPath(), {
    args,
    stdin: "inherit",
    stdout: "inherit",
    stderr: "inherit",
  });
  const status = await command.spawn().status;
  if (!status.success) Deno.exit(status.code);
}

async function toolingFiles(): Promise<string[]> {
  const patterns = [
    "scripts/*.ts",
    "*.config.ts",
    "tests/browser/*.ts",
  ];
  const files = new Set<string>();

  for (const pattern of patterns) {
    for await (const entry of expandGlob(pattern, { globstar: false })) {
      if (entry.isFile) files.add(entry.path);
    }
  }

  return [...files].sort();
}

await run(["check", ...await toolingFiles()]);

if (!Deno.args.includes("--tooling")) {
  await run(["run", "-A", "npm:@sveltejs/kit", "sync"]);
  await run([
    "run",
    "-A",
    "npm:svelte-check",
    "--tsconfig",
    "./tsconfig.json",
  ]);
}
