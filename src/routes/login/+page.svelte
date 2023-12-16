<script lang="ts">
  import { goto } from "$app/navigation";
  import Logo from "$lib/components/Logo.svelte";
  import { login } from "$lib/utils";
  import _ from "lodash";
  let username = "";
  let password = "";

  let invalid = false;
  let invalidErrorMessage = "";

  $: loginDisabled = _.isEmpty(username) || _.isEmpty(password);

  async function tryLogin() {
    if (loginDisabled) return;

    const { success, error } = await login(username, password);
    invalid = !success;
    if (success) {
      goto("/");
    } else if (error) {
      invalidErrorMessage = error;
    }
  }
</script>

<section class="section m-0 p-0">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12 p-0">
        <div class="flex justify-center items-center h-screen">
          <div class="box px-5 w-80">
            <div class="flex justify-center items-center mb-2">
              <div class="mt-1 mr-1"><Logo size={32} /></div>
              <div class="is-size-3">
                <a href="https://paisa.fyi" class="is-primary-color">Paisa</a>
              </div>
            </div>
            <form on:submit|preventDefault={tryLogin}>
              <div class="field">
                <label for="" class="label">Username</label>
                <div class="control">
                  <input class="input" type="text" bind:value={username} />
                </div>
              </div>

              <div class="field">
                <label for="" class="label">Password</label>
                <div class="control">
                  <input class="input" type="password" bind:value={password} />
                </div>
                {#if invalid}
                  <p class="help is-danger">{invalidErrorMessage}</p>
                {/if}
              </div>

              <div class="field is-grouped is-grouped-right">
                <div class="control">
                  <button class="button is-link" disabled={loginDisabled}>Login</button>
                </div>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</section>
