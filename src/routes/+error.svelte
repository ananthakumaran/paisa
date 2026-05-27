<script lang="ts">
  import { page } from "$app/stores";
  // Aliased to `t` to keep the convention consistent with other shells where
  // lodash's `_` is in scope.
  import { _ as t } from "$lib/i18n";
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-4 is-offset-4 mt-5">
        {#if $page.error.status === 404}
          <article class="message invertable is-danger">
            <div class="message-header">{$t("error.page_not_found.title")}</div>
            <div class="message-body">
              <p>
                {@html $t("error.page_not_found.body", {
                  values: {
                    homeLink: `<a href="/">${$t("error.page_not_found.homepage")}</a>`
                  }
                })}
              </p>
            </div>
          </article>
        {:else}
          <article class="message invertable is-danger">
            <div class="message-header">{$t("error.critical.title")}</div>
            <div class="message-body">
              <p>{$t("error.critical.intro")}</p>
              <p class="mt-2">{$page.error.message}</p>
              {#if $page.error.stack}
                <pre class="mt-2">{$page.error.stack}</pre>
              {/if}
              <p class="mt-5">
                {@html $t("error.critical.report", {
                  values: {
                    link: `<a href="https://github.com/ananthakumaran/paisa/issues">https://github.com/ananthakumaran/paisa/issues</a>`
                  }
                })}
              </p>
            </div>
          </article>
        {/if}
      </div>
    </div>
  </div>
</section>
