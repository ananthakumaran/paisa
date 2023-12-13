import type { HandleClientError } from "@sveltejs/kit";

export const handleError: HandleClientError = async ({ error, event }) => {
  console.error("handle error", error, event);
  return { message: "Whoops!" };
};
