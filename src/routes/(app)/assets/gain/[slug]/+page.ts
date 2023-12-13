import type { PageLoad } from "./$types";

export const load = (async ({ params }) => {
  return {
    name: params.slug
  };
}) satisfies PageLoad;
