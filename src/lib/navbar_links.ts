import _ from "lodash";

export interface Link {
  label: string;
  href: string;
  tag?: string;
  help?: string;
  hide?: boolean;
  dateRangeSelector?: boolean;
  monthPicker?: boolean;
  calendarYearPicker?: boolean;
  maxDepthSelector?: boolean;
  recurringIcons?: boolean;
  children?: Link[];
  disablePreload?: boolean;
}

export interface SelectedLinks {
  selectedLink: Link | null;
  selectedSubLink: Link | null;
  selectedSubSubLink: Link | null;
}

/**
 * Resolve which top-level / sub / sub-sub link in the navbar matches the
 * given URL path.
 *
 * This function MUST be tolerant of paths that have no matching entry
 * in the menu tree. The Tax submenu, for example, is only registered
 * when default_currency is "INR" — but the tax routes themselves still
 * exist at `/more/tax/*` and may be reached via direct navigation or
 * deep links. Previously a missing parent / sub-link caused a
 * `Cannot read properties of undefined (reading 'children')` crash that
 * took the whole route bundle down. We now return whatever prefix we
 * managed to resolve instead.
 *
 * See issue #1.
 */
export function resolveSelectedLinks(links: Link[], normalizedPath: string): SelectedLinks {
  const result: SelectedLinks = {
    selectedLink: null,
    selectedSubLink: null,
    selectedSubSubLink: null
  };

  if (!normalizedPath) {
    return result;
  }

  result.selectedLink = _.find(links, (l) => normalizedPath == l.href) ?? null;
  if (result.selectedLink) {
    return result;
  }

  result.selectedLink =
    _.find(links, (l) => !_.isEmpty(l.children) && normalizedPath.startsWith(l.href)) ?? null;

  if (!result.selectedLink) {
    return result;
  }

  const parentHref = result.selectedLink.href;
  const subChildren = result.selectedLink.children ?? [];

  result.selectedSubLink =
    _.find(subChildren, (l) => normalizedPath == parentHref + l.href) ?? null;

  if (!result.selectedSubLink) {
    result.selectedSubLink =
      _.find(subChildren, (l) => normalizedPath.startsWith(parentHref + l.href)) ?? null;
  }

  if (!result.selectedSubLink) {
    return result;
  }

  const subSubChildren = result.selectedSubLink.children ?? [];
  if (_.isEmpty(subSubChildren)) {
    return result;
  }

  const subHref = result.selectedSubLink.href;
  result.selectedSubSubLink =
    _.find(subSubChildren, (l) => normalizedPath.startsWith(parentHref + subHref + l.href)) ?? null;

  return result;
}
