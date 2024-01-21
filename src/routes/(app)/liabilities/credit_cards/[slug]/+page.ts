import type { PageLoad } from "./$types";

export const load = (async ({ params }) => {
  return {
    account: params.slug
  };
}) satisfies PageLoad;
