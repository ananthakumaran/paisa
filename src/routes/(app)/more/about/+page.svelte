<script lang="ts">
  import Logo from "$lib/components/Logo.svelte";
  import { _ as t } from "$lib/i18n";
  import { get } from "svelte/store";
  const tt = get(t);
  const links = [
    {
      name: tt("page.more.about.chat"),
      href: "https://matrix.to/#/#paisa:matrix.org",
      icon: "fa-solid fa-headset"
    },
    {
      name: tt("page.more.about.issue"),
      href: "https://github.com/ananthakumaran/paisa/issues",
      icon: "fas fa-bug"
    },
    {
      name: tt("page.more.about.discussions"),
      href: "https://github.com/ananthakumaran/paisa/discussions",
      icon: "fa-regular fa-comments"
    },
    {
      name: tt("page.more.about.source_code"),
      href: "https://github.com/ananthakumaran/paisa",
      icon: "fa-solid fa-code"
    },
    {
      name: tt("page.more.about.documentation"),
      href: "https://paisa.fyi",
      icon: "fa-solid fa-book"
    },
    {
      name: tt("page.more.about.releases"),
      href: "https://github.com/ananthakumaran/paisa/releases",
      icon: "fa-solid fa-download"
    },
    {
      name: tt("page.more.about.demo"),
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
          <div><Logo size={128} /></div>
          <div class="is-size-3 is-primary-color">Paisa</div>
          <div>
            {$t("page.more.about.version")}: <b>0.7.4</b>
          </div>
        </div>

        <div class="box px-3 mx-auto" style="max-width: 400px;">
          <h3 class="is-size-5 mb-1">{$t("page.more.about.links")}</h3>
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
