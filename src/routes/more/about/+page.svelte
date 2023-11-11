<script lang="ts">
  import Logo from "$lib/components/Logo.svelte";
  const links = [
    { name: "Chat", href: "https://matrix.to/#/#paisa:matrix.org", icon: "fa-solid fa-headset" },
    { name: "Issue", href: "https://github.com/ananthakumaran/paisa/issues", icon: "fas fa-bug" },
    {
      name: "Discussions",
      href: "https://github.com/ananthakumaran/paisa/discussions",
      icon: "fa-regular fa-comments"
    },
    {
      name: "Source Code",
      href: "https://github.com/ananthakumaran/paisa",
      icon: "fa-solid fa-code"
    },
    { name: "Documentation", href: "https://paisa.fyi", icon: "fa-solid fa-book" },
    {
      name: "Releases",
      href: "https://github.com/ananthakumaran/paisa/releases",
      icon: "fa-solid fa-download"
    },
    {
      name: "Demo",
      href: "https://demo.paisa.fyi",
      icon: "fa-solid fa-video"
    }
  ];

  function externalLink(url: string) {
    if (window.runtime) {
      window.runtime.BrowserOpenURL(url);
    } else {
      window.open(url, "_blank");
    }
  }
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box has-text-centered px-3 mx-auto" style="max-width: 400px;">
          <div><Logo size={64} /></div>
          <div class="is-size-3">Paisa</div>
          <div>
            Version: <b>0.5.7</b>
          </div>
        </div>

        <div class="box px-3 mx-auto" style="max-width: 400px;">
          <h3 class="is-size-5 mb-1">Links</h3>
          <ul>
            {#each links as link}
              <li>
                <a href={link.href} on:click|preventDefault={(_e) => externalLink(link.href)}>
                  <span class="icon is-small">
                    <i class={link.icon} />
                  </span>
                  <span>{link.name}</span>
                </a>
              </li>
            {/each}
          </ul>
        </div>
      </div>
    </div>
  </div>
</section>
