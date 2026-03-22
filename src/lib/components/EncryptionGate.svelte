<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import * as toast from "bulma-toast";
  import { postEncryptionPassword } from "$lib/encryptionUnlock";

  let password = "";
  let errorMessage = "";
  let submitting = false;

  async function submit() {
    errorMessage = "";
    if (!password) {
      errorMessage = "Enter a password";
      return;
    }
    submitting = true;
    try {
      const { ok, message } = await postEncryptionPassword(password);
      if (!ok) {
        errorMessage = message || "Could not unlock";
        return;
      }
      password = "";
      toast.toast({ message: "Encryption password saved for this session", type: "is-success" });
      await invalidateAll();
    } catch (e) {
      errorMessage = e instanceof Error ? e.message : "Request failed";
    } finally {
      submitting = false;
    }
  }
</script>

<!-- Renders before the rest of the app so /api/* data loads do not run without a session password. -->
<div class="modal is-active">
  <div class="modal-background"></div>
  <div class="modal-card" style="width: min(480px, 100vw)">
    <header class="modal-card-head">
      <p class="modal-card-title">Unlock encrypted ledger</p>
    </header>
    <section class="modal-card-body">
      <p class="mb-3">
        Encrypted journal files were found. Enter the password to decrypt them for this session.
      </p>
      {#if errorMessage}
        <article class="message is-danger mb-3">
          <div class="message-body">{errorMessage}</div>
        </article>
      {/if}
      <div class="field">
        <label class="label" for="enc-gate-password">Password</label>
        <div class="control">
          <input
            id="enc-gate-password"
            class="input"
            type="password"
            autocomplete="off"
            bind:value={password}
            on:keydown={(e) => e.key === "Enter" && submit()}
          />
        </div>
      </div>
    </section>
    <footer class="modal-card-foot">
      <span class="is-size-7 has-text-grey">Enter your password to continue.</span>
      <button
        type="button"
        class="button is-success"
        class:is-loading={submitting}
        on:click={submit}>Unlock</button
      >
    </footer>
  </div>
</div>
