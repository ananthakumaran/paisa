import type { HandleClientError } from "@sveltejs/kit";
import * as toast from "bulma-toast";

export const handleError: HandleClientError = async ({ error, status, message }) => {
  let stack = null;
  if (error instanceof Error) {
    stack = error.stack;
  }
  return { message, stack, status, detail: error.toString() };
};

function formatError(error: any) {
  if (error.stack) {
    return error.stack;
  }

  if (error.message) {
    return error.message;
  } else {
    return error.toString();
  }
}

const footer = `
<p class="mt-3">
  Please report this issue at <a href="https://github.com/ananthakumaran/paisa/issues"
    >https://github.com/ananthakumaran/paisa/issues</a
  >. Closing and reopening the app may help.
</p>
`;

function displayError(error: any) {
  const message = formatError(error);
  toast.toast({
    message: `<div class="message invertable is-danger"><div class="message-header">Something Went Wrong</div><div class="message-body">${message}${footer}</div></div>`,
    type: "is-danger",
    dismissible: true,
    pauseOnHover: true,
    duration: 10000,
    position: "center",
    animate: { in: "fadeIn", out: "fadeOut" }
  });
}

window.addEventListener("unhandledrejection", (event) => {
  displayError(event.reason);
});
window.addEventListener("error", (event) => {
  displayError(event.error);
});
