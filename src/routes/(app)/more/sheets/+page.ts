import { redirect } from "@sveltejs/kit";
import { ajax } from "$lib/utils";
import type { PageLoad } from "./$types";

export const load: PageLoad = async () => {
  const { files } = await ajax("/api/sheets/files");
  if (files.length > 0) {
    redirect(307, `/more/sheets/${files[0].name}`);
  }
};
