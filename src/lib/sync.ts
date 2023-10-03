import * as toast from "bulma-toast";
import { ajax } from "./utils";

export async function sync(request: Record<string, any>) {
  const { success, message } = await ajax("/api/sync", {
    method: "POST",
    body: JSON.stringify(request)
  });

  if (!success) {
    toast.toast({
      message: `<b>Failed to sync</b>\n${message}`,
      type: "is-danger",
      duration: 10000
    });
  }
}
