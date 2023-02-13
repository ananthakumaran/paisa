import { sveltekit } from "@sveltejs/kit/vite";

/** @type {import('vite').UserConfig} */
const config = {
  plugins: [sveltekit()],
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:7500"
      }
    }
  }
};

export default config;
