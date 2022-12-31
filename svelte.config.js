import adapter from "@sveltejs/adapter-static";
import { vitePreprocess } from "@sveltejs/kit/vite";

// https://github.com/sveltejs/kit/issues/3998
const host = process.env["PAISA_HOST"] || "http://localhost:7500";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://kit.svelte.dev/docs/integrations#preprocessors
  // for more information about preprocessors
  preprocess: vitePreprocess(),

  onwarn: (warning, handler) => {
    if (warning.code.startsWith("a11y-")) return;
    handler(warning);
  },

  kit: {
    adapter: adapter({ pages: "web/static", assets: "web/static", out: "web/static" }),
    paths: {
      assets: host + "/static",
      base: ""
    }
  }
};

export default config;
