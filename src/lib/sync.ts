import * as toast from "bulma-toast";
import _ from "lodash";
import { ajax } from "./utils";

export async function sync(request: Record<string, any>) {
  const {
    success,
    message,
    warnings = []
  } = await ajax("/api/sync", {
    method: "POST",
    body: JSON.stringify(request)
  });

  if (!success) {
    toast.toast({
      message: `<b>Failed to sync</b><br>${_.escape(message).replaceAll("\n", "<br>")}`,
      type: "is-danger",
      duration: 10000
    });
  }

  if (warnings.length > 0) {
    const details = warnings.map((warning) => _.escape(warning)).join("<br>");
    toast.toast({
      message: `<b>Price update warnings</b><br>${details}`,
      type: "is-warning",
      duration: 10000
    });
  }
}
