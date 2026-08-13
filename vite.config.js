import { sveltekit } from "@sveltejs/kit/vite";
import { nodePolyfills } from "vite-plugin-node-polyfills";

/** @type {import('vite').UserConfig} */
const config = {
  cacheDir: "node_modules/.vite",
  css: {
    preprocessorOptions: {
      sass: {
        api: "modern-compiler",
      },
      scss: {
        api: "modern-compiler",
      },
    },
  },
  build: {
    target: "es2021",
  },
  plugins: [
    sveltekit(),
    // xlsx-populate uses Buffer when the lazy-loaded import route is opened.
    nodePolyfills({ include: ["buffer"], globals: { Buffer: true } }),
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
