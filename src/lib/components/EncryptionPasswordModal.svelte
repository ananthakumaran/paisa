<script lang="ts">
  import Modal from "$lib/components/Modal.svelte";
  import * as toast from "bulma-toast";
  import {
    cancelEncryptionModal,
    encryptionModalOpen,
    encryptionModalMode,
    postEncryptionPassword,
    runEncryptionAfterSubmit
  } from "$lib/encryptionUnlock";

  let password = "";
  let confirmPassword = "";
  let errorMessage = "";
  let submitting = false;

  $: mode = $encryptionModalMode;

  function resetFields() {
    password = "";
    confirmPassword = "";
    errorMessage = "";
  }

  $: if (!$encryptionModalOpen) {
    resetFields();
  }

  async function submit() {
    errorMessage = "";
    if ((mode === "set" || mode === "disable") && password !== confirmPassword) {
      errorMessage = "Passwords do not match";
      return;
    }
    if (!password) {
      errorMessage = "Enter a password";
      return;
    }
    submitting = true;
    try {
      const { ok, message } = await postEncryptionPassword(password);
      if (!ok) {
        errorMessage = message || "Could not set password";
        return;
      }
      resetFields();
      if (mode === "unlock" || mode === "set") {
        toast.toast({ message: "Encryption password saved for this session", type: "is-success" });
      }
      await runEncryptionAfterSubmit();
    } catch (e) {
      errorMessage = e instanceof Error ? e.message : "Request failed";
    } finally {
      submitting = false;
    }
  }
</script>

<Modal
  bind:active={$encryptionModalOpen}
  width="min(480px, 100vw)"
  dismissable={mode === "set" || mode === "disable"}
  onDismiss={mode === "set" || mode === "disable" ? () => cancelEncryptionModal() : undefined}
>
  <span slot="head" let:close>
    <p class="modal-card-title">
      {mode === "unlock"
        ? "Unlock encrypted ledger"
        : mode === "disable"
          ? "Confirm to disable encryption"
          : "Set encryption password"}
    </p>
  </span>
  <div slot="body">
    {#if mode === "unlock"}
      <p class="mb-3">
        Encrypted journal files were found. Enter the password to decrypt them for this session.
      </p>
    {:else if mode === "disable"}
      <p class="mb-3">
        Turning off encryption will decrypt ledger files on disk. Enter your encryption password
        twice to confirm. After saving, the password will be cleared from this session.
      </p>
    {:else}
      <p class="mb-3">
        Choose a password for encrypting your ledger. You will be asked again each time you start
        the app (the password is not stored on disk).
      </p>
    {/if}
    {#if errorMessage}
      <article class="message is-danger mb-3">
        <div class="message-body">{errorMessage}</div>
      </article>
    {/if}
    <div class="field">
      <label class="label" for="enc-password">Password</label>
      <div class="control">
        <input
          id="enc-password"
          class="input"
          type="password"
          autocomplete="off"
          bind:value={password}
          on:keydown={(e) => e.key === "Enter" && submit()}
        />
      </div>
    </div>
    {#if mode === "set" || mode === "disable"}
      <div class="field">
        <label class="label" for="enc-password2">Confirm password</label>
        <div class="control">
          <input
            id="enc-password2"
            class="input"
            type="password"
            autocomplete="off"
            bind:value={confirmPassword}
            on:keydown={(e) => e.key === "Enter" && submit()}
          />
        </div>
      </div>
    {/if}
  </div>
  <span slot="foot">
    {#if mode === "unlock"}
      <span class="is-size-7 has-text-grey"
        >This window stays open until the password is correct.</span
      >
    {:else}
      <button type="button" class="button is-light" on:click={cancelEncryptionModal}>Cancel</button>
    {/if}
    <button
      type="button"
      class="button is-success"
      class:is-loading={submitting}
      on:click={submit}
      >{mode === "unlock" ? "Unlock" : mode === "disable" ? "Confirm and save" : "Continue"}</button
    >
  </span>
</Modal>
