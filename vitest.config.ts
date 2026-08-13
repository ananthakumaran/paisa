import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [svelte()],
  resolve: {
    alias: {
      "$app/navigation":
        new URL("./src/test/navigation.ts", import.meta.url).pathname,
      $lib: new URL("./src/lib", import.meta.url).pathname,
    },
  },
  test: {
    environment: "happy-dom",
    include: ["src/**/*.component.test.ts"],
    setupFiles: ["./src/test/setup.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "html", "lcov", "json-summary"],
      reportsDirectory: "coverage/component",
      include: ["src/**/*.{ts,svelte}"],
      exclude: [
        "src/**/*.test.ts",
        "src/**/*.d.ts",
        "src/**/parser.ts",
        "src/**/parser.terms.ts",
      ],
      thresholds: { lines: 60, statements: 60, functions: 60, branches: 60 },
    },
  },
});
