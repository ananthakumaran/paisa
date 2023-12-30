<script lang="ts">
  import { page } from "$app/stores";
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-4 is-offset-4 mt-5">
        {#if $page.error.status === 404}
          <article class="message invertable is-danger">
            <div class="message-header">Page not found</div>
            <div class="message-body">
              <p>
                The page you are looking for does not exist. It may have been moved, or removed
                altogether. Perhaps you can return back to the site's <a href="/">homepage</a> and see
                if you can find what you are looking for.
              </p>
            </div>
          </article>
        {:else}
          <article class="message invertable is-danger">
            <div class="message-header">Something Went Wrong</div>
            <div class="message-body">
              <p>Paisa has encountered a critical error</p>
              <p class="mt-2">{$page.error.message}</p>
              {#if $page.error.stack}
                <pre class="mt-2">{$page.error.stack}</pre>
              {/if}
              <p class="mt-5">
                Please report this issue at <a href="https://github.com/ananthakumaran/paisa/issues"
                  >https://github.com/ananthakumaran/paisa/issues</a
                >. Closing and reopening the app may help.
              </p>
            </div>
          </article>
        {/if}
      </div>
    </div>
  </div>
</section>
