import _ from "lodash";
import { get, writable } from "svelte/store";
import { tokenKey } from "$lib/utils";

export type EncryptionModalMode = "unlock" | "set" | "disable";

export const encryptionModalOpen = writable(false);
export const encryptionModalMode = writable<EncryptionModalMode>("unlock");

/** If set, invoked after a successful POST /api/encryption/password */
const encryptionAfterSubmit = writable<null | (() => void | Promise<void>)>(null);

export function openEncryptionModal(
  mode: EncryptionModalMode,
  afterSubmit?: () => void | Promise<void>
) {
  encryptionModalMode.set(mode);
  encryptionAfterSubmit.set(afterSubmit ?? null);
  encryptionModalOpen.set(true);
}

export async function runEncryptionAfterSubmit() {
  const fn = get(encryptionAfterSubmit);
  encryptionAfterSubmit.set(null);
  encryptionModalOpen.set(false);
  if (fn) {
    await fn();
  }
}

/** Cancel without submitting (clears pending save callback). */
export function cancelEncryptionModal() {
  encryptionAfterSubmit.set(null);
  encryptionModalOpen.set(false);
}

export async function postEncryptionPassword(password: string): Promise<{
  ok: boolean;
  message?: string;
}> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  const tok = localStorage.getItem(tokenKey);
  if (!_.isEmpty(tok)) {
    headers["X-Auth"] = tok;
  }
  const res = await fetch("/api/encryption/password", {
    method: "POST",
    headers,
    body: JSON.stringify({ password })
  });
  let body: { success?: boolean; message?: string } = {};
  try {
    body = await res.json();
  } catch {
    /* ignore */
  }
  if (!body.success) {
    return { ok: false, message: body.message };
  }
  return { ok: true };
}
