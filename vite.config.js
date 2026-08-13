import { sveltekit } from "@sveltejs/kit/vite";

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
  plugins: [sveltekit()],
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
