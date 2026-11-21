/* global document$ */

(() => {
  function syncTheme() {
    // Read Material's resolved palette, not the saved automatic preference.
    const theme =
      document.body.getAttribute("data-md-color-scheme") === "slate" ? "transparent_dark" : "light";
    const script = document.querySelector('script[src="https://giscus.app/client.js"]');
    if (script) script.setAttribute("data-theme", theme);

    const frame = document.querySelector(".giscus-frame");
    if (frame && frame.contentWindow) {
      frame.contentWindow.postMessage({ giscus: { setConfig: { theme } } }, "https://giscus.app");
    }
  }

  // Material updates this attribute for both manual and system theme changes.
  new MutationObserver(syncTheme).observe(document.body, {
    attributes: true,
    attributeFilter: ["data-md-color-scheme"]
  });

  // A lazy-loaded iframe may not have been ready for an earlier theme update.
  document.addEventListener(
    "load",
    (event) => {
      if (event.target instanceof HTMLIFrameElement && event.target.matches(".giscus-frame")) {
        syncTheme();
      }
    },
    true
  );

  // Emits on initial load and after each Material instant-navigation transition.
  document$.subscribe(syncTheme);
})();
