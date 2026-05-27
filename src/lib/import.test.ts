import { describe, expect, test } from "bun:test";

import helpers from "./template_helpers";

describe("template helpers", () => {
  test("acronym", () => {
    expect(helpers.acronym("Foo Bar baz")).toBe("FBB");
    expect(helpers.acronym("foo   the bar")).toBe("FB");
    expect(helpers.acronym("Motital S & P 500")).toBe("MSP");
    expect(helpers.acronym("Axis Liquid Growth Direct Plan")).toBe("AL");
  });
});
