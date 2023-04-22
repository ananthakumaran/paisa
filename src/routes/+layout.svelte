<script lang="ts">
  import { afterNavigate, beforeNavigate } from "$app/navigation";
  import { followCursor, delegate, hideAll } from "tippy.js";
  import _ from "lodash";
  import Spinner from "$lib/components/Spinner.svelte";
  import Navbar from "$lib/components/Navbar.svelte";
  import { willClearTippy } from "../store";

  let isBurger: boolean = null;

  function clearTippy() {
    hideAll();
  }

  willClearTippy.subscribe(clearTippy);
  beforeNavigate(clearTippy);

  afterNavigate(() => {
    isBurger = null;
    delegate("section,nav", {
      target: "[data-tippy-content]",
      theme: "light",
      onShow: (instance) => {
        const content = instance.reference.getAttribute("data-tippy-content");
        if (!_.isEmpty(content)) {
          instance.setContent(content);
        } else {
          return false;
        }
      },
      maxWidth: "none",
      delay: 0,
      allowHTML: true,
      followCursor: true,
      plugins: [followCursor]
    });
  });
</script>

<Navbar {isBurger} />

<Spinner>
  <slot />
</Spinner>
