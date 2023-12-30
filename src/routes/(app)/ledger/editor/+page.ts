import { redirect } from "@sveltejs/kit";
import { ajax } from "$lib/utils";
import type { PageLoad } from "./$types";

export const load: PageLoad = async () => {
  const { files } = await ajax("/api/editor/files");
  redirect(307, `/ledger/editor/${files[0].name}`);
};
