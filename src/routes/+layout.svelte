<script lang="ts">
  import { afterNavigate, beforeNavigate } from "$app/navigation";
  import { followCursor, type Instance, delegate } from "tippy.js";
  import _ from "lodash";
  import Spinner from "$lib/components/Spinner.svelte";
  import Navbar from "$lib/components/Navbar.svelte";

  let isBurger: boolean = null;
  let tippyInstances: Instance[] = [];

  beforeNavigate(() => {
    tippyInstances.forEach((t) => t.destroy());
  });

  afterNavigate(() => {
    isBurger = null;
    tippyInstances = delegate("section,nav", {
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
