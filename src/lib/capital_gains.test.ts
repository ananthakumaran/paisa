import { describe, expect, test } from "bun:test";
import { resolveSelectedLinks, type Link } from "./navbar_links";

// The Capital Gains and Schedule AL pages live at /more/tax/capital_gains
// and /more/tax/schedule_al respectively. When the Tax submenu has NOT
// been added to the navbar (for example, when default_currency is not
// "INR"), the navbar resolver used to crash with
//   `Cannot read properties of undefined (reading 'children')`
// on every navigation to those pages, taking the whole route bundle
// down with it. The tests below pin that behaviour: the resolver must
// degrade gracefully, returning whatever prefix it could match rather
// than throwing.
//
// See issue #1.

const linksWithoutTax: Link[] = [
  { label: "Income", href: "/income" },
  {
    label: "More",
    href: "/more",
    children: [
      { label: "Configuration", href: "/config" },
      { label: "About", href: "/about" }
    ]
  }
];

const linksWithTax: Link[] = [
  { label: "Income", href: "/income" },
  {
    label: "More",
    href: "/more",
    children: [
      { label: "Configuration", href: "/config" },
      {
        label: "Tax",
        href: "/tax",
        children: [
          { label: "Harvest", href: "/harvest" },
          { label: "Capital Gains", href: "/capital_gains" },
          { label: "Schedule AL", href: "/schedule_al" }
        ]
      },
      { label: "About", href: "/about" }
    ]
  }
];

describe("navbar link resolver / capital_gains crash", () => {
  test("does not throw on /more/tax/capital_gains when Tax submenu is missing", () => {
    expect(() => resolveSelectedLinks(linksWithoutTax, "/more/tax/capital_gains")).not.toThrow();
  });

  test("does not throw on /more/tax/schedule_al when Tax submenu is missing", () => {
    expect(() => resolveSelectedLinks(linksWithoutTax, "/more/tax/schedule_al")).not.toThrow();
  });

  test("does not throw on /more/tax/harvest when Tax submenu is missing", () => {
    expect(() => resolveSelectedLinks(linksWithoutTax, "/more/tax/harvest")).not.toThrow();
  });

  test("falls back to the closest matching parent link", () => {
    const result = resolveSelectedLinks(linksWithoutTax, "/more/tax/capital_gains");
    expect(result.selectedLink?.href).toBe("/more");
    expect(result.selectedSubLink).toBeNull();
    expect(result.selectedSubSubLink).toBeNull();
  });

  test("still resolves the full path when Tax submenu IS registered", () => {
    const result = resolveSelectedLinks(linksWithTax, "/more/tax/capital_gains");
    expect(result.selectedLink?.href).toBe("/more");
    expect(result.selectedSubLink?.href).toBe("/tax");
    expect(result.selectedSubSubLink?.href).toBe("/capital_gains");
  });

  test("resolves a direct top-level link", () => {
    const result = resolveSelectedLinks(linksWithTax, "/income");
    expect(result.selectedLink?.href).toBe("/income");
    expect(result.selectedSubLink).toBeNull();
    expect(result.selectedSubSubLink).toBeNull();
  });

  test("handles an empty path without throwing", () => {
    expect(() => resolveSelectedLinks(linksWithTax, "")).not.toThrow();
  });
});
