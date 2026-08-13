import { sveltekit } from "@sveltejs/kit/vite";
import { nodePolyfills } from "vite-plugin-node-polyfills";

/** @type {import('vite').UserConfig} */
const config = {
  cacheDir: "node_modules/.vite",
  build: {
    target: "es2021",
  },
  plugins: [
    sveltekit(),
    nodePolyfills({
      globals: {
        Buffer: true,
      },
    }),
  ],
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:7500",
      },
    },
    fs: {
      allow: ["./fonts"],
    },
  },
};

export default config;
