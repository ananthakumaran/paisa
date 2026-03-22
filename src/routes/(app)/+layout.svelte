<script lang="ts">
  import { afterNavigate, beforeNavigate } from "$app/navigation";
  import { followCursor, delegate, hideAll } from "tippy.js";
  import _ from "lodash";
  import Spinner from "$lib/components/Spinner.svelte";
  import Navbar from "$lib/components/Navbar.svelte";
  import EncryptionGate from "$lib/components/EncryptionGate.svelte";
  import EncryptionPasswordModal from "$lib/components/EncryptionPasswordModal.svelte";
  import { willClearTippy, willRefresh } from "../../store";

  export let data: { encryptionLocked: boolean };

  let isBurger: boolean = null;

  function clearTippy() {
    hideAll();
  }

  function setupTippy() {
    delegate("body", {
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
      popperOptions: {
        modifiers: [
          {
            name: "flip",
            options: {
              fallbackPlacements: ["auto"]
            }
          }
        ]
      },
      plugins: [followCursor]
    });
  }

  willClearTippy.subscribe(clearTippy);
  beforeNavigate(clearTippy);
  willRefresh.subscribe(() => {
    clearTippy();
    setupTippy();
  });

  afterNavigate(() => {
    isBurger = null;
    setupTippy();
  });
</script>

{#if data.encryptionLocked}
  <EncryptionGate />
{:else}
  <EncryptionPasswordModal />

  {#key $willRefresh}
    <Navbar bind:isBurger />

    <Spinner>
      <slot />
    </Spinner>
  {/key}
{/if}
